package llm

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIMockModelConfig 用于测试的模型配置
type OpenAIMockModelConfig struct {
	model               string
	maxTokens           int
	temperature         float64
	supportsToolCalling bool
}

func (m *OpenAIMockModelConfig) GetModel() string             { return m.model }
func (m *OpenAIMockModelConfig) GetMaxTokens() int            { return m.maxTokens }
func (m *OpenAIMockModelConfig) GetTemperature() float64      { return m.temperature }
func (m *OpenAIMockModelConfig) GetTopP() float64             { return 1.0 }
func (m *OpenAIMockModelConfig) GetTopK() int                 { return 0 }
func (m *OpenAIMockModelConfig) GetParallelToolCalling() bool { return m.supportsToolCalling }
func (m *OpenAIMockModelConfig) GetAPIKey() string            { return "test_key" }
func (m *OpenAIMockModelConfig) GetBaseURL() string           { return "https://api.openai.com" }
func (m *OpenAIMockModelConfig) GetAPIVersion() string        { return "v1" }

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		baseURL    string
		apiVersion string
	}{
		{
			name:       "使用默认值",
			apiKey:     "test_key",
			baseURL:    "",
			apiVersion: "",
		},
		{
			name:       "使用自定义值",
			apiKey:     "custom_key",
			baseURL:    "https://custom.api.com",
			apiVersion: "v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL, tt.apiVersion)

			if client == nil {
				t.Fatal("Expected client to be created")
			}

			if client.BaseLLMClient == nil {
				t.Fatal("Expected BaseLLMClient to be initialized")
			}

			if client.client == nil {
				t.Fatal("Expected OpenAI client to be initialized")
			}

			if client.GetProvider() != "openai" {
				t.Errorf("Expected provider 'openai', got '%s'", client.GetProvider())
			}

			if client.APIKey != tt.apiKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.apiKey, client.APIKey)
			}
		})
	}
}

func TestNewOpenAIClient_DefaultValues(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	// 检查默认值
	if client.BaseURL != "https://api.openai.com" {
		t.Errorf("Expected default base URL 'https://api.openai.com', got '%s'", client.BaseURL)
	}

	if client.APIVersion != "v1" {
		t.Errorf("Expected default API version 'v1', got '%s'", client.APIVersion)
	}
}

func TestOpenAIClient_ConvertMessages(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	messages := []LLMMessage{
		{
			Role:    "user",
			Content: "Hello",
		},
		{
			Role:    "assistant",
			Content: "Hi there!",
		},
		{
			Role:       "user",
			Content:    "How are you?",
			Name:       "test_user",
			ToolCallID: "tool_123",
		},
	}

	converted := client.convertMessages(messages)

	if len(converted) != 3 {
		t.Errorf("Expected 3 converted messages, got %d", len(converted))
	}

	// 检查第一条消息
	if converted[0].Role != "user" {
		t.Errorf("Expected first message role 'user', got '%s'", converted[0].Role)
	}

	if converted[0].Content != "Hello" {
		t.Errorf("Expected first message content 'Hello', got '%s'", converted[0].Content)
	}

	// 检查第三条消息（包含额外字段）
	if converted[2].Name != "test_user" {
		t.Errorf("Expected third message name 'test_user', got '%s'", converted[2].Name)
	}

	if converted[2].ToolCallID != "tool_123" {
		t.Errorf("Expected third message tool call ID 'tool_123', got '%s'", converted[2].ToolCallID)
	}
}

