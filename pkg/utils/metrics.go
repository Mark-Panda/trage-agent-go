package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric 指标接口
type Metric interface {
	GetName() string
	GetType() MetricType
	GetValue() interface{}
	GetDescription() string
}

// Counter 计数器指标
type Counter struct {
	name        string
	description string
	value       int64
}

// NewCounter 创建新的计数器
func NewCounter(name, description string) *Counter {
	return &Counter{
		name:        name,
		description: description,
		value:       0,
	}
}

// GetName 获取指标名称
func (c *Counter) GetName() string {
	return c.name
}

// GetType 获取指标类型
func (c *Counter) GetType() MetricType {
	return MetricTypeCounter
}

// GetValue 获取指标值
func (c *Counter) GetValue() interface{} {
	return atomic.LoadInt64(&c.value)
}

// GetDescription 获取指标描述
func (c *Counter) GetDescription() string {
	return c.description
}

// Increment 增加计数器
func (c *Counter) Increment() {
	atomic.AddInt64(&c.value, 1)
}

// Add 增加指定值
func (c *Counter) Add(value int64) {
	atomic.AddInt64(&c.value, value)
}

// Reset 重置计数器
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Gauge 仪表指标
type Gauge struct {
	name        string
	description string
	value       int64
}

// NewGauge 创建新的仪表
func NewGauge(name, description string) *Gauge {
	return &Gauge{
		name:        name,
		description: description,
		value:       0,
	}
}

// GetName 获取指标名称
func (c *Gauge) GetName() string {
	return c.name
}

// GetType 获取指标类型
func (c *Gauge) GetType() MetricType {
	return MetricTypeGauge
}

// GetValue 获取指标值
func (c *Gauge) GetValue() interface{} {
	return atomic.LoadInt64(&c.value)
}

// GetDescription 获取指标描述
func (c *Gauge) GetDescription() string {
	return c.description
}

// Set 设置仪表值
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Add 增加仪表值
func (g *Gauge) Add(value int64) {
	atomic.AddInt64(&g.value, value)
}

// Subtract 减少仪表值
func (g *Gauge) Subtract(value int64) {
	atomic.AddInt64(&g.value, -value)
}

// Histogram 直方图指标
type Histogram struct {
	name        string
	description string
	buckets     []float64
	counts      []int64
	total       int64
	sum         float64
	min         float64
	max         float64
	mutex       sync.RWMutex
}

// NewHistogram 创建新的直方图
func NewHistogram(name, description string, buckets []float64) *Histogram {
	if len(buckets) == 0 {
		buckets = []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0}
	}
	
	h := &Histogram{
		name:        name,
		description: description,
		buckets:     buckets,
		counts:      make([]int64, len(buckets)),
		min:         float64(^uint(0) >> 1), // 最大float64值
		max:         -float64(^uint(0) >> 1), // 最小float64值
	}
	
	return h
}

// GetName 获取指标名称
func (h *Histogram) GetName() string {
	return h.name
}

// GetType 获取指标类型
func (h *Histogram) GetType() MetricType {
	return MetricTypeHistogram
}

// GetValue 获取指标值
func (h *Histogram) GetValue() interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return map[string]interface{}{
		"buckets": h.buckets,
		"counts":  h.counts,
		"total":   h.total,
		"sum":     h.sum,
		"min":     h.min,
		"max":     h.max,
		"mean":    h.getMean(),
	}
}

// GetDescription 获取指标描述
func (h *Histogram) GetDescription() string {
	return h.description
}

// Observe 观察一个值
func (h *Histogram) Observe(value float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.total++
	h.sum += value
	
	if value < h.min {
		h.min = value
	}
	if value > h.max {
		h.max = value
	}
	
	// 更新桶计数
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			break
		}
	}
}

// getMean 计算平均值
func (h *Histogram) getMean() float64 {
	if h.total == 0 {
		return 0
	}
	return h.sum / float64(h.total)
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	metrics map[string]Metric
	mutex   sync.RWMutex
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]Metric),
	}
}

// RegisterMetric 注册指标
func (mc *MetricsCollector) RegisterMetric(metric Metric) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if _, exists := mc.metrics[metric.GetName()]; exists {
		return fmt.Errorf("metric %s already exists", metric.GetName())
	}
	
	mc.metrics[metric.GetName()] = metric
	return nil
}

// GetMetric 获取指标
func (mc *MetricsCollector) GetMetric(name string) (Metric, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	metric, exists := mc.metrics[name]
	return metric, exists
}

