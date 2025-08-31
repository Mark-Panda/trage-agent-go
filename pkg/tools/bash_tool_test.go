package tools

import (
	"context"
	"testing"
)

func TestNewBashTool(t *testing.T) {
	tool := NewBashTool()
	
	if tool.GetName() != "bash" {
		t.Errorf("Expected tool name 'bash', got '%s'", tool.GetName())
	}
	
	if tool.GetDescription() != "执行bash命令并返回结果" {
		t.Errorf("Expected description '执行bash命令并返回结果', got '%s'", tool.GetDescription())
	}
	
	params := tool.GetParameters()
	if len(params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(params))
	}
	
	// 检查command参数
	commandParam := params[0]
	if commandParam.Name != "command" {
		t.Errorf("Expected first parameter name 'command', got '%s'", commandParam.Name)
	}
	if !commandParam.Required {
		t.Errorf("Expected command parameter to be required")
	}
}

func TestBashTool_Execute(t *testing.T) {
	tool := NewBashTool()
	ctx := context.Background()
	
	// 测试有效参数
	args := ToolCallArguments{
		"command": "echo 'Hello World'",
	}
	
	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	
	if result.Result == "" {
		t.Errorf("Expected result, got empty string")
	}
}

func TestBashTool_Execute_InvalidArgs(t *testing.T) {
	tool := NewBashTool()
	ctx := context.Background()
	
	// 测试缺少必需参数
	args := ToolCallArguments{}
	
	_, err := tool.Execute(ctx, args)
	if err == nil {
		t.Errorf("Expected error for missing command parameter")
	}
	
	// 测试错误参数类型
	args = ToolCallArguments{
		"command": 123, // 应该是string
	}
	
	_, err = tool.Execute(ctx, args)
	if err == nil {
		t.Errorf("Expected error for wrong parameter type")
	}
}

func TestBashTool_ValidateArgs(t *testing.T) {
	tool := NewBashTool()
	
	// 测试有效参数
	args := ToolCallArguments{
		"command": "echo 'test'",
	}
	
	err := tool.ValidateArgs(args)
	if err != nil {
		t.Errorf("Expected no error for valid args, got %v", err)
	}
	
	// 测试缺少必需参数
	args = ToolCallArguments{}
	
	err = tool.ValidateArgs(args)
	if err == nil {
		t.Errorf("Expected error for missing required args")
	}
}
