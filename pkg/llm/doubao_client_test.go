package llm

import (
	"testing"
)

func TestNewDoubaoClient(t *testing.T) {
	// 测试创建豆包客户端
	client := NewDoubaoClient("test_key", "https://api.doubao.com", "v1")
	
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	
	if client.client == nil {
		t.Fatal("Expected OpenAI client to be initialized")
	}
	
	if client.GetProvider() != "doubao" {
		t.Errorf("Expected provider 'doubao', got '%s'", client.GetProvider())
	}
}

func TestDoubaoClient_ConvertMessages(t *testing.T) {
	client := NewDoubaoClient("test_key", "", "")
	
	messages := []LLMMessage{
		{
			Role:    "user",
			Content: "Hello",
		},
		{
			Role:    "assistant",
			Content: "Hi there!",
		},
	}
	
	converted := client.convertMessages(messages)
	
	if len(converted) != 2 {
		t.Errorf("Expected 2 converted messages, got %d", len(converted))
	}
	
	if converted[0].Role != "user" {
		t.Errorf("Expected first message role 'user', got '%s'", converted[0].Role)
	}
	
	if converted[0].Content != "Hello" {
		t.Errorf("Expected first message content 'Hello', got '%s'", converted[0].Content)
	}
}

func TestDoubaoClient_ConvertTools(t *testing.T) {
	client := NewDoubaoClient("test_key", "", "")
	
	tools := []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "test_tool",
				Description: "A test tool",
				Parameters: map[string]interface{}{
					"param1": "string",
				},
			},
		},
	}
	
	converted := client.convertTools(tools)
	
	if len(converted) != 1 {
		t.Errorf("Expected 1 converted tool, got %d", len(converted))
	}
	
	if converted[0].Type != "function" {
		t.Errorf("Expected tool type 'function', got '%s'", converted[0].Type)
	}
	
	if converted[0].Function.Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", converted[0].Function.Name)
	}
	
	if converted[0].Function.Description != "A test tool" {
		t.Errorf("Expected tool description 'A test tool', got '%s'", converted[0].Function.Description)
	}
}
