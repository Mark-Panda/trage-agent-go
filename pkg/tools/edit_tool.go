package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// EditTool 文件编辑工具
type EditTool struct {
	*BaseTool
}

// NewEditTool 创建编辑工具
func NewEditTool() *EditTool {
	parameters := []ToolParameter{
		{
			Name:        "file_path",
			Type:        "string",
			Description: "要编辑的文件路径",
			Required:    true,
		},
		{
			Name:        "content",
			Type:        "string",
			Description: "新的文件内容",
			Required:    true,
		},
		{
			Name:        "mode",
			Type:        "string",
			Description: "编辑模式：'replace'（替换整个文件）或'append'（追加到文件末尾）",
			Required:    false,
		},
		{
			Name:        "backup",
			Type:        "boolean",
			Description: "是否创建备份文件",
			Required:    false,
		},
	}

	return &EditTool{
		BaseTool: NewBaseTool(
			"edit_file",
			"编辑文件内容，支持替换整个文件或追加内容",
			"",
			parameters,
		),
	}
}

// Execute 执行文件编辑
func (et *EditTool) Execute(ctx context.Context, args ToolCallArguments) (*ToolResult, error) {
	// 验证参数
	if err := et.ValidateArgs(args); err != nil {
		return nil, err
	}

	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)
	
	// 获取编辑模式，默认为替换
	mode := "replace"
	if modeStr, exists := args["mode"]; exists {
		if m, ok := modeStr.(string); ok {
			mode = m
		}
	}
	
	// 获取是否备份，默认为true
	backup := true
	if backupFlag, exists := args["backup"]; exists {
		if b, ok := backupFlag.(bool); ok {
			backup = b
		}
	}

	// 检查文件是否存在
	fileExists := false
	if _, err := os.Stat(filePath); err == nil {
		fileExists = true
	}

	// 如果需要备份且文件存在
	if backup && fileExists {
		if err := et.createBackup(filePath); err != nil {
			return nil, &ToolError{
				Message: fmt.Sprintf("failed to create backup: %v", err),
				Code:    500,
			}
		}
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, &ToolError{
			Message: fmt.Sprintf("failed to create directory: %v", err),
			Code:    500,
		}
	}

	var finalContent string
	var operation string

	switch mode {
	case "replace":
		finalContent = content
		operation = "替换"
	case "append":
		if fileExists {
			// 读取现有内容
			existingContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, &ToolError{
					Message: fmt.Sprintf("failed to read existing file: %v", err),
					Code:    500,
				}
			}
			finalContent = string(existingContent) + "\n" + content
		} else {
			finalContent = content
		}
		operation = "追加"
	default:
		return nil, &ToolError{
			Message: fmt.Sprintf("unsupported mode: %s", mode),
			Code:    400,
		}
	}

	// 写入文件
	if err := ioutil.WriteFile(filePath, []byte(finalContent), 0644); err != nil {
		return nil, &ToolError{
			Message: fmt.Sprintf("failed to write file: %v", err),
			Code:    500,
		}
	}

	// 构建结果消息
	var resultMsg string
	if fileExists {
		resultMsg = fmt.Sprintf("文件 '%s' 已成功%s", filePath, operation)
	} else {
		resultMsg = fmt.Sprintf("文件 '%s' 已成功创建", filePath)
	}

	if backup && fileExists {
		resultMsg += "（已创建备份）"
	}

	return &ToolResult{
		Success: true,
		Result:  resultMsg,
	}, nil
}

// createBackup 创建备份文件
func (et *EditTool) createBackup(filePath string) error {
	backupPath := filePath + ".backup"
	
	// 读取原文件
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	// 写入备份文件
	return ioutil.WriteFile(backupPath, content, 0644)
}

// ValidateArgs 验证参数
func (et *EditTool) ValidateArgs(args ToolCallArguments) error {
	// 调用基础验证
	if err := et.BaseTool.ValidateArgs(args); err != nil {
		return err
	}

	// 检查文件路径
	if filePath, exists := args["file_path"]; exists {
		if path, ok := filePath.(string); ok {
			if strings.TrimSpace(path) == "" {
				return &ToolError{
					Message: "file_path cannot be empty",
					Code:    400,
				}
			}
			
			// 检查路径是否包含非法字符
			if strings.Contains(path, "..") {
				return &ToolError{
					Message: "file_path cannot contain '..'",
					Code:    400,
				}
			}
		} else {
			return &ToolError{
				Message: "file_path must be a string",
				Code:    400,
			}
		}
	}

	// 检查内容
	if content, exists := args["content"]; exists {
		if _, ok := content.(string); !ok {
			return &ToolError{
				Message: "content must be a string",
				Code:    400,
			}
		}
	}

	// 检查模式
	if mode, exists := args["mode"]; exists {
		if modeStr, ok := mode.(string); ok {
			if modeStr != "replace" && modeStr != "append" {
				return &ToolError{
					Message: "mode must be either 'replace' or 'append'",
					Code:    400,
				}
			}
		} else {
			return &ToolError{
				Message: "mode must be a string",
				Code:    400,
			}
		}
	}

	return nil
}

// GetFileInfo 获取文件信息
func (et *EditTool) GetFileInfo(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}

// ReadFile 读取文件内容
func (et *EditTool) ReadFile(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile 写入文件内容
func (et *EditTool) WriteFile(filePath, content string) error {
	return ioutil.WriteFile(filePath, []byte(content), 0644)
}

// FileExists 检查文件是否存在
func (et *EditTool) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// IsDirectory 检查路径是否为目录
func (et *EditTool) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
