package tools

import (
	"context"
	"fmt"
)

// TaskDoneTool 任务完成工具实现
type TaskDoneTool struct {
	*BaseTool
}

// NewTaskDoneTool 创建任务完成工具
func NewTaskDoneTool() *TaskDoneTool {
	parameters := []ToolParameter{
		{
			Name:        "summary",
			Type:        "string",
			Description: "任务完成总结，描述完成的工作和结果",
			Required:    true,
		},
		{
			Name:        "success",
			Type:        "boolean",
			Description: "任务是否成功完成",
			Required:    true,
		},
		{
			Name:        "output",
			Type:        "string",
			Description: "任务的最终输出或结果",
			Required:    false,
		},
	}

	return &TaskDoneTool{
		BaseTool: NewBaseTool(
			"task_done",
			"标记任务完成，提供任务总结和结果",
			"用于明确表示任务已完成，提供执行总结和最终结果",
			parameters,
		),
	}
}

// Execute 执行任务完成
func (tdt *TaskDoneTool) Execute(ctx context.Context, args ToolCallArguments) (*ToolResult, error) {
	// 验证参数
	if err := tdt.ValidateArgs(args); err != nil {
		return nil, err
	}

	summary, ok := args["summary"].(string)
	if !ok {
		return nil, &ToolError{
			Message: "summary parameter must be a string",
			Code:    400,
		}
	}

	success, ok := args["success"].(bool)
	if !ok {
		return nil, &ToolError{
			Message: "success parameter must be a boolean",
			Code:    400,
		}
	}

	output := ""
	if outputRaw, exists := args["output"]; exists {
		if outputStr, ok := outputRaw.(string); ok {
			output = outputStr
		}
	}

	// 格式化完成消息
	var statusIcon string
	if success {
		statusIcon = "✅"
	} else {
		statusIcon = "❌"
	}

	result := fmt.Sprintf("%s 任务完成\n总结: %s", statusIcon, summary)
	if output != "" {
		result += fmt.Sprintf("\n输出: %s", output)
	}

	// 打印完成消息
	fmt.Println(result)

	return &ToolResult{
		Success: true,
		Result:  result,
	}, nil
}
