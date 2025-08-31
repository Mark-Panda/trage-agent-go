package llm

import (
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         100,
		TTL:             30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
		EnableStats:     true,
	}

	cache := NewMemoryCache(config)

	if cache == nil {
		t.Fatal("Expected cache to be created")
	}

	if cache.config != config {
		t.Error("Expected config to be set correctly")
	}

	if len(cache.cache) != 0 {
		t.Error("Expected cache to be empty initially")
	}

	// 清理
	cache.Stop()
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(DefaultCacheConfig())
	defer cache.Stop()

	// 创建测试消息
	message := &LLMMessage{
		Role:    "assistant",
		Content: "Hello, world!",
	}

	// 设置缓存
	err := cache.Set("test_key", message)
	if err != nil {
		t.Errorf("Expected no error setting cache, got: %v", err)
	}

	// 获取缓存
	cached, exists := cache.Get("test_key")
	if !exists {
		t.Fatal("Expected cached message to exist")
	}

	if cached.Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", cached.Content)
	}

	// 验证统计
	stats := cache.GetStats()
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Expected 0 misses, got %d", stats.Misses)
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         100,
		TTL:             10 * time.Millisecond, // 很短的TTL用于测试
		CleanupInterval: 5 * time.Millisecond,
		EnableStats:     true,
	}

	cache := NewMemoryCache(config)
	defer cache.Stop()

	message := &LLMMessage{
		Role:    "assistant",
		Content: "Expiring message",
	}

	// 设置缓存
	cache.Set("expire_key", message)

	// 立即获取应该存在
	if _, exists := cache.Get("expire_key"); !exists {
		t.Error("Expected message to exist immediately after setting")
	}

	// 等待过期
	time.Sleep(20 * time.Millisecond)

	// 现在应该不存在
	if _, exists := cache.Get("expire_key"); exists {
		t.Error("Expected message to be expired")
	}

	// 验证统计
	stats := cache.GetStats()
	if stats.Evictions == 0 {
		t.Error("Expected evictions to occur")
	}
}

func TestMemoryCache_MaxSize(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         2,
		TTL:             1 * time.Hour,
		CleanupInterval: 0, // 禁用自动清理
		EnableStats:     true,
	}

	cache := NewMemoryCache(config)
	defer cache.Stop()

	// 添加3个条目，应该触发LRU清理
	cache.Set("key1", &LLMMessage{Content: "Message 1"})
	cache.Set("key2", &LLMMessage{Content: "Message 2"})
	cache.Set("key3", &LLMMessage{Content: "Message 3"})

	// 验证大小限制
	stats := cache.GetStats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}

	// 验证LRU清理
	if stats.Evictions == 0 {
		t.Error("Expected LRU eviction to occur")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(DefaultCacheConfig())
	defer cache.Stop()

	message := &LLMMessage{Content: "To be deleted"}
	cache.Set("delete_key", message)

	// 验证存在
	if _, exists := cache.Get("delete_key"); !exists {
		t.Error("Expected message to exist before deletion")
	}

	// 删除
	cache.Delete("delete_key")

	// 验证不存在
	if _, exists := cache.Get("delete_key"); exists {
		t.Error("Expected message to not exist after deletion")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(DefaultCacheConfig())
	defer cache.Stop()

	// 添加几个条目
	cache.Set("key1", &LLMMessage{Content: "Message 1"})
	cache.Set("key2", &LLMMessage{Content: "Message 2"})

	// 验证存在
	if _, exists := cache.Get("key1"); !exists {
		t.Error("Expected key1 to exist")
	}

	// 清空缓存
	cache.Clear()

	// 验证不存在
	if _, exists := cache.Get("key1"); exists {
		t.Error("Expected key1 to not exist after clear")
	}

	// 验证统计重置
	stats := cache.GetStats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", stats.Size)
	}
}

func TestCachedLLMClient(t *testing.T) {
	// 创建模拟客户端
	mockClient := &MockLLMClient{
		BaseLLMClient: NewBaseLLMClient("test_key", "https://test.com", "v1", "mock"),
		shouldFail:    false,
	}

	// 创建缓存
	cache := NewMemoryCache(DefaultCacheConfig())
	defer cache.Stop()

	// 创建带缓存的客户端
	cachedClient := NewCachedLLMClient(mockClient, cache)

	messages := []LLMMessage{{Role: "user", Content: "test"}}
	tools := []Tool{}
	config := &MockModelConfig{}

	// 第一次调用，应该缓存未命中
	response1, err := cachedClient.Chat(messages, tools, config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 第二次调用，应该缓存命中
	response2, err := cachedClient.Chat(messages, tools, config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 验证响应相同
	if response1.Content != response2.Content {
		t.Error("Expected cached responses to be identical")
	}

	// 验证缓存统计
	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	mockClient := &MockLLMClient{
		BaseLLMClient: NewBaseLLMClient("test_key", "https://test.com", "v1", "mock"),
	}

	cache := NewMemoryCache(DefaultCacheConfig())
	defer cache.Stop()

	cachedClient := NewCachedLLMClient(mockClient, cache)

	messages1 := []LLMMessage{{Role: "user", Content: "Hello"}}
	messages2 := []LLMMessage{{Role: "user", Content: "World"}}
	tools := []Tool{}
	config := &MockModelConfig{}

	// 生成不同的缓存键
	key1 := cachedClient.generateCacheKey(messages1, tools, config)
	key2 := cachedClient.generateCacheKey(messages2, tools, config)

	if key1 == key2 {
		t.Error("Expected different cache keys for different messages")
	}

	// 相同消息应该生成相同键
	key3 := cachedClient.generateCacheKey(messages1, tools, config)
	if key1 != key3 {
		t.Error("Expected same cache key for identical messages")
	}
}
