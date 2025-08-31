package llm

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryableLLMClient 可重试的LLM客户端包装器
type RetryableLLMClient struct {
	client      LLMClient
	retryConfig *RetryConfig
}

// NewRetryableLLMClient 创建可重试的LLM客户端
func NewRetryableLLMClient(client LLMClient, retryConfig *RetryConfig) *RetryableLLMClient {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	return &RetryableLLMClient{
		client:      client,
		retryConfig: retryConfig,
	}
}

// Chat 实现LLMClient接口，带重试机制
func (rlc *RetryableLLMClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	var lastErr error
	var response *LLMMessage

	for attempt := 0; attempt <= rlc.retryConfig.MaxRetries; attempt++ {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		// 尝试调用
		response, lastErr = rlc.client.Chat(messages, tools, config)

		// 如果没有错误，直接返回
		if lastErr == nil {
			cancel()
			return response, nil
		}

		// 检查是否是可重试的错误
		if !IsRetryableError(lastErr) {
			cancel()
			return nil, fmt.Errorf("non-retryable error: %w", lastErr)
		}

		// 如果是最后一次尝试，返回错误
		if attempt == rlc.retryConfig.MaxRetries {
			cancel()
			return nil, fmt.Errorf("max retries exceeded, last error: %w", lastErr)
		}

		// 计算延迟时间
		delay := rlc.calculateDelay(attempt)

		// 记录重试信息
		if rlc.client.GetProvider() != "" {
			fmt.Printf("Retrying %s API call in %v (attempt %d/%d): %v\n",
				rlc.client.GetProvider(), delay, attempt+1, rlc.retryConfig.MaxRetries+1, lastErr)
		}

		// 等待延迟时间
		select {
		case <-ctx.Done():
			cancel()
			return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			cancel()
		}
	}

	return nil, fmt.Errorf("unexpected retry loop exit: %w", lastErr)
}

// calculateDelay 计算重试延迟时间
func (rlc *RetryableLLMClient) calculateDelay(attempt int) time.Duration {
	// 指数退避算法
	delay := float64(rlc.retryConfig.BaseDelay) * math.Pow(rlc.retryConfig.BackoffRate, float64(attempt))

	// 添加随机抖动（±10%）
	jitter := delay * 0.1
	delay = delay + (jitter * (float64(time.Now().UnixNano()) / float64(time.Second)))

	// 确保不超过最大延迟
	if delay > float64(rlc.retryConfig.MaxDelay) {
		delay = float64(rlc.retryConfig.MaxDelay)
	}

	return time.Duration(delay)
}

// SetTrajectoryRecorder 设置轨迹记录器
func (rlc *RetryableLLMClient) SetTrajectoryRecorder(recorder TrajectoryRecorder) {
	rlc.client.SetTrajectoryRecorder(recorder)
}

// GetProvider 获取提供商名称
func (rlc *RetryableLLMClient) GetProvider() string {
	return rlc.client.GetProvider()
}

// SupportsToolCalling 检查是否支持工具调用
func (rlc *RetryableLLMClient) SupportsToolCalling() bool {
	return rlc.client.SupportsToolCalling()
}

// EnhancedRetryConfig 增强的重试配置
type EnhancedRetryConfig struct {
	*RetryConfig
	// 特定错误类型的重试策略
	ErrorRetryMap map[string]bool
	// 自定义重试条件
	RetryCondition func(error) bool
	// 重试前的回调
	BeforeRetry func(attempt int, err error)
	// 重试后的回调
	AfterRetry func(attempt int, err error, delay time.Duration)
}

// NewEnhancedRetryConfig 创建增强的重试配置
func NewEnhancedRetryConfig() *EnhancedRetryConfig {
	return &EnhancedRetryConfig{
		RetryConfig:   DefaultRetryConfig(),
		ErrorRetryMap: make(map[string]bool),
		RetryCondition: func(err error) bool {
			return IsRetryableError(err)
		},
		BeforeRetry: func(attempt int, err error) {
			// 默认实现：记录重试信息
		},
		AfterRetry: func(attempt int, err error, delay time.Duration) {
			// 默认实现：记录重试完成信息
		},
	}
}

// AddRetryableError 添加可重试的错误类型
func (erc *EnhancedRetryConfig) AddRetryableError(errorType string) {
	erc.ErrorRetryMap[errorType] = true
}

// SetRetryCondition 设置自定义重试条件
func (erc *EnhancedRetryConfig) SetRetryCondition(condition func(error) bool) {
	erc.RetryCondition = condition
}

// SetBeforeRetry 设置重试前回调
func (erc *EnhancedRetryConfig) SetBeforeRetry(callback func(attempt int, err error)) {
	erc.BeforeRetry = callback
}

// SetAfterRetry 设置重试后回调
func (erc *EnhancedRetryConfig) SetAfterRetry(callback func(attempt int, err error, delay time.Duration)) {
	erc.AfterRetry = callback
}
