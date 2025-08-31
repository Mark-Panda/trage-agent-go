package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// BashTool Bash工具实现
type BashTool struct {
	*BaseTool
	timeout time.Duration
}

// NewBashTool 创建Bash工具
func NewBashTool() *BashTool {
	parameters := []ToolParameter{
		{
			Name:        "command",
			Type:        "string",
			Description: "要执行的bash命令",
			Required:    true,
		},
		{
			Name:        "timeout",
			Type:        "integer",
			Description: "命令超时时间（秒），默认120秒",
			Required:    false,
		},
	}

	return &BashTool{
		BaseTool: NewBaseTool(
			"bash",
			"执行bash命令并返回结果",
			"",
			parameters,
		),
		timeout: 120 * time.Second,
	}
}

// Execute 执行Bash命令
func (bt *BashTool) Execute(ctx context.Context, args ToolCallArguments) (*ToolResult, error) {
	// 验证参数
	if err := bt.ValidateArgs(args); err != nil {
		return nil, err
	}

	command, ok := args["command"].(string)
	if !ok {
		return nil, &ToolError{
			Message: "command parameter must be a string",
			Code:    400,
		}
	}

	// 检查超时参数
	if timeoutStr, exists := args["timeout"]; exists {
		if timeout, ok := timeoutStr.(float64); ok {
			bt.timeout = time.Duration(timeout) * time.Second
		}
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, bt.timeout)
	defer cancel()

	// 根据操作系统选择shell
	var shell string
	var shellArgs []string

	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArgs = []string{"/C"}
	} else {
		shell = "/bin/bash"
		shellArgs = []string{"-c"}
	}

	// 构建完整命令
	shellArgs = append(shellArgs, command)

	// 创建命令
	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	
	// 设置工作目录为当前目录
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	
	// 检查是否超时
	if ctx.Err() == context.DeadlineExceeded {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("command timed out after %v", bt.timeout),
		}, nil
	}

	// 处理执行结果
	if err != nil {
		// 命令执行失败，但可能有输出
		outputStr := strings.TrimSpace(string(output))
		if outputStr == "" {
			outputStr = "命令执行失败，无输出"
		}
		
		return &ToolResult{
			Success: false,
			Result:  outputStr,
			Error:   err.Error(),
		}, nil
	}

	// 命令执行成功
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		outputStr = "命令执行成功，无输出"
	}

	return &ToolResult{
		Success: true,
		Result:  outputStr,
	}, nil
}

// ValidateArgs 验证参数
func (bt *BashTool) ValidateArgs(args ToolCallArguments) error {
	// 调用基础验证
	if err := bt.BaseTool.ValidateArgs(args); err != nil {
		return err
	}

	// 检查命令是否为空
	if command, exists := args["command"]; exists {
		if cmdStr, ok := command.(string); ok {
			if strings.TrimSpace(cmdStr) == "" {
				return &ToolError{
					Message: "command cannot be empty",
					Code:    400,
				}
			}
		} else {
			return &ToolError{
				Message: "command must be a string",
				Code:    400,
			}
		}
	}

	return nil
}

// GetTimeout 获取超时时间
func (bt *BashTool) GetTimeout() time.Duration {
	return bt.timeout
}

// SetTimeout 设置超时时间
func (bt *BashTool) SetTimeout(timeout time.Duration) {
	bt.timeout = timeout
}

// IsWindows 检查是否为Windows系统
func (bt *BashTool) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// GetShell 获取当前系统的shell
func (bt *BashTool) GetShell() (string, []string) {
	if bt.IsWindows() {
		return "cmd", []string{"/C"}
	}
	return "/bin/bash", []string{"-c"}
}
