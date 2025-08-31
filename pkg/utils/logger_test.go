package utils

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}
	
	structuredLogger, ok := logger.(*StructuredLogger)
	if !ok {
		t.Fatal("Expected StructuredLogger type")
	}
	
	if structuredLogger.level != LogLevelInfo {
		t.Errorf("Expected level LogLevelInfo, got %v", structuredLogger.level)
	}
}

func TestLogLevel_String(t *testing.T) {
	testCases := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
	}
	
	for _, tc := range testCases {
		result := tc.level.String()
		if result != tc.expected {
			t.Errorf("Expected %s for level %v, got %s", tc.expected, tc.level, result)
		}
	}
}

func TestField_Creation(t *testing.T) {
	field := F("test_key", "test_value")
	
	if field.Key != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", field.Key)
	}
	
	if field.Value != "test_value" {
		t.Errorf("Expected value 'test_value', got '%v'", field.Value)
	}
}

func TestStructuredLogger_WithFields(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	
	fields := []Field{
		F("key1", "value1"),
		F("key2", "value2"),
	}
	
	loggerWithFields := logger.WithFields(fields...)
	
	structuredLogger, ok := loggerWithFields.(*StructuredLogger)
	if !ok {
		t.Fatal("Expected StructuredLogger type")
	}
	
	if len(structuredLogger.fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(structuredLogger.fields))
	}
	
	if structuredLogger.fields[0].Key != "key1" {
		t.Errorf("Expected first field key 'key1', got '%s'", structuredLogger.fields[0].Key)
	}
	
	if structuredLogger.fields[1].Key != "key2" {
		t.Errorf("Expected second field key 'key2', got '%s'", structuredLogger.fields[1].Key)
	}
}

func TestStructuredLogger_SetLevel(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	
	logger.SetLevel(LogLevelDebug)
	
	structuredLogger, ok := logger.(*StructuredLogger)
	if !ok {
		t.Fatal("Expected StructuredLogger type")
	}
	
	if structuredLogger.level != LogLevelDebug {
		t.Errorf("Expected level LogLevelDebug, got %v", structuredLogger.level)
	}
}

func TestGlobalLogger(t *testing.T) {
	// 测试全局日志记录器
	originalLogger := GlobalLogger
	
	// 创建新的日志记录器
	newLogger := NewLogger(LogLevelDebug)
	SetGlobalLogger(newLogger)
	
	if GlobalLogger != newLogger {
		t.Error("Expected global logger to be set")
	}
	
	// 测试全局日志函数
	Debug("test debug message")
	Info("test info message")
	Warn("test warning message")
	Error("test error message")
	
	// 恢复原始日志记录器
	SetGlobalLogger(originalLogger)
}

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	
	if logger == nil {
		t.Fatal("Expected default logger to be created")
	}
	
	structuredLogger, ok := logger.(*StructuredLogger)
	if !ok {
		t.Fatal("Expected StructuredLogger type")
	}
	
	// 默认应该是INFO级别
	if structuredLogger.level != LogLevelInfo {
		t.Errorf("Expected default level LogLevelInfo, got %v", structuredLogger.level)
	}
}
