package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"trage-agent-go/pkg/llm"
)

// TrajectoryRecorder 轨迹记录器
type TrajectoryRecorder struct {
	filePath    string
	messages    []llm.LLMMessage
	toolCalls   []llm.ToolCall
	toolResults []interface{}
	metadata    map[string]interface{}
	startTime   time.Time
}

// NewTrajectoryRecorder 创建新的轨迹记录器
func NewTrajectoryRecorder(filePath string) *TrajectoryRecorder {
	if filePath == "" {
		filePath = generateTrajectoryPath()
	}

	return &TrajectoryRecorder{
		filePath:    filePath,
		messages:    make([]llm.LLMMessage, 0),
		toolCalls:   make([]llm.ToolCall, 0),
		toolResults: make([]interface{}, 0),
		metadata:    make(map[string]interface{}),
		startTime:   time.Now(),
	}
}

// RecordMessage 记录消息
func (tr *TrajectoryRecorder) RecordMessage(message llm.LLMMessage) error {
	tr.messages = append(tr.messages, message)
	return nil
}

// RecordToolCall 记录工具调用
func (tr *TrajectoryRecorder) RecordToolCall(toolCall llm.ToolCall) error {
	tr.toolCalls = append(tr.toolCalls, toolCall)
	return nil
}

// RecordToolResult 记录工具结果
func (tr *TrajectoryRecorder) RecordToolResult(toolCall llm.ToolCall, result interface{}) error {
	tr.toolResults = append(tr.toolResults, result)
	return nil
}

