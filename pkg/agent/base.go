package agent

import (
	"context"
	"fmt"
	"time"

	"trage-agent-go/pkg/config"
	"trage-agent-go/pkg/llm"
	"trage-agent-go/pkg/tools"
)

// AgentType 代理类型
type AgentType string

const (
	AgentTypeTraeAgent AgentType = "trae_agent"
)

// AgentError 代理错误
type AgentError struct {
	Message string
	Code    int
}

func (e AgentError) Error() string {
	return e.Message
}

// AgentExecution 代理执行结果
type AgentExecution struct {
	Success     bool                   `json:"success"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Steps       []ExecutionStep        `json:"steps"`
	Duration    time.Duration          `json:"duration"`
	ToolResults []*tools.ToolResult    `json:"tool_results,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionStep 执行步骤
type ExecutionStep struct {
	StepNumber int                    `json:"step_number"`
	Action     string                 `json:"action"`
	Input      string                 `json:"input,omitempty"`
	Output     string                 `json:"output,omitempty"`
	ToolCall   *llm.ToolCall          `json:"tool_call,omitempty"`
	ToolResult *tools.ToolResult      `json:"tool_result,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Agent 代理接口
type Agent interface {
	// Run 运行代理
	Run(ctx context.Context, task string, extraArgs map[string]string, toolNames []string) (*AgentExecution, error)

	// NewTask 创建新任务
	NewTask(task string, extraArgs map[string]string, toolNames []string) error

	// ExecuteTask 执行任务
	ExecuteTask(ctx context.Context) (*AgentExecution, error)

	// SetTrajectoryRecorder 设置轨迹记录器
	SetTrajectoryRecorder(recorder llm.TrajectoryRecorder)

	// GetConfig 获取配置
	GetConfig() *config.AgentConfig
}

// BaseAgent 基础代理实现
type BaseAgent struct {
	agentType          AgentType
	config             *config.AgentConfig
	modelConfig        *config.ModelConfig
	llmClient          llm.LLMClient
	tools              []tools.Tool
	toolRegistry       *tools.ToolRegistry
	trajectoryRecorder llm.TrajectoryRecorder
	executionTracker   *tools.ToolExecutionTracker
	stepCount          int
	maxSteps           int
}

// NewBaseAgent 创建基础代理
func NewBaseAgent(agentType AgentType, config *config.AgentConfig, modelConfig *config.ModelConfig, llmClient llm.LLMClient) *BaseAgent {
	return &BaseAgent{
		agentType:        agentType,
		config:           config,
		modelConfig:      modelConfig,
		llmClient:        llmClient,
		tools:            make([]tools.Tool, 0),
		toolRegistry:     tools.NewToolRegistry(),
		executionTracker: tools.NewToolExecutionTracker(),
		maxSteps:         config.MaxSteps,
	}
}

// GetConfig 获取配置
func (ba *BaseAgent) GetConfig() *config.AgentConfig {
	return ba.config
}

// SetTrajectoryRecorder 设置轨迹记录器
func (ba *BaseAgent) SetTrajectoryRecorder(recorder llm.TrajectoryRecorder) {
	ba.trajectoryRecorder = recorder
	if ba.llmClient != nil {
		ba.llmClient.SetTrajectoryRecorder(recorder)
	}
}

// AddTool 添加工具
func (ba *BaseAgent) AddTool(tool tools.Tool) {
	ba.tools = append(ba.tools, tool)
	ba.toolRegistry.Register(tool)
}

// GetTools 获取工具列表
func (ba *BaseAgent) GetTools() []tools.Tool {
	return ba.tools
}

// GetToolRegistry 获取工具注册表
func (ba *BaseAgent) GetToolRegistry() *tools.ToolRegistry {
	return ba.toolRegistry
}

// NewTask 创建新任务
func (ba *BaseAgent) NewTask(task string, extraArgs map[string]string, toolNames []string) error {
	// 重置步数计数
	ba.stepCount = 0

	// 如果指定了工具名称，过滤工具
	if len(toolNames) > 0 {
		filteredTools := make([]tools.Tool, 0)
		for _, tool := range ba.tools {
			for _, name := range toolNames {
				if tool.GetName() == name {
					filteredTools = append(filteredTools, tool)
					break
				}
			}
		}
		ba.tools = filteredTools
	}

	return nil
}

// ExecuteTask 执行任务（基础实现，子类需要重写）
func (ba *BaseAgent) ExecuteTask(ctx context.Context) (*AgentExecution, error) {
	return nil, &AgentError{
		Message: "ExecuteTask must be implemented by concrete agent",
		Code:    501,
	}
}

// Run 运行代理（基础实现）
func (ba *BaseAgent) Run(ctx context.Context, task string, extraArgs map[string]string, toolNames []string) (*AgentExecution, error) {
	startTime := time.Now()

	// 创建新任务
	if err := ba.NewTask(task, extraArgs, toolNames); err != nil {
		return nil, err
	}

	// 执行任务
	execution, err := ba.ExecuteTask(ctx)
	if err != nil {
		return nil, err
	}

	// 设置执行时间
	execution.Duration = time.Since(startTime)

	return execution, nil
}

// AddExecutionStep 添加执行步骤
func (ba *BaseAgent) AddExecutionStep(action, input, output string, toolCall *llm.ToolCall, toolResult *tools.ToolResult) {
	ba.stepCount++

	// 这里可以添加步骤到执行历史中
	// 具体实现取决于子类的需求
	_ = ExecutionStep{
		StepNumber: ba.stepCount,
		Action:     action,
		Input:      input,
		Output:     output,
		ToolCall:   toolCall,
		ToolResult: toolResult,
		Timestamp:  time.Now(),
	}
}

// CheckStepLimit 检查步数限制
func (ba *BaseAgent) CheckStepLimit() error {
	if ba.stepCount >= ba.maxSteps {
		return &AgentError{
			Message: fmt.Sprintf("maximum number of steps (%d) exceeded", ba.maxSteps),
			Code:    429,
		}
	}
	return nil
}

// GetStepCount 获取当前步数
func (ba *BaseAgent) GetStepCount() int {
	return ba.stepCount
}

// GetMaxSteps 获取最大步数
func (ba *BaseAgent) GetMaxSteps() int {
	return ba.maxSteps
}

// GetExecutionTracker 获取执行跟踪器
func (ba *BaseAgent) GetExecutionTracker() *tools.ToolExecutionTracker {
	return ba.executionTracker
}

// TrackToolExecution 跟踪工具执行
func (ba *BaseAgent) TrackToolExecution(toolName string, startTime time.Time, success bool, err error) {
	ba.executionTracker.TrackExecution(toolName, startTime, success, err)
}

// AgentFactory 代理工厂
type AgentFactory struct{}

// NewAgentFactory 创建代理工厂
func NewAgentFactory() *AgentFactory {
	return &AgentFactory{}
}

// CreateAgent 创建代理
func (af *AgentFactory) CreateAgent(
	agentType AgentType,
	config *config.Config,
	trajectoryFile string,
) (Agent, error) {
	// 获取代理配置
	agentConfig, err := config.GetTraeAgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent config: %w", err)
	}

	// 获取模型配置
	modelConfig, err := config.GetModelConfig(agentConfig.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	// 创建LLM客户端
	llmClient, err := af.createLLMClient(modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// 根据代理类型创建具体代理
	switch agentType {
	case AgentTypeTraeAgent:
		return NewTraeAgent(agentConfig, modelConfig, llmClient), nil
	default:
		return nil, &AgentError{
			Message: fmt.Sprintf("unsupported agent type: %s", agentType),
			Code:    400,
		}
	}
}

// createLLMClient 创建LLM客户端
func (af *AgentFactory) createLLMClient(modelConfig *config.ModelConfig) (llm.LLMClient, error) {
	provider := modelConfig.ModelProvider

	switch provider {
	case "openai":
		return af.createOpenAIClient(modelConfig)
	case "doubao":
		return af.createDoubaoClient(modelConfig)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s, only 'openai' and 'doubao' are supported", provider)
	}
}

// createOpenAIClient 创建OpenAI客户端
func (af *AgentFactory) createOpenAIClient(modelConfig *config.ModelConfig) (llm.LLMClient, error) {
	provider := modelConfig.ResolvedProvider
	if provider == nil {
		return nil, fmt.Errorf("provider not resolved for model %s", modelConfig.Model)
	}
	return llm.NewOpenAIClient(
		provider.APIKey,
		provider.BaseURL,
		provider.APIVersion,
	), nil
}

// createDoubaoClient 创建Doubao客户端
func (af *AgentFactory) createDoubaoClient(modelConfig *config.ModelConfig) (llm.LLMClient, error) {
	provider := modelConfig.ResolvedProvider
	if provider == nil {
		return nil, fmt.Errorf("provider not resolved for model %s", modelConfig.Model)
	}
	return llm.NewDoubaoClient(
		provider.APIKey,
		provider.BaseURL,
		provider.APIVersion,
	), nil
}
