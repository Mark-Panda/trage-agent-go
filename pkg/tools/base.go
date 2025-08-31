package tools

import (
	"context"
	"fmt"
	"time"

	"trage-agent-go/pkg/llm"
)

// ToolError 工具错误
type ToolError struct {
	Message string
	Code    int
}

func (e ToolError) Error() string {
	return e.Message
}

// ToolExecResult 工具执行结果
type ToolExecResult struct {
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode int    `json:"error_code,omitempty"`
}

// ToolResult 工具结果
type ToolResult struct {
	CallID string `json:"call_id"`
	Name   string `json:"name"`
	Success bool  `json:"success"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
	ID     string `json:"id,omitempty"`
}

// ToolCallArguments 工具调用参数
type ToolCallArguments map[string]interface{}

// ToolCall 工具调用
type ToolCall struct {
	Name      string            `json:"name"`
	CallID    string            `json:"call_id"`
	Arguments ToolCallArguments `json:"arguments"`
	ID        string            `json:"id,omitempty"`
}

func (tc ToolCall) String() string {
	return fmt.Sprintf("ToolCall(name=%s, arguments=%v, call_id=%s, id=%s)", 
		tc.Name, tc.Arguments, tc.CallID, tc.ID)
}

// ToolParameter 工具参数
type ToolParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Enum        []string    `json:"enum,omitempty"`
	Items       interface{} `json:"items,omitempty"`
	Required    bool        `json:"required"`
}

// Tool 工具接口
type Tool interface {
	// GetName 获取工具名称
	GetName() string
	
	// GetDescription 获取工具描述
	GetDescription() string
	
	// GetParameters 获取工具参数
	GetParameters() []ToolParameter
	
	// GetModelProvider 获取模型提供商
	GetModelProvider() string
	
	// Execute 执行工具
	Execute(ctx context.Context, args ToolCallArguments) (*ToolResult, error)
	
	// ValidateArgs 验证参数
	ValidateArgs(args ToolCallArguments) error
}

// BaseTool 基础工具实现
type BaseTool struct {
	name          string
	description   string
	parameters    []ToolParameter
	modelProvider string
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, description, modelProvider string, parameters []ToolParameter) *BaseTool {
	return &BaseTool{
		name:          name,
		description:   description,
		parameters:    parameters,
		modelProvider: modelProvider,
	}
}

// GetName 获取工具名称
func (t *BaseTool) GetName() string {
	return t.name
}

// GetDescription 获取工具描述
func (t *BaseTool) GetDescription() string {
	return t.description
}

// GetParameters 获取工具参数
func (t *BaseTool) GetParameters() []ToolParameter {
	return t.parameters
}

// GetModelProvider 获取模型提供商
func (t *BaseTool) GetModelProvider() string {
	return t.modelProvider
}

// ValidateArgs 验证参数（基础实现）
func (t *BaseTool) ValidateArgs(args ToolCallArguments) error {
	for _, param := range t.parameters {
		if param.Required {
			if _, exists := args[param.Name]; !exists {
				return &ToolError{
					Message: fmt.Sprintf("required parameter '%s' is missing", param.Name),
					Code:    400,
				}
			}
		}
	}
	return nil
}

// ToolExecutor 工具执行器
type ToolExecutor struct {
	tools map[string]Tool
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		tools: make(map[string]Tool),
	}
}

// RegisterTool 注册工具
func (te *ToolExecutor) RegisterTool(tool Tool) {
	te.tools[tool.GetName()] = tool
}

// GetTool 获取工具
func (te *ToolExecutor) GetTool(name string) (Tool, bool) {
	tool, exists := te.tools[name]
	return tool, exists
}

// ListTools 列出所有工具
func (te *ToolExecutor) ListTools() []string {
	var names []string
	for name := range te.tools {
		names = append(names, name)
	}
	return names
}

// ListTools 列出所有工具
func (tr *ToolRegistry) ListTools() []string {
	var names []string
	for name := range tr.tools {
		names = append(names, name)
	}
	return names
}

// ExecuteTool 执行工具
func (te *ToolExecutor) ExecuteTool(ctx context.Context, toolCall *llm.ToolCall) (*ToolResult, error) {
	tool, exists := te.tools[toolCall.Function.Name]
	if !exists {
		return nil, &ToolError{
			Message: fmt.Sprintf("tool '%s' not found", toolCall.Function.Name),
			Code:    404,
		}
	}

	// 转换参数格式
	args := make(ToolCallArguments)
	for key, value := range toolCall.Function.Arguments {
		args[key] = value
	}

	// 验证参数
	if err := tool.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 执行工具
	result, err := tool.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	// 设置调用ID
	result.CallID = toolCall.ID
	result.Name = toolCall.Function.Name

	return result, nil
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (tr *ToolRegistry) Register(tool Tool) {
	tr.tools[tool.GetName()] = tool
}

// Get 获取工具
func (tr *ToolRegistry) Get(name string) (Tool, bool) {
	tool, exists := tr.tools[name]
	return tool, exists
}

// GetAll 获取所有工具
func (tr *ToolRegistry) GetAll() map[string]Tool {
	return tr.tools
}

// GetToolDefinitions 获取工具定义（用于LLM）
func (tr *ToolRegistry) GetToolDefinitions() []llm.Tool {
	var tools []llm.Tool
	for _, tool := range tr.tools {
		// 转换参数格式
		params := make(map[string]interface{})
		for _, param := range tool.GetParameters() {
			paramMap := map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
				"required":    param.Required,
			}
			
			if param.Enum != nil {
				paramMap["enum"] = param.Enum
			}
			
			if param.Items != nil {
				paramMap["items"] = param.Items
			}
			
			params[param.Name] = paramMap
		}

		llmTool := llm.Tool{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        tool.GetName(),
				Description: tool.GetDescription(),
				Parameters:  params,
			},
		}
		tools = append(tools, llmTool)
	}
	return tools
}

// ToolExecutionStats 工具执行统计
type ToolExecutionStats struct {
	ToolName     string        `json:"tool_name"`
	ExecutionTime time.Duration `json:"execution_time"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// ToolExecutionTracker 工具执行跟踪器
type ToolExecutionTracker struct {
	stats []ToolExecutionStats
}

// NewToolExecutionTracker 创建工具执行跟踪器
func NewToolExecutionTracker() *ToolExecutionTracker {
	return &ToolExecutionTracker{
		stats: make([]ToolExecutionStats, 0),
	}
}

// TrackExecution 跟踪工具执行
func (tet *ToolExecutionTracker) TrackExecution(toolName string, startTime time.Time, success bool, err error) {
	duration := time.Since(startTime)
	
	stat := ToolExecutionStats{
		ToolName:     toolName,
		ExecutionTime: duration,
		Success:      success,
		Timestamp:    startTime,
	}
	
	if err != nil {
		stat.Error = err.Error()
	}
	
	tet.stats = append(tet.stats, stat)
}

// GetStats 获取执行统计
func (tet *ToolExecutionTracker) GetStats() []ToolExecutionStats {
	return tet.stats
}

// GetSuccessRate 获取成功率
func (tet *ToolExecutionTracker) GetSuccessRate() float64 {
	if len(tet.stats) == 0 {
		return 0.0
	}
	
	successCount := 0
	for _, stat := range tet.stats {
		if stat.Success {
			successCount++
		}
	}
	
	return float64(successCount) / float64(len(tet.stats))
}

// GetAverageExecutionTime 获取平均执行时间
func (tet *ToolExecutionTracker) GetAverageExecutionTime() time.Duration {
	if len(tet.stats) == 0 {
		return 0
	}
	
	totalTime := time.Duration(0)
	for _, stat := range tet.stats {
		totalTime += stat.ExecutionTime
	}
	
	return totalTime / time.Duration(len(tet.stats))
}
