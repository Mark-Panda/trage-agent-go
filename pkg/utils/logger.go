package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithFields(fields ...Field) Logger
	SetLevel(level LogLevel)
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// F 创建日志字段的便捷函数
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// StructuredLogger 结构化日志记录器
type StructuredLogger struct {
	level  LogLevel
	fields []Field
}

// NewLogger 创建新的日志记录器
func NewLogger(level LogLevel) Logger {
	return &StructuredLogger{
		level:  level,
		fields: make([]Field, 0),
	}
}

// NewDefaultLogger 创建默认日志记录器
func NewDefaultLogger() Logger {
	level := LogLevelInfo
	if os.Getenv("LOG_LEVEL") != "" {
		switch os.Getenv("LOG_LEVEL") {
		case "DEBUG":
			level = LogLevelDebug
		case "WARN":
			level = LogLevelWarn
		case "ERROR":
			level = LogLevelError
		}
	}
	return NewLogger(level)
}

// Debug 记录调试日志
func (sl *StructuredLogger) Debug(msg string, fields ...Field) {
	if sl.level <= LogLevelDebug {
		sl.log(LogLevelDebug, msg, fields...)
	}
}

// Info 记录信息日志
func (sl *StructuredLogger) Info(msg string, fields ...Field) {
	if sl.level <= LogLevelInfo {
		sl.log(LogLevelInfo, msg, fields...)
	}
}

// Warn 记录警告日志
func (sl *StructuredLogger) Warn(msg string, fields ...Field) {
	if sl.level <= LogLevelWarn {
		sl.log(LogLevelWarn, msg, fields...)
	}
}

// Error 记录错误日志
func (sl *StructuredLogger) Error(msg string, fields ...Field) {
	if sl.level <= LogLevelError {
		sl.log(LogLevelError, msg, fields...)
	}
}

// Fatal 记录致命错误日志并退出
func (sl *StructuredLogger) Fatal(msg string, fields ...Field) {
	if sl.level <= LogLevelFatal {
		sl.log(LogLevelFatal, msg, fields...)
		os.Exit(1)
	}
}

// WithFields 添加字段到日志记录器
func (sl *StructuredLogger) WithFields(fields ...Field) Logger {
	newLogger := &StructuredLogger{
		level:  sl.level,
		fields: append(sl.fields, fields...),
	}
	return newLogger
}

// SetLevel 设置日志级别
func (sl *StructuredLogger) SetLevel(level LogLevel) {
	sl.level = level
}

// log 内部日志记录方法
func (sl *StructuredLogger) log(level LogLevel, msg string, fields ...Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// 构建日志消息
	logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), msg)
	
	// 添加字段
	if len(sl.fields) > 0 || len(fields) > 0 {
		logMsg += " | "
		allFields := append(sl.fields, fields...)
		for i, field := range allFields {
			if i > 0 {
				logMsg += ", "
			}
			logMsg += fmt.Sprintf("%s=%v", field.Key, field.Value)
		}
	}
	
	// 根据级别选择输出方法
	switch level {
	case LogLevelFatal:
		log.Fatal(logMsg)
	case LogLevelError:
		log.Printf("ERROR: %s", logMsg)
	case LogLevelWarn:
		log.Printf("WARN: %s", logMsg)
	case LogLevelInfo:
		log.Printf("INFO: %s", logMsg)
	case LogLevelDebug:
		log.Printf("DEBUG: %s", logMsg)
	}
}

// GlobalLogger 全局日志记录器
var GlobalLogger Logger = NewDefaultLogger()

// SetGlobalLogger 设置全局日志记录器
func SetGlobalLogger(logger Logger) {
	GlobalLogger = logger
}

// Debug 全局调试日志
func Debug(msg string, fields ...Field) {
	GlobalLogger.Debug(msg, fields...)
}

// Info 全局信息日志
func Info(msg string, fields ...Field) {
	GlobalLogger.Info(msg, fields...)
}

// Warn 全局警告日志
func Warn(msg string, fields ...Field) {
	GlobalLogger.Warn(msg, fields...)
}

// Error 全局错误日志
func Error(msg string, fields ...Field) {
	GlobalLogger.Error(msg, fields...)
}

// Fatal 全局致命错误日志
func Fatal(msg string, fields ...Field) {
	GlobalLogger.Fatal(msg, fields...)
}
