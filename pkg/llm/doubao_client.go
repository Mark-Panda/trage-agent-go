package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
				// 清理参数字符串，移除可能的无效字符
				cleanArgs := strings.TrimSpace(tc.Function.Arguments)
				// 尝试解析JSON字符串为map
				if err := json.Unmarshal([]byte(cleanArgs), &arguments); err != nil {
					// 如果解析失败，尝试手动解析关键参数
					fmt.Printf("Warning: failed to parse tool call arguments: %v, attempting manual parsing\n", err)
					arguments = dc.manualParseArguments(cleanArgs)
				}
			} else {
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
		// 确保消息有内容，豆包API要求messages.content不能为空
		content := msg.Content
		if content == "" {
			content = " " // 使用空格作为默认内容
		}

		result[i] = openai.ChatCompletionMessage{
			Role:      msg.Role,
			Content:   content,
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

// manualParseArguments 手动解析工具调用参数
func (dc *DoubaoClient) manualParseArguments(argsStr string) map[string]interface{} {
	arguments := make(map[string]interface{})

	// 尝试提取常见的参数模式
	// 查找 file_path, content, command 等关键参数

	// 查找 file_path
	if strings.Contains(argsStr, "file_path") {
		// 简单的字符串提取，查找引号内的内容
		if start := strings.Index(argsStr, `"file_path"`); start != -1 {
			if colon := strings.Index(argsStr[start:], ":"); colon != -1 {
				if quoteStart := strings.Index(argsStr[start+colon:], `"`); quoteStart != -1 {
					startPos := start + colon + quoteStart + 1
					if quoteEnd := strings.Index(argsStr[startPos:], `"`); quoteEnd != -1 {
						filePath := argsStr[startPos : startPos+quoteEnd]
						arguments["file_path"] = filePath
					}
				}
			}
		}
	}

	// 查找 content
	if strings.Contains(argsStr, "content") {
		if start := strings.Index(argsStr, `"content"`); start != -1 {
			if colon := strings.Index(argsStr[start:], ":"); colon != -1 {
				if quoteStart := strings.Index(argsStr[start+colon:], `"`); quoteStart != -1 {
					startPos := start + colon + quoteStart + 1
					if quoteEnd := strings.Index(argsStr[startPos:], `"`); quoteEnd != -1 {
						content := argsStr[startPos : startPos+quoteEnd]
						arguments["content"] = content
					}
				}
			}
		}
	}

	// 查找 command
	if strings.Contains(argsStr, "command") {
		if start := strings.Index(argsStr, `"command"`); start != -1 {
			if colon := strings.Index(argsStr[start:], ":"); colon != -1 {
				if quoteStart := strings.Index(argsStr[start+colon:], `"`); quoteStart != -1 {
					startPos := start + colon + quoteStart + 1
					if quoteEnd := strings.Index(argsStr[startPos:], `"`); quoteEnd != -1 {
						command := argsStr[startPos : startPos+quoteEnd]
						arguments["command"] = command
					}
				}
			}
		}
	}

	// 如果没有找到任何参数，返回空map
	if len(arguments) == 0 {
		fmt.Printf("Warning: manual parsing failed, no valid arguments found in: %s\n", argsStr)
	}

	return arguments
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