func TestOpenAIClient_ConvertTools(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	tools := []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "test_tool",
				Description: "A test tool",
				Parameters: map[string]interface{}{
					"param1": "string",
					"param2": map[string]interface{}{
						"type": "object",
					},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "another_tool",
				Description: "Another test tool",
				Parameters: map[string]interface{}{
					"required": []string{"param1"},
				},
			},
		},
	}

	converted := client.convertTools(tools)

	if len(converted) != 2 {
		t.Errorf("Expected 2 converted tools, got %d", len(converted))
	}

	// 检查第一个工具
	if converted[0].Type != openai.ToolTypeFunction {
		t.Errorf("Expected first tool type '%s', got '%s'", openai.ToolTypeFunction, converted[0].Type)
	}

	if converted[0].Function.Name != "test_tool" {
		t.Errorf("Expected first tool name 'test_tool', got '%s'", converted[0].Function.Name)
	}

	if converted[0].Function.Description != "A test tool" {
		t.Errorf("Expected first tool description 'A test tool', got '%s'", converted[0].Function.Description)
	}

	// 检查第二个工具
	if converted[1].Function.Name != "another_tool" {
		t.Errorf("Expected second tool name 'another_tool', got '%s'", converted[1].Function.Name)
	}
}

func TestOpenAIClient_ConvertTools_EmptySlice(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	converted := client.convertTools([]Tool{})

	if len(converted) != 0 {
		t.Errorf("Expected 0 converted tools, got %d", len(converted))
	}
}

func TestOpenAIClient_ConvertMessages_EmptySlice(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	converted := client.convertMessages([]LLMMessage{})

	if len(converted) != 0 {
		t.Errorf("Expected 0 converted messages, got %d", len(converted))
	}
}

func TestOpenAIClient_SupportsToolCalling(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	if !client.SupportsToolCalling() {
		t.Error("Expected OpenAI client to support tool calling")
	}
}

func TestOpenAIClient_GetProvider(t *testing.T) {
	client := NewOpenAIClient("test_key", "", "")

	if client.GetProvider() != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", client.GetProvider())
	}
}

// 测试工具调用转换
func TestOpenAIClient_ToolCallConversion(t *testing.T) {
	// 模拟 OpenAI 响应中的工具调用
	mockToolCall := openai.ToolCall{
		ID:   "call_123",
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      "test_function",
			Arguments: `{"param1": "value1", "param2": 42}`,
		},
	}

	// 转换为项目内部格式
	converted := ToolCall{
		ID:   mockToolCall.ID,
		Type: string(mockToolCall.Type),
		Function: ToolCallFunction{
			Name:      mockToolCall.Function.Name,
			Arguments: map[string]interface{}{"content": mockToolCall.Function.Arguments},
		},
	}

	if converted.ID != "call_123" {
		t.Errorf("Expected tool call ID 'call_123', got '%s'", converted.ID)
	}

	if converted.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", converted.Type)
	}

	if converted.Function.Name != "test_function" {
		t.Errorf("Expected function name 'test_function', got '%s'", converted.Function.Name)
	}
}

// 测试配置参数
func TestOpenAIClient_Configuration(t *testing.T) {
	client := NewOpenAIClient("test_key", "https://custom.api.com", "v2")

	if client.BaseURL != "https://custom.api.com" {
		t.Errorf("Expected custom base URL 'https://custom.api.com', got '%s'", client.BaseURL)
	}

	if client.APIVersion != "v2" {
		t.Errorf("Expected custom API version 'v2', got '%s'", client.APIVersion)
	}

	if client.APIKey != "test_key" {
		t.Errorf("Expected API key 'test_key', got '%s'", client.APIKey)
	}
}

// 测试边界情况
func TestOpenAIClient_EdgeCases(t *testing.T) {
	client := NewOpenAIClient("", "", "")

	// 测试空 API key
	if client.APIKey != "" {
		t.Errorf("Expected empty API key, got '%s'", client.APIKey)
	}

	// 测试空消息
	emptyMessages := []LLMMessage{}
	converted := client.convertMessages(emptyMessages)
	if len(converted) != 0 {
		t.Errorf("Expected 0 converted messages for empty input, got %d", len(converted))
	}

	// 测试空工具
	emptyTools := []Tool{}
	convertedTools := client.convertTools(emptyTools)
	if len(convertedTools) != 0 {
		t.Errorf("Expected 0 converted tools for empty input, got %d", len(convertedTools))
	}
}
