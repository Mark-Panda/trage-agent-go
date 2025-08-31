package llm

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Response   *LLMMessage `json:"response"`
	Timestamp  time.Time   `json:"timestamp"`
	Expiration time.Time   `json:"expiration"`
	HitCount   int         `json:"hit_count"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MaxSize         int           `json:"max_size"`         // 最大缓存条目数
	TTL             time.Duration `json:"ttl"`              // 生存时间
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理间隔
	EnableStats     bool          `json:"enable_stats"`     // 是否启用统计
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:         1000,
		TTL:             1 * time.Hour,
		CleanupInterval: 10 * time.Minute,
		EnableStats:     true,
	}
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	MaxSize   int     `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
}

// LLMCache LLM缓存接口
type LLMCache interface {
	// Get 获取缓存
	Get(key string) (*LLMMessage, bool)
	// Set 设置缓存
	Set(key string, response *LLMMessage) error
	// Delete 删除缓存
	Delete(key string)
	// Clear 清空缓存
	Clear()
	// GetStats 获取统计信息
	GetStats() *CacheStats
	// Cleanup 清理过期条目
	Cleanup()
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	config *CacheConfig
	cache  map[string]*CacheEntry
	stats  *CacheStats
	mutex  sync.RWMutex
	stop   chan struct{}
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(config *CacheConfig) *MemoryCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cache := &MemoryCache{
		config: config,
		cache:  make(map[string]*CacheEntry),
		stats: &CacheStats{
			MaxSize: config.MaxSize,
		},
		stop: make(chan struct{}),
	}

	// 启动清理协程
	if config.CleanupInterval > 0 {
		go cache.cleanupWorker()
	}

	return cache
}

// Get 获取缓存
func (mc *MemoryCache) Get(key string) (*LLMMessage, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	entry, exists := mc.cache[key]
	if !exists {
		mc.stats.Misses++
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.Expiration) {
		mc.mutex.RUnlock()
		mc.mutex.Lock()
		delete(mc.cache, key)
		mc.stats.Size--
		mc.stats.Evictions++
		mc.mutex.Unlock()
		mc.mutex.RLock()
		mc.stats.Misses++
		return nil, false
	}

	// 更新命中统计
	entry.HitCount++
	mc.stats.Hits++

	// 计算命中率
	if mc.stats.Hits+mc.stats.Misses > 0 {
		mc.stats.HitRate = float64(mc.stats.Hits) / float64(mc.stats.Hits+mc.stats.Misses)
	}

	return entry.Response, true
}

// Set 设置缓存
func (mc *MemoryCache) Set(key string, response *LLMMessage) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// 检查缓存大小限制
	if len(mc.cache) >= mc.config.MaxSize {
		// 执行LRU清理
		mc.evictLRU()
	}

	entry := &CacheEntry{
		Response:   response,
		Timestamp:  time.Now(),
		Expiration: time.Now().Add(mc.config.TTL),
		HitCount:   0,
	}

	mc.cache[key] = entry
	mc.stats.Size = len(mc.cache)

	return nil
}

// Delete 删除缓存
func (mc *MemoryCache) Delete(key string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.cache[key]; exists {
		delete(mc.cache, key)
		mc.stats.Size = len(mc.cache)
	}
}

// Clear 清空缓存
func (mc *MemoryCache) Clear() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.cache = make(map[string]*CacheEntry)
	mc.stats.Size = 0
	mc.stats.Hits = 0
	mc.stats.Misses = 0
	mc.stats.Evictions = 0
	mc.stats.HitRate = 0
}

// GetStats 获取统计信息
func (mc *MemoryCache) GetStats() *CacheStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	// 复制统计信息
	stats := *mc.stats
	return &stats
}

// Cleanup 清理过期条目
func (mc *MemoryCache) Cleanup() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	now := time.Now()
	for key, entry := range mc.cache {
		if now.After(entry.Expiration) {
			delete(mc.cache, key)
			mc.stats.Evictions++
		}
	}
	mc.stats.Size = len(mc.cache)
}

// evictLRU 执行LRU清理
func (mc *MemoryCache) evictLRU() {
	if len(mc.cache) == 0 {
		return
	}

	// 找到最少使用的条目
	var oldestKey string
	var oldestEntry *CacheEntry

	for key, entry := range mc.cache {
		if oldestEntry == nil || entry.Timestamp.Before(oldestEntry.Timestamp) {
			oldestKey = key
			oldestEntry = entry
		}
	}

	// 删除最旧的条目
	if oldestKey != "" {
		delete(mc.cache, oldestKey)
		mc.stats.Evictions++
	}
}

// cleanupWorker 清理工作协程
func (mc *MemoryCache) cleanupWorker() {
	ticker := time.NewTicker(mc.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.Cleanup()
		case <-mc.stop:
			return
		}
	}
}

// Stop 停止缓存
func (mc *MemoryCache) Stop() {
	close(mc.stop)
}

// CachedLLMClient 带缓存的LLM客户端包装器
type CachedLLMClient struct {
	client LLMClient
	cache  LLMCache
}

// NewCachedLLMClient 创建带缓存的LLM客户端
func NewCachedLLMClient(client LLMClient, cache LLMCache) *CachedLLMClient {
	return &CachedLLMClient{
		client: client,
		cache:  cache,
	}
}

// Chat 实现LLMClient接口，带缓存
func (clc *CachedLLMClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	// 生成缓存键
	cacheKey := clc.generateCacheKey(messages, tools, config)

	// 尝试从缓存获取
	if cached, exists := clc.cache.Get(cacheKey); exists {
		return cached, nil
	}

	// 缓存未命中，调用实际客户端
	response, err := clc.client.Chat(messages, tools, config)
	if err != nil {
		return nil, err
	}

	// 将响应存入缓存
	clc.cache.Set(cacheKey, response)

	return response, nil
}

// generateCacheKey 生成缓存键
func (clc *CachedLLMClient) generateCacheKey(messages []LLMMessage, tools []Tool, config ModelConfig) string {
	// 创建包含所有相关信息的结构
	cacheData := struct {
		Messages []LLMMessage `json:"messages"`
		Tools    []Tool       `json:"tools"`
		Model    string       `json:"model"`
		Provider string       `json:"provider"`
	}{
		Messages: messages,
		Tools:    tools,
		Model:    config.GetModel(),
		Provider: clc.client.GetProvider(),
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		// 如果序列化失败，使用简单的字符串拼接
		return fmt.Sprintf("%s-%s-%s", config.GetModel(), clc.client.GetProvider(), time.Now().Format("2006-01-02"))
	}

	// 计算SHA256哈希作为缓存键
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

// SetTrajectoryRecorder 设置轨迹记录器
func (clc *CachedLLMClient) SetTrajectoryRecorder(recorder TrajectoryRecorder) {
	clc.client.SetTrajectoryRecorder(recorder)
}

// GetProvider 获取提供商名称
func (clc *CachedLLMClient) GetProvider() string {
	return clc.client.GetProvider()
}

// SupportsToolCalling 检查是否支持工具调用
func (clc *CachedLLMClient) SupportsToolCalling() bool {
	return clc.client.SupportsToolCalling()
}

// GetCache 获取缓存实例
func (clc *CachedLLMClient) GetCache() LLMCache {
	return clc.cache
}
