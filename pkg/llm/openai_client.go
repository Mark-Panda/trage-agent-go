package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient OpenAI客户端实现
type OpenAIClient struct {
	*BaseLLMClient
	httpClient *http.Client
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(apiKey, baseURL, apiVersion string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	if apiVersion == "" {
		apiVersion = "v1"
	}

	return &OpenAIClient{
		BaseLLMClient: NewBaseLLMClient(apiKey, baseURL, apiVersion, "openai"),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Chat 实现OpenAI聊天接口
func (oac *OpenAIClient) Chat(messages []LLMMessage, tools []Tool, config ModelConfig) (*LLMMessage, error) {
	// 构建OpenAI API请求
	requestBody := map[string]interface{}{
		"model":       config.GetModel(),
		"messages":    oac.convertMessages(messages),
		"max_tokens":  config.GetMaxTokens(),
		"temperature": config.GetTemperature(),
	}

	// 如果支持工具调用且有工具，添加工具定义
	if config.GetSupportsToolCalling() && len(tools) > 0 {
		requestBody["tools"] = oac.convertTools(tools)
		requestBody["tool_choice"] = "auto"
	}

	// 序列化请求体
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/%s/chat/completions", oac.BaseURL, oac.APIVersion)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oac.APIKey)

	// 发送请求
	resp, err := oac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s, body: %s", resp.Status, string(body))
	}

	// 解析响应
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查是否有选择
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := openAIResp.Choices[0]

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
				Type: tc.Type,
				Function: ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return response, nil
}

// convertMessages 转换消息格式
func (oac *OpenAIClient) convertMessages(messages []LLMMessage) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		result[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		if msg.Name != "" {
			result[i]["name"] = msg.Name
		}
		if msg.ToolCallID != "" {
			result[i]["tool_call_id"] = msg.ToolCallID
		}
	}
	return result
}

// convertTools 转换工具定义格式
func (oac *OpenAIClient) convertTools(tools []Tool) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"parameters":  tool.Function.Parameters,
			},
		}
	}
	return result
}

// OpenAIResponse OpenAI API响应结构
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