// GetAllMetrics 获取所有指标
func (mc *MetricsCollector) GetAllMetrics() map[string]Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	result := make(map[string]Metric)
	for name, metric := range mc.metrics {
		result[name] = metric
	}
	return result
}

// ResetAllMetrics 重置所有指标
func (mc *MetricsCollector) ResetAllMetrics() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	for _, metric := range mc.metrics {
		switch m := metric.(type) {
		case *Counter:
			m.Reset()
		case *Gauge:
			m.Set(0)
		case *Histogram:
			// 直方图重置需要重新创建
		}
	}
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	collector *MetricsCollector
	logger    Logger
}

// NewPerformanceMonitor 创建新的性能监控器
func NewPerformanceMonitor(logger Logger) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		collector: NewMetricsCollector(),
		logger:    logger,
	}
	
	// 注册默认指标
	pm.registerDefaultMetrics()
	
	return pm
}

// registerDefaultMetrics 注册默认指标
func (pm *PerformanceMonitor) registerDefaultMetrics() {
	// API调用计数
	pm.collector.RegisterMetric(NewCounter("api_calls_total", "Total number of API calls"))
	
	// API调用延迟
	pm.collector.RegisterMetric(NewHistogram("api_latency_seconds", "API call latency in seconds", nil))
	
	// 缓存命中率
	pm.collector.RegisterMetric(NewGauge("cache_hit_rate", "Cache hit rate percentage"))
	
	// 重试次数
	pm.collector.RegisterMetric(NewCounter("retry_attempts_total", "Total number of retry attempts"))
	
	// 错误计数
	pm.collector.RegisterMetric(NewCounter("errors_total", "Total number of errors"))
}

// RecordAPICall 记录API调用
func (pm *PerformanceMonitor) RecordAPICall(provider string, duration time.Duration, success bool) {
	// 增加API调用计数
	if metric, exists := pm.collector.GetMetric("api_calls_total"); exists {
		if counter, ok := metric.(*Counter); ok {
			counter.Increment()
		}
	}
	
	// 记录延迟
	if metric, exists := pm.collector.GetMetric("api_latency_seconds"); exists {
		if histogram, ok := metric.(*Histogram); ok {
			histogram.Observe(duration.Seconds())
		}
	}
	
	// 记录错误
	if !success {
		if metric, exists := pm.collector.GetMetric("errors_total"); exists {
			if counter, ok := metric.(*Counter); ok {
				counter.Increment()
			}
		}
	}
	
	// 记录日志
	pm.logger.Info("API call recorded",
		F("provider", provider),
		F("duration_ms", duration.Milliseconds()),
		F("success", success),
	)
}

// RecordCacheHit 记录缓存命中
func (pm *PerformanceMonitor) RecordCacheHit(hit bool) {
	if metric, exists := pm.collector.GetMetric("cache_hit_rate"); exists {
		if gauge, ok := metric.(*Gauge); ok {
			// 这里简化处理，实际应该计算滑动窗口的命中率
			if hit {
				gauge.Add(1)
			}
		}
	}
}

// RecordRetry 记录重试
func (pm *PerformanceMonitor) RecordRetry(attempt int) {
	if metric, exists := pm.collector.GetMetric("retry_attempts_total"); exists {
		if counter, ok := metric.(*Counter); ok {
			counter.Add(int64(attempt))
		}
	}
}

// GetMetrics 获取所有指标
func (pm *PerformanceMonitor) GetMetrics() map[string]Metric {
	return pm.collector.GetAllMetrics()
}

// ExportMetrics 导出指标为Prometheus格式
func (pm *PerformanceMonitor) ExportMetrics() string {
	metrics := pm.collector.GetAllMetrics()
	var result string
	
	for name, metric := range metrics {
		result += fmt.Sprintf("# HELP %s %s\n", name, metric.GetDescription())
		result += fmt.Sprintf("# TYPE %s %s\n", name, metric.GetType())
		
		switch m := metric.(type) {
		case *Counter:
			result += fmt.Sprintf("%s %v\n", name, m.GetValue())
		case *Gauge:
			result += fmt.Sprintf("%s %v\n", name, m.GetValue())
		case *Histogram:
			value := m.GetValue().(map[string]interface{})
			buckets := value["buckets"].([]float64)
			counts := value["counts"].([]int64)
			
			for i, bucket := range buckets {
				result += fmt.Sprintf("%s_bucket{le=\"%g\"} %d\n", name, bucket, counts[i])
			}
			result += fmt.Sprintf("%s_sum %g\n", name, value["sum"])
			result += fmt.Sprintf("%s_count %d\n", name, value["total"])
		}
		result += "\n"
	}
	
	return result
}
