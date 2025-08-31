package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// DoubaoClient 豆包客户端实现
type DoubaoClient struct {
	*BaseLLMClient
	client *openai.Client
}

// NewDoubaoClient 创建豆包客户端
func NewDoubaoClient(apiKey, baseURL, apiVersion string) *DoubaoClient {
	// 创建OpenAI兼容的客户端配置
	config := openai.DefaultConfig(apiKey)

	// 豆包使用不同的基础URL
	if baseURL == "" {
		baseURL = "https://api.doubao.com"
	}
	config.BaseURL = baseURL

	// 如果提供了API版本，使用它
	if apiVersion != "" {
		config.APIVersion = apiVersion
	}

	// 创建客户端
	client := openai.NewClientWithConfig(config)

	return &DoubaoClient{
		BaseLLMClient: NewBaseLLMClient(apiKey, baseURL, apiVersion, "doubao"),
		client:        client,
	}
}

// Chat 实现豆包聊天接口
func (dc *DoubaoClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	// 转换消息格式
	openAIMessages := dc.convertMessages(messages)

	// 创建请求
	req := openai.ChatCompletionRequest{
		Model:       config.GetModel(),
		Messages:    openAIMessages,
		MaxTokens:   config.GetMaxTokens(),
		Temperature: float32(config.GetTemperature()),
	}

	// 如果支持工具调用且有工具，添加工具定义
	if config.GetSupportsToolCalling() && len(tools) > 0 {
		req.Tools = dc.convertTools(tools)
		req.ToolChoice = "auto"
	}

	// 调用豆包API
	resp, err := dc.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("Doubao API call failed: %w", err)
	}

	// 检查是否有选择
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in Doubao response")
	}

	choice := resp.Choices[0]

	// 转换为LLMMessage
	response := &LLMMessage{
		Role:    choice.Message.Role,
		Content: choice.Message.Content,
	}

	// 如果有工具调用，转换工具调用
	if len(choice.Message.ToolCalls) > 0 {
		response.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			// 解析工具调用参数
			var arguments map[string]interface{}
			if tc.Function.Arguments != "" {
				// 这里需要解析JSON字符串为map
				// 暂时使用空map，实际使用时需要JSON解析
				arguments = make(map[string]interface{})
			}

			response.ToolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: arguments,
				},
			}
		}
	}

	return response, nil
}

// convertMessages 转换消息格式
func (dc *DoubaoClient) convertMessages(messages []LLMMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		result[i] = openai.ChatCompletionMessage{
			Role:      msg.Role,
			Content:   msg.Content,
			Name:      msg.Name,
			ToolCalls: nil, // 工具调用在响应中处理
		}

		// 如果有工具调用ID，设置它
		if msg.ToolCallID != "" {
			result[i].ToolCallID = msg.ToolCallID
		}
	}
	return result
}

// convertTools 转换工具定义格式
func (dc *DoubaoClient) convertTools(tools []Tool) []openai.Tool {
	result := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		// 转换参数定义
		parameters := make(map[string]interface{})
		if tool.Function.Parameters != nil {
			parameters = tool.Function.Parameters
		}

		result[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  parameters,
			},
		}
	}
	return result
}
