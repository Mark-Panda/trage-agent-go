package llm

import (
	"fmt"
	"strings"
	"time"
)

// LLMMessage LLM消息结构
type LLMMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content"`
	Name       string                 `json:"name,omitempty"`
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall 工具调用结构
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction 工具调用函数结构
type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// LLMResponse LLM响应结构
type LLMResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []Choice               `json:"choices"`
	Usage             Usage                  `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// Choice 选择结构
type Choice struct {
	Index        int                    `json:"index"`
	Message      LLMMessage             `json:"message"`
	Logprobs     interface{}            `json:"logprobs,omitempty"`
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Usage 使用统计结构
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolParameter 工具参数结构
type ToolParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Items       interface{} `json:"items,omitempty"`
}

// Tool 工具定义结构
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数结构
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// LLMClient LLM客户端接口
type LLMClient interface {
	// Chat 发送聊天消息
	Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error)

	// SetTrajectoryRecorder 设置轨迹记录器
	SetTrajectoryRecorder(recorder TrajectoryRecorder)

	// GetProvider 获取提供商名称
	GetProvider() string

	// SupportsToolCalling 检查是否支持工具调用
	SupportsToolCalling() bool
}

// ModelConfig 模型配置接口
type ModelConfig interface {
	GetModel() string
	GetMaxTokens() int
	GetTemperature() float64
	GetTopP() float64
	GetTopK() int
	GetParallelToolCalls() bool
	GetMaxRetries() int
	GetSupportsToolCalling() bool
	GetAPIKey() string
	GetBaseURL() string
	GetAPIVersion() string
}

// TrajectoryRecorder 轨迹记录器接口
type TrajectoryRecorder interface {
	// RecordMessage 记录消息
	RecordMessage(message LLMMessage) error

	// RecordToolCall 记录工具调用
	RecordToolCall(toolCall ToolCall) error

	// RecordToolResult 记录工具结果
	RecordToolResult(toolCall ToolCall, result interface{}) error

	// Save 保存轨迹
	Save() error

	// GetTrajectoryPath 获取轨迹文件路径
	GetTrajectoryPath() string
}

// BaseLLMClient 基础LLM客户端
type BaseLLMClient struct {
	APIKey     string
	BaseURL    string
	APIVersion string
	Recorder   TrajectoryRecorder
	Provider   string
}

// Chat 实现LLMClient接口
func (b *BaseLLMClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	// 简单的测试实现，返回一个工具调用
	// 在实际应用中，这里应该调用真正的LLM API

	// 获取最后一条用户消息
	var userMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			userMessage = messages[i].Content
			break
		}
	}

	// 检查是否是第一次调用（没有工具结果）
	hasToolResults := false
	for _, msg := range messages {
		if msg.Role == "tool" {
			hasToolResults = true
			break
		}
	}

	if !hasToolResults {
		// 第一次调用，返回工具调用
		var toolName string
		var toolArgs map[string]interface{}

		if strings.Contains(strings.ToLower(userMessage), "hello") || strings.Contains(strings.ToLower(userMessage), "world") {
			// 如果是问候消息，调用bash工具
			toolName = "bash"
			toolArgs = map[string]interface{}{
				"command": "echo 'Hello from Trae Agent!'",
			}
		} else {
			// 默认调用edit_file工具
			toolName = "edit_file"
			toolArgs = map[string]interface{}{
				"file_path": "response.txt",
				"content":   fmt.Sprintf("Response to: %s", userMessage),
			}
		}

		// 创建工具调用
		toolCall := ToolCall{
			ID: "test_tool_call_1",
			Function: ToolCallFunction{
				Name:      toolName,
				Arguments: toolArgs,
			},
		}

		// 创建响应消息
		response := &LLMMessage{
			Role:      "assistant",
			Content:   fmt.Sprintf("我将使用%s工具来处理您的请求: %s", toolName, userMessage),
			ToolCalls: []ToolCall{toolCall},
		}

		return response, nil
	} else {
		// 已经有工具结果，返回任务完成消息
		response := &LLMMessage{
			Role:    "assistant",
			Content: "任务已完成！我已经成功处理了您的请求。",
		}

		return response, nil
	}
}

// NewBaseLLMClient 创建基础LLM客户端
func NewBaseLLMClient(apiKey, baseURL, apiVersion, provider string) *BaseLLMClient {
	return &BaseLLMClient{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		APIVersion: apiVersion,
		Provider:   provider,
	}
}

// SetTrajectoryRecorder 设置轨迹记录器
func (b *BaseLLMClient) SetTrajectoryRecorder(recorder TrajectoryRecorder) {
	b.Recorder = recorder
}

// GetProvider 获取提供商名称
func (b *BaseLLMClient) GetProvider() string {
	return b.Provider
}

// SupportsToolCalling 检查是否支持工具调用
func (b *BaseLLMClient) SupportsToolCalling() bool {
	return true
}

// RecordMessage 记录消息
func (b *BaseLLMClient) RecordMessage(message LLMMessage) error {
	if b.Recorder != nil {
		return b.Recorder.RecordMessage(message)
	}
	return nil
}

// RecordToolCall 记录工具调用
func (b *BaseLLMClient) RecordToolCall(toolCall ToolCall) error {
	if b.Recorder != nil {
		return b.Recorder.RecordToolCall(toolCall)
	}
	return nil
}

// RecordToolResult 记录工具结果
func (b *BaseLLMClient) RecordToolResult(toolCall ToolCall, result interface{}) error {
	if b.Recorder != nil {
		return b.Recorder.RecordToolResult(toolCall, result)
	}
	return nil
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	BackoffRate float64
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:  3,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		BackoffRate: 2.0,
	}
}

// Error 错误类型
type Error struct {
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是特定的不可重试错误类型
	if llmErr, ok := err.(*Error); ok {
		switch llmErr.Type {
		case "invalid_request":
			// 无效请求错误通常不可重试
			return false
		case "authentication_error":
			// 认证错误不可重试
			return false
		case "permission_error":
			// 权限错误不可重试
			return false
		case "quota_exceeded":
			// 配额超限错误不可重试
			return false
		}
	}

	// 检查错误消息中是否包含不可重试的关键词
	errMsg := err.Error()
	if strings.Contains(errMsg, "invalid api key") ||
		strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "permission denied") {
		return false
	}

	// 默认情况下，网络错误、超时错误等是可重试的
	return true
}
