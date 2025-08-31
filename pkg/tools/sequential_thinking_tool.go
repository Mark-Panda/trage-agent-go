package tools

import (
	"context"
	"fmt"
)

// SequentialThinkingTool 顺序思考工具实现
type SequentialThinkingTool struct {
	*BaseTool
}

// NewSequentialThinkingTool 创建顺序思考工具
func NewSequentialThinkingTool() *SequentialThinkingTool {
	parameters := []ToolParameter{
		{
			Name:        "thought",
			Type:        "string",
			Description: "思考内容，描述当前步骤的思考过程",
			Required:    true,
		},
		{
			Name:        "step_number",
			Type:        "integer",
			Description: "当前步骤编号",
			Required:    true,
		},
		{
			Name:        "total_steps",
			Type:        "integer",
			Description: "总步骤数",
			Required:    false,
		},
	}

	return &SequentialThinkingTool{
		BaseTool: NewBaseTool(
			"sequential_thinking",
			"记录顺序思考过程，帮助代理进行结构化思考",
			"用于记录代理在解决问题过程中的思考步骤，帮助跟踪推理过程",
			parameters,
		),
	}
}

// Execute 执行顺序思考
func (stt *SequentialThinkingTool) Execute(ctx context.Context, args ToolCallArguments) (*ToolResult, error) {
	// 验证参数
	if err := stt.ValidateArgs(args); err != nil {
		return nil, err
	}

	thought, ok := args["thought"].(string)
	if !ok {
		return nil, &ToolError{
			Message: "thought parameter must be a string",
			Code:    400,
		}
	}

	stepNumber, ok := args["step_number"].(float64)
	if !ok {
		return nil, &ToolError{
			Message: "step_number parameter must be a number",
			Code:    400,
		}
	}

	totalSteps := "未知"
	if totalStepsRaw, exists := args["total_steps"]; exists {
		if total, ok := totalStepsRaw.(float64); ok {
			totalSteps = fmt.Sprintf("%.0f", total)
		}
	}

	// 格式化思考输出
	output := fmt.Sprintf("🤔 思考步骤 %.0f", stepNumber)
	if totalSteps != "未知" {
		output += fmt.Sprintf("/%s", totalSteps)
	}
	output += fmt.Sprintf(": %s", thought)

	// 记录思考过程（这里可以扩展为保存到文件或数据库）
	fmt.Println(output)

	return &ToolResult{
		Success: true,
		Result:  output,
	}, nil
}
