package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient OpenAI客户端实现
type OpenAIClient struct {
	*BaseLLMClient
	client *openai.Client
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(apiKey, baseURL, apiVersion string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	if apiVersion == "" {
		apiVersion = "v1"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	// 设置超时时间
	config.HTTPClient = &http.Client{
		Timeout: 120 * time.Second,
	}

	return &OpenAIClient{
		BaseLLMClient: NewBaseLLMClient(apiKey, baseURL, apiVersion, "openai"),
		client:        openai.NewClientWithConfig(config),
	}
}

// Chat 实现OpenAI聊天接口
func (oac *OpenAIClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	// 转换消息格式
	openaiMessages := oac.convertMessages(messages)

	// 构建请求
	req := openai.ChatCompletionRequest{
		Model:       config.GetModel(),
		Messages:    openaiMessages,
		MaxTokens:   config.GetMaxTokens(),
		Temperature: float32(config.GetTemperature()),
	}

	// 如果支持工具调用且有工具，添加工具定义
	if config.GetSupportsToolCalling() && len(tools) > 0 {
		req.Tools = oac.convertTools(tools)
		req.ToolChoice = "auto"
	}

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := oac.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// 检查是否有选择
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in openai response")
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
			response.ToolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: map[string]interface{}{"content": tc.Function.Arguments},
				},
			}
		}
	}

	return response, nil
}

// convertMessages 转换消息格式
func (oac *OpenAIClient) convertMessages(messages []LLMMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		result[i] = openai.ChatCompletionMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
		}
	}
	return result
}

// convertTools 转换工具定义格式
func (oac *OpenAIClient) convertTools(tools []Tool) []openai.Tool {
	result := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		result[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}
	return result
}
