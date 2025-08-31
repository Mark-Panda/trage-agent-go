package utils

import (
	"testing"
	"time"
)

func TestNewCounter(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter description")
	
	if counter == nil {
		t.Fatal("Expected counter to be created")
	}
	
	if counter.GetName() != "test_counter" {
		t.Errorf("Expected name 'test_counter', got '%s'", counter.GetName())
	}
	
	if counter.GetDescription() != "Test counter description" {
		t.Errorf("Expected description 'Test counter description', got '%s'", counter.GetDescription())
	}
	
	if counter.GetType() != MetricTypeCounter {
		t.Errorf("Expected type MetricTypeCounter, got %v", counter.GetType())
	}
	
	if counter.GetValue() != int64(0) {
		t.Errorf("Expected initial value 0, got %v", counter.GetValue())
	}
}

func TestCounter_Increment(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter")
	
	// 初始值应该是0
	if counter.GetValue() != int64(0) {
		t.Errorf("Expected initial value 0, got %v", counter.GetValue())
	}
	
	// 增加1
	counter.Increment()
	if counter.GetValue() != int64(1) {
		t.Errorf("Expected value 1 after increment, got %v", counter.GetValue())
	}
	
	// 增加指定值
	counter.Add(5)
	if counter.GetValue() != int64(6) {
		t.Errorf("Expected value 6 after adding 5, got %v", counter.GetValue())
	}
	
	// 重置
	counter.Reset()
	if counter.GetValue() != int64(0) {
		t.Errorf("Expected value 0 after reset, got %v", counter.GetValue())
	}
}

func TestNewGauge(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge description")
	
	if gauge == nil {
		t.Fatal("Expected gauge to be created")
	}
	
	if gauge.GetName() != "test_gauge" {
		t.Errorf("Expected name 'test_gauge', got '%s'", gauge.GetName())
	}
	
	if gauge.GetType() != MetricTypeGauge {
		t.Errorf("Expected type MetricTypeGauge, got %v", gauge.GetType())
	}
}

func TestGauge_Operations(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge")
	
	// 设置值
	gauge.Set(10)
	if gauge.GetValue() != int64(10) {
		t.Errorf("Expected value 10 after set, got %v", gauge.GetValue())
	}
	
	// 增加值
	gauge.Add(5)
	if gauge.GetValue() != int64(15) {
		t.Errorf("Expected value 15 after adding 5, got %v", gauge.GetValue())
	}
	
	// 减少值
	gauge.Subtract(3)
	if gauge.GetValue() != int64(12) {
		t.Errorf("Expected value 12 after subtracting 3, got %v", gauge.GetValue())
	}
}

func TestNewHistogram(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0}
	histogram := NewHistogram("test_histogram", "Test histogram", buckets)
	
	if histogram == nil {
		t.Fatal("Expected histogram to be created")
	}
	
	if histogram.GetName() != "test_histogram" {
		t.Errorf("Expected name 'test_histogram', got '%s'", histogram.GetName())
	}
	
	if histogram.GetType() != MetricTypeHistogram {
		t.Errorf("Expected type MetricTypeHistogram, got %v", histogram.GetType())
	}
}

func TestHistogram_Observe(t *testing.T) {
	histogram := NewHistogram("test_histogram", "Test histogram", nil)
	
	// 观察一些值
	histogram.Observe(0.1)
	histogram.Observe(0.5)
	histogram.Observe(1.0)
	histogram.Observe(2.0)
	
	value := histogram.GetValue().(map[string]interface{})
	
	if value["total"] != int64(4) {
		t.Errorf("Expected total 4, got %v", value["total"])
	}
	
	if value["sum"] != 3.6 {
		t.Errorf("Expected sum 3.6, got %v", value["sum"])
	}
	
	if value["min"] != 0.1 {
		t.Errorf("Expected min 0.1, got %v", value["min"])
	}
	
	if value["max"] != 2.0 {
		t.Errorf("Expected max 2.0, got %v", value["max"])
	}
	
	mean := value["mean"].(float64)
	if mean != 0.9 {
		t.Errorf("Expected mean 0.9, got %v", mean)
	}
}

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	
	if collector == nil {
		t.Fatal("Expected collector to be created")
	}
	
	metrics := collector.GetAllMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected empty metrics map, got %d metrics", len(metrics))
	}
}

func TestMetricsCollector_RegisterMetric(t *testing.T) {
	collector := NewMetricsCollector()
	
	counter := NewCounter("test_counter", "Test counter")
	
	// 注册指标
	err := collector.RegisterMetric(counter)
	if err != nil {
		t.Errorf("Expected no error registering metric, got: %v", err)
	}
	
	// 验证指标已注册
	metric, exists := collector.GetMetric("test_counter")
	if !exists {
		t.Error("Expected metric to exist after registration")
	}
	
	if metric != counter {
		t.Error("Expected returned metric to be the same as registered metric")
	}
	
	// 尝试重复注册应该失败
	err = collector.RegisterMetric(counter)
	if err == nil {
		t.Error("Expected error when registering duplicate metric")
	}
}

func TestNewPerformanceMonitor(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	monitor := NewPerformanceMonitor(logger)
	
	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}
	
	metrics := monitor.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected default metrics to be registered")
	}
}

func TestPerformanceMonitor_RecordAPICall(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	monitor := NewPerformanceMonitor(logger)
	
	// 记录API调用
	monitor.RecordAPICall("openai", 100*time.Millisecond, true)
	monitor.RecordAPICall("doubao", 200*time.Millisecond, false)
	
	// 验证指标
	metrics := monitor.GetMetrics()
	
	// 检查API调用计数
	if apiCallsMetric, exists := metrics["api_calls_total"]; exists {
		if counter, ok := apiCallsMetric.(*Counter); ok {
			if counter.GetValue() != int64(2) {
				t.Errorf("Expected 2 API calls, got %v", counter.GetValue())
			}
		}
	}
	
	// 检查错误计数
	if errorsMetric, exists := metrics["errors_total"]; exists {
		if counter, ok := errorsMetric.(*Counter); ok {
			if counter.GetValue() != int64(1) {
				t.Errorf("Expected 1 error, got %v", counter.GetValue())
			}
		}
	}
}

func TestPerformanceMonitor_ExportMetrics(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	monitor := NewPerformanceMonitor(logger)
	
	// 记录一些指标
	monitor.RecordAPICall("test", 100*time.Millisecond, true)
	
	// 导出指标
	exported := monitor.ExportMetrics()
	
	if exported == "" {
		t.Error("Expected non-empty exported metrics")
	}
	
	// 验证包含必要的Prometheus格式
	if len(exported) < 100 {
		t.Error("Expected substantial metrics export")
	}
}
