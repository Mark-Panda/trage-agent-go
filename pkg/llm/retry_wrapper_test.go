package llm

import (
	"errors"
	"testing"
	"time"
)

// MockLLMClient 模拟LLM客户端，用于测试重试机制
type MockLLMClient struct {
	*BaseLLMClient
	failCount    int
	shouldFail   bool
	lastAttempts int
}

// NewMockLLMClient 创建模拟LLM客户端
func NewMockLLMClient(failCount int) *MockLLMClient {
	return &MockLLMClient{
		BaseLLMClient: NewBaseLLMClient("test_key", "https://test.com", "v1", "mock"),
		failCount:     failCount,
		shouldFail:    true,
		lastAttempts:  0,
	}
}

// Chat 实现LLMClient接口
func (m *MockLLMClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	m.lastAttempts++

	if m.shouldFail && m.lastAttempts <= m.failCount {
		return nil, errors.New("mock error for testing retry")
	}

	// 成功返回
	return &LLMMessage{
		Role:    "assistant",
		Content: "Success after retries!",
	}, nil
}

func TestNewRetryableLLMClient(t *testing.T) {
	mockClient := NewMockLLMClient(0)
	retryConfig := DefaultRetryConfig()

	retryableClient := NewRetryableLLMClient(mockClient, retryConfig)

	if retryableClient == nil {
		t.Fatal("Expected retryable client to be created")
	}

	if retryableClient.client != mockClient {
		t.Error("Expected client to be set correctly")
	}

	if retryableClient.retryConfig != retryConfig {
		t.Error("Expected retry config to be set correctly")
	}
}

func TestRetryableLLMClient_SuccessAfterRetries(t *testing.T) {
	mockClient := NewMockLLMClient(2) // 前2次失败，第3次成功
	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BaseDelay:   10 * time.Millisecond, // 使用较短的延迟进行测试
		MaxDelay:    100 * time.Millisecond,
		BackoffRate: 2.0,
	}

	retryableClient := NewRetryableLLMClient(mockClient, retryConfig)

	messages := []LLMMessage{{Role: "user", Content: "test"}}
	tools := []Tool{}
	config := &MockModelConfig{}

	start := time.Now()
	response, err := retryableClient.Chat(messages, tools, config)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Content != "Success after retries!" {
		t.Errorf("Expected content 'Success after retries!', got '%s'", response.Content)
	}

	// 验证重试次数
	if mockClient.lastAttempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", mockClient.lastAttempts)
	}

	// 验证总时间应该大于延迟时间
	expectedMinDelay := 10*time.Millisecond + 20*time.Millisecond // 第1次和第2次延迟
	if duration < expectedMinDelay {
		t.Errorf("Expected duration >= %v, got %v", expectedMinDelay, duration)
	}
}

func TestRetryableLLMClient_MaxRetriesExceeded(t *testing.T) {
	mockClient := NewMockLLMClient(5) // 总是失败
	retryConfig := &RetryConfig{
		MaxRetries:  2,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		BackoffRate: 2.0,
	}

	retryableClient := NewRetryableLLMClient(mockClient, retryConfig)

	messages := []LLMMessage{{Role: "user", Content: "test"}}
	tools := []Tool{}
	config := &MockModelConfig{}

	response, err := retryableClient.Chat(messages, tools, config)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response, got response")
	}

	// 验证错误消息
	expectedError := "max retries exceeded"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// 验证尝试次数
	if mockClient.lastAttempts != 3 { // 初始尝试 + 2次重试
		t.Errorf("Expected 3 attempts, got %d", mockClient.lastAttempts)
	}
}

func TestRetryableLLMClient_NonRetryableError(t *testing.T) {
	// 创建一个自定义的重试配置
	customRetryConfig := &RetryConfig{
		MaxRetries:  3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		BackoffRate: 2.0,
	}

	// 创建一个特殊的模拟客户端，它返回一个特殊的错误
	specialMockClient := &SpecialMockLLMClient{
		BaseLLMClient: NewBaseLLMClient("test_key", "https://test.com", "v1", "mock"),
	}

	retryableClient := NewRetryableLLMClient(specialMockClient, customRetryConfig)

	messages := []LLMMessage{{Role: "user", Content: "test"}}
	tools := []Tool{}
	config := &MockModelConfig{}

	response, err := retryableClient.Chat(messages, tools, config)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response, got response")
	}

	// 验证只尝试了一次
	if specialMockClient.attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", specialMockClient.attempts)
	}
}

// SpecialMockLLMClient 特殊的模拟客户端，用于测试不可重试错误
type SpecialMockLLMClient struct {
	*BaseLLMClient
	attempts int
}

func (sm *SpecialMockLLMClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	sm.attempts++
	// 返回一个特殊的错误，这个错误在types.go中被标记为不可重试
	return nil, &Error{
		Type:    "invalid_request",
		Code:    "invalid_api_key",
		Message: "Invalid API key",
	}
}

// MockModelConfig 模拟模型配置
type MockModelConfig struct{}

func (m *MockModelConfig) GetModel() string             { return "test-model" }
func (m *MockModelConfig) GetMaxTokens() int            { return 1000 }
func (m *MockModelConfig) GetTemperature() float64      { return 0.5 }
func (m *MockModelConfig) GetTopP() float64             { return 1.0 }
func (m *MockModelConfig) GetTopK() int                 { return 1 }
func (m *MockModelConfig) GetParallelToolCalls() bool   { return true }
func (m *MockModelConfig) GetMaxRetries() int           { return 3 }
func (m *MockModelConfig) GetSupportsToolCalling() bool { return true }
func (m *MockModelConfig) GetAPIKey() string            { return "test-key" }
func (m *MockModelConfig) GetBaseURL() string           { return "https://test.com" }
func (m *MockModelConfig) GetAPIVersion() string        { return "v1" }

func TestEnhancedRetryConfig(t *testing.T) {
	enhancedConfig := NewEnhancedRetryConfig()

	// 测试默认值
	if enhancedConfig.RetryConfig == nil {
		t.Fatal("Expected RetryConfig to be initialized")
	}

	if enhancedConfig.ErrorRetryMap == nil {
		t.Fatal("Expected ErrorRetryMap to be initialized")
	}

	// 测试添加可重试错误
	enhancedConfig.AddRetryableError("network_error")
	if !enhancedConfig.ErrorRetryMap["network_error"] {
		t.Error("Expected network_error to be added to retryable errors")
	}

	// 测试自定义重试条件
	customCondition := func(err error) bool {
		return err.Error() == "custom_error"
	}
	enhancedConfig.SetRetryCondition(customCondition)

	// 测试回调设置
	enhancedConfig.SetBeforeRetry(func(attempt int, err error) {
		// 回调被设置
	})

	enhancedConfig.SetAfterRetry(func(attempt int, err error, delay time.Duration) {
		// 回调被设置
	})

	// 验证回调被设置
	if enhancedConfig.BeforeRetry == nil {
		t.Error("Expected BeforeRetry to be set")
	}

	if enhancedConfig.AfterRetry == nil {
		t.Error("Expected AfterRetry to be set")
	}
}