// Save 保存轨迹
func (tr *TrajectoryRecorder) Save() error {
	// 确保目录存在
	dir := filepath.Dir(tr.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 构建轨迹数据
	trajectory := map[string]interface{}{
		"metadata": map[string]interface{}{
			"start_time":   tr.startTime.Format(time.RFC3339),
			"end_time":     time.Now().Format(time.RFC3339),
			"duration":     time.Since(tr.startTime).String(),
			"total_messages": len(tr.messages),
			"total_tool_calls": len(tr.toolCalls),
			"total_tool_results": len(tr.toolResults),
		},
		"messages":     tr.messages,
		"tool_calls":  tr.toolCalls,
		"tool_results": tr.toolResults,
		"custom_metadata": tr.metadata,
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(trajectory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal trajectory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(tr.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write trajectory file: %w", err)
	}

	return nil
}

// GetTrajectoryPath 获取轨迹文件路径
func (tr *TrajectoryRecorder) GetTrajectoryPath() string {
	return tr.filePath
}

// AddMetadata 添加元数据
func (tr *TrajectoryRecorder) AddMetadata(key string, value interface{}) {
	tr.metadata[key] = value
}

// GetMetadata 获取元数据
func (tr *TrajectoryRecorder) GetMetadata(key string) (interface{}, bool) {
	value, exists := tr.metadata[key]
	return value, exists
}

// GetMessageCount 获取消息数量
func (tr *TrajectoryRecorder) GetMessageCount() int {
	return len(tr.messages)
}

// GetToolCallCount 获取工具调用数量
func (tr *TrajectoryRecorder) GetToolCallCount() int {
	return len(tr.toolCalls)
}

// GetDuration 获取执行持续时间
func (tr *TrajectoryRecorder) GetDuration() time.Duration {
	return time.Since(tr.startTime)
}

// Reset 重置记录器
func (tr *TrajectoryRecorder) Reset() {
	tr.messages = make([]llm.LLMMessage, 0)
	tr.toolCalls = make([]llm.ToolCall, 0)
	tr.toolResults = make([]interface{}, 0)
	tr.metadata = make(map[string]interface{})
	tr.startTime = time.Now()
}

// generateTrajectoryPath 生成轨迹文件路径
func generateTrajectoryPath() string {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("trajectory_%s.json", timestamp)
	
	// 创建trajectories目录
	dir := "trajectories"
	if err := os.MkdirAll(dir, 0755); err == nil {
		return filepath.Join(dir, filename)
	}
	
	// 如果无法创建目录，返回当前目录下的文件名
	return filename
}

// LoadTrajectory 加载轨迹文件
func LoadTrajectory(filePath string) (*TrajectoryRecorder, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read trajectory file: %w", err)
	}

	var trajectory map[string]interface{}
	if err := json.Unmarshal(data, &trajectory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trajectory: %w", err)
	}

	// 创建新的记录器
	recorder := NewTrajectoryRecorder(filePath)

	// 恢复消息
	if messagesData, exists := trajectory["messages"]; exists {
		if messagesBytes, err := json.Marshal(messagesData); err == nil {
			var messages []llm.LLMMessage
			if err := json.Unmarshal(messagesBytes, &messages); err == nil {
				recorder.messages = messages
			}
		}
	}

	// 恢复工具调用
	if toolCallsData, exists := trajectory["tool_calls"]; exists {
		if toolCallsBytes, err := json.Marshal(toolCallsData); err == nil {
			var toolCalls []llm.ToolCall
			if err := json.Unmarshal(toolCallsBytes, &toolCalls); err == nil {
				recorder.toolCalls = toolCalls
			}
		}
	}

	// 恢复元数据
	if metadataData, exists := trajectory["custom_metadata"]; exists {
		if metadataBytes, err := json.Marshal(metadataData); err == nil {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataBytes, &metadata); err == nil {
				recorder.metadata = metadata
			}
		}
	}

	return recorder, nil
}

// GetTrajectorySummary 获取轨迹摘要
func (tr *TrajectoryRecorder) GetTrajectorySummary() map[string]interface{} {
	return map[string]interface{}{
		"file_path":        tr.filePath,
		"message_count":    len(tr.messages),
		"tool_call_count":  len(tr.toolCalls),
		"duration":         tr.GetDuration().String(),
		"start_time":       tr.startTime.Format(time.RFC3339),
		"end_time":         time.Now().Format(time.RFC3339),
	}
}

// ExportToFormat 导出为指定格式
func (tr *TrajectoryRecorder) ExportToFormat(format string, outputPath string) error {
	switch format {
	case "json":
		return tr.exportToJSON(outputPath)
	case "yaml":
		return tr.exportToYAML(outputPath)
	case "txt":
		return tr.exportToText(outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportToJSON 导出为JSON格式
func (tr *TrajectoryRecorder) exportToJSON(outputPath string) error {
	return tr.Save()
}

// exportToYAML 导出为YAML格式
func (tr *TrajectoryRecorder) exportToYAML(outputPath string) error {
	// 这里需要实现YAML导出
	// 暂时返回错误
	return fmt.Errorf("YAML export not implemented yet")
}

// exportToText 导出为文本格式
func (tr *TrajectoryRecorder) exportToText(outputPath string) error {
	// 构建文本内容
	content := fmt.Sprintf("Trae Agent 执行轨迹\n")
	content += fmt.Sprintf("==================\n\n")
	content += fmt.Sprintf("开始时间: %s\n", tr.startTime.Format(time.RFC3339))
	content += fmt.Sprintf("结束时间: %s\n", time.Now().Format(time.RFC3339))
	content += fmt.Sprintf("持续时间: %s\n", tr.GetDuration().String())
	content += fmt.Sprintf("消息数量: %d\n", len(tr.messages))
	content += fmt.Sprintf("工具调用数量: %d\n", len(tr.toolCalls))
	content += fmt.Sprintf("\n")

	// 添加消息
	content += fmt.Sprintf("消息记录:\n")
	content += fmt.Sprintf("--------\n")
	for i, msg := range tr.messages {
		content += fmt.Sprintf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
	}
	content += fmt.Sprintf("\n")

	// 添加工具调用
	content += fmt.Sprintf("工具调用记录:\n")
	content += fmt.Sprintf("------------\n")
	for i, toolCall := range tr.toolCalls {
		content += fmt.Sprintf("%d. %s\n", i+1, toolCall.Function.Name)
	}

	// 写入文件
	return os.WriteFile(outputPath, []byte(content), 0644)
}
