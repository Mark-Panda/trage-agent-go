package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"trage-agent-go/pkg/config"
	"trage-agent-go/pkg/llm"
	"trage-agent-go/pkg/tools"
)

// TraeAgent Trae代理实现
type TraeAgent struct {
	*BaseAgent
	projectPath         string
	baseCommit          string
	mustPatch           string
	patchPath           string
	mcpServersConfig    map[string]config.MCPServerConfig
	allowMCPServers     []string
	mcpTools            []tools.Tool
	mcpClients          []interface{} // MCP客户端列表，用于清理
	cliConsole          Console
	allowMCPServersFlag bool
	conversationHistory []llm.LLMMessage // 对话历史记录
}

// NewTraeAgent 创建TraeAgent
func NewTraeAgent(agentConfig *config.AgentConfig, modelConfig *config.ModelConfig, llmClient llm.LLMClient) *TraeAgent {
	baseAgent := NewBaseAgent(AgentTypeTraeAgent, agentConfig, modelConfig, llmClient)

	return &TraeAgent{
		BaseAgent:           baseAgent,
		projectPath:         "",
		baseCommit:          "",
		mustPatch:           "false",
		patchPath:           "",
		mcpServersConfig:    nil,
		allowMCPServers:     agentConfig.Tools, // 使用配置中的工具列表
		mcpTools:            make([]tools.Tool, 0),
		mcpClients:          make([]interface{}, 0),
		cliConsole:          nil,
		allowMCPServersFlag: true,
	}
}

// SetCLIConsole 设置CLI控制台
func (ta *TraeAgent) SetCLIConsole(console Console) {
	ta.cliConsole = console
}

// SetTrajectoryRecorder 设置轨迹记录器
func (ta *TraeAgent) SetTrajectoryRecorder(recorder llm.TrajectoryRecorder) {
	ta.BaseAgent.SetTrajectoryRecorder(recorder)
}

// NewTask 创建新任务
func (ta *TraeAgent) NewTask(task string, extraArgs map[string]string, toolNames []string) error {
	// 调用父类方法
	if err := ta.BaseAgent.NewTask(task, extraArgs, toolNames); err != nil {
		return err
	}

	// 将新任务添加到对话历史
	ta.AddToConversationHistory(llm.LLMMessage{
		Role:    "user",
		Content: task,
	})

	// TraeAgent特定的任务初始化逻辑
	if extraArgs != nil {
		if projectPath, exists := extraArgs["project_path"]; exists {
			ta.projectPath = projectPath
		}
		if baseCommit, exists := extraArgs["base_commit"]; exists {
			ta.baseCommit = baseCommit
		}
		if mustPatch, exists := extraArgs["must_patch"]; exists {
			ta.mustPatch = mustPatch
		}
		if patchPath, exists := extraArgs["patch_path"]; exists {
			ta.patchPath = patchPath
		}
	}

	// 初始化MCP工具
	if ta.allowMCPServersFlag && ta.mcpServersConfig != nil {
		if err := ta.initializeMCP(); err != nil {
			return fmt.Errorf("failed to initialize MCP: %w", err)
		}
	}

	return nil
}

// AddToConversationHistory 添加消息到对话历史
func (ta *TraeAgent) AddToConversationHistory(message llm.LLMMessage) {
	ta.conversationHistory = append(ta.conversationHistory, message)
}

// GetConversationHistory 获取对话历史
func (ta *TraeAgent) GetConversationHistory() []llm.LLMMessage {
	return ta.conversationHistory
}

// ClearConversationHistory 清空对话历史
func (ta *TraeAgent) ClearConversationHistory() {
	ta.conversationHistory = make([]llm.LLMMessage, 0)
}

// Run 运行代理（重写BaseAgent的实现）
func (ta *TraeAgent) Run(ctx context.Context, task string, extraArgs map[string]string, toolNames []string) (*AgentExecution, error) {
	startTime := time.Now()

	// 创建新任务
	if err := ta.NewTask(task, extraArgs, toolNames); err != nil {
		return nil, err
	}

	// 执行任务 - 调用TraeAgent的实现
	execution, err := ta.ExecuteTask(ctx)
	if err != nil {
		return nil, err
	}

	// 设置执行时间
	execution.Duration = time.Since(startTime)

	return execution, nil
}

// ExecuteTask 执行任务
func (ta *TraeAgent) ExecuteTask(ctx context.Context) (*AgentExecution, error) {
	execution := &AgentExecution{
		Success:     false,
		Steps:       make([]ExecutionStep, 0),
		ToolResults: make([]*tools.ToolResult, 0),
		Metadata:    make(map[string]interface{}),
	}

	// 检查步数限制
	if err := ta.CheckStepLimit(); err != nil {
		execution.Error = err.Error()
		return execution, err
	}

	// 构建系统提示
	systemPrompt := ta.buildSystemPrompt()

	// 创建初始消息，包含系统提示和对话历史
	messages := []llm.LLMMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加对话历史（如果有的话）
	if len(ta.conversationHistory) > 0 {
		messages = append(messages, ta.conversationHistory...)
	}

	// 主执行循环
	for ta.GetStepCount() < ta.GetMaxSteps() {
		// 检查步数限制
		if err := ta.CheckStepLimit(); err != nil {
			execution.Error = err.Error()
			break
		}

		// 添加用户消息（如果是第一步且没有对话历史）
		if ta.GetStepCount() == 0 && len(ta.conversationHistory) == 0 {
			messages = append(messages, llm.LLMMessage{
				Role:    "user",
				Content: ta.BaseAgent.GetTask(), // 使用实际的用户任务输入
			})
		}

		// 调用LLM
		llmConfig := ta.modelConfig.ToLLMModelConfig().(llm.ModelConfig)
		response, err := ta.llmClient.Chat(messages, ta.toolRegistry.GetToolDefinitions(), llmConfig)
		if err != nil {
			// 改进错误处理，提供更详细的错误信息
			errorMsg := fmt.Sprintf("LLM call failed: %v", err)
			if strings.Contains(err.Error(), "404") {
				errorMsg += " - 模型不存在或无访问权限，请检查模型名称和API密钥"
			} else if strings.Contains(err.Error(), "401") {
				errorMsg += " - API密钥无效，请检查API密钥"
			} else if strings.Contains(err.Error(), "rate limit") {
				errorMsg += " - 达到API调用限制，请稍后重试"
			}
			execution.Error = errorMsg
			break
		}

		// 记录消息
		messages = append(messages, *response)

		// 将响应添加到对话历史
		ta.AddToConversationHistory(*response)

		// 检查是否有工具调用
		if len(response.ToolCalls) > 0 {
			// 执行工具调用
			for _, toolCall := range response.ToolCalls {
				startTime := time.Now()

				// 执行工具
				tool, exists := ta.toolRegistry.Get(toolCall.Function.Name)
				if !exists {
					execution.Error = fmt.Sprintf("tool '%s' not found", toolCall.Function.Name)
					break
				}

				// 将工具调用的参数转换为ToolCallArguments
				args := make(tools.ToolCallArguments)
				for key, value := range toolCall.Function.Arguments {
					args[key] = value
				}

				toolResult, err := tool.Execute(ctx, args)

				// 添加调试信息
				fmt.Printf("工具执行: %s, 参数: %v, 错误: %v\n", toolCall.Function.Name, args, err)
				if toolResult != nil {
					fmt.Printf("工具结果: 成功=%v, 结果=%s\n", toolResult.Success, toolResult.Result)
				}

				// 跟踪执行
				ta.TrackToolExecution(toolCall.Function.Name, startTime, err == nil, err)

				if err != nil {
					// 记录工具执行错误
					toolResult = &tools.ToolResult{
						CallID:  toolCall.ID,
						Name:    toolCall.Function.Name,
						Success: false,
						Error:   err.Error(),
					}
				}

				// 添加执行步骤
				ta.AddExecutionStep(
					"tool_execution",
					fmt.Sprintf("Tool: %s, Args: %v", toolCall.Function.Name, toolCall.Function.Arguments),
					toolResult.Result,
					&toolCall,
					toolResult,
				)

				// 添加工具结果到执行历史
				execution.ToolResults = append(execution.ToolResults, toolResult)

				// 将工具结果添加到消息历史
				toolMessage := llm.LLMMessage{
					Role:       "tool",
					Content:    toolResult.Result,
					ToolCallID: toolCall.ID,
				}
				messages = append(messages, toolMessage)

				// 将工具结果添加到对话历史
				ta.AddToConversationHistory(toolMessage)
			}

			// 工具执行完成后，检查是否应该停止
			if ta.shouldStopExecution(execution) {
				execution.Success = true
				execution.Output = "任务通过工具执行完成"
				break
			}
		} else {
			// 没有工具调用，检查是否是最终答案
			if ta.isTaskComplete(response.Content) {
				execution.Success = true
				execution.Output = response.Content
				break
			}
		}

		// 添加执行步骤
		ta.AddExecutionStep(
			"llm_response",
			"",
			response.Content,
			nil,
			nil,
		)
	}

	// 设置执行统计
	execution.Steps = ta.getExecutionSteps()

	return execution, nil
}

// buildSystemPrompt 构建系统提示
func (ta *TraeAgent) buildSystemPrompt() string {
	prompt := `你是一个专业的软件工程代理，专门用于处理软件工程任务。

你的能力包括：
- 代码分析和理解
- 代码编辑和重构
- 执行命令行操作
- 结构化思考
- 任务完成判断

可用工具：
`

	// 添加工具描述
	for _, tool := range ta.tools {
		prompt += fmt.Sprintf("- %s: %s\n", tool.GetName(), tool.GetDescription())
	}

	prompt += `
请按照以下步骤工作：
1. 分析任务需求
2. 使用适当的工具执行任务
3. 验证结果
4. 报告完成状态

重要提示：
- 当用户要求创建文件时，必须使用edit_file工具实际创建文件
- 当用户要求执行命令时，必须使用bash工具实际执行
- 不要只提供代码示例，要实际完成任务
- 始终使用工具来完成任务，不要假设或猜测。`

	return prompt
}

// getCurrentTask 获取当前任务
func (ta *TraeAgent) getCurrentTask() string {
	if ta.BaseAgent != nil && ta.BaseAgent.config != nil {
		// 从配置中获取任务描述，如果没有则返回默认值
		if ta.BaseAgent.config.Model != "" {
			return fmt.Sprintf("使用模型 %s 执行软件工程任务", ta.BaseAgent.config.Model)
		}
	}
	return "执行软件工程任务"
}

// isTaskComplete 检查任务是否完成
func (ta *TraeAgent) isTaskComplete(content string) bool {
	// 简单的任务完成检测逻辑
	lowerContent := strings.ToLower(content)

	// 检查是否包含完成相关的关键词
	completionKeywords := []string{
		"任务完成", "完成", "done", "finished", "complete",
		"任务已结束", "执行完毕", "over", "success",
	}

	for _, keyword := range completionKeywords {
		if strings.Contains(lowerContent, keyword) {
			return true
		}
	}

	return false
}

// shouldStopExecution 检查是否应该停止执行
func (ta *TraeAgent) shouldStopExecution(execution *AgentExecution) bool {
	// 检查是否有成功的工具执行结果
	successfulTools := 0
	for _, result := range execution.ToolResults {
		if result.Success {
			successfulTools++
		}
	}

	// 如果工具执行成功且没有错误，考虑停止
	if successfulTools > 0 && execution.Error == "" {
		// 检查最近的工具结果是否表明任务完成
		if len(execution.ToolResults) > 0 {
			lastResult := execution.ToolResults[len(execution.ToolResults)-1]
			if lastResult.Success && ta.isTaskComplete(lastResult.Result) {
				return true
			}
		}

		// 检查是否达到了合理的执行步数
		if ta.GetStepCount() >= 3 { // 允许至少3步来完成任务
			return true
		}
	}

	return false
}

// getExecutionSteps 获取执行步骤
func (ta *TraeAgent) getExecutionSteps() []ExecutionStep {
	if ta.BaseAgent != nil {
		return ta.BaseAgent.getExecutionSteps()
	}
	return make([]ExecutionStep, 0)
}

// initializeMCP 初始化MCP
func (ta *TraeAgent) initializeMCP() error {
	// MCP初始化逻辑
	// 这里需要实现具体的MCP客户端创建和工具发现
	return nil
}

// cleanupMCPClients 清理MCP客户端
func (ta *TraeAgent) cleanupMCPClients() error {
	// MCP客户端清理逻辑
	return nil
}

// Console 控制台接口
type Console interface {
	// Print 打印消息
	Print(message string)

	// PrintTaskDetails 打印任务详情
	PrintTaskDetails(details map[string]string)

	// Start 启动控制台
	Start() error

	// SetLakeview 设置Lakeview
	SetLakeview(config *config.LakeviewConfig)
}

// ModelConfig 模型配置接口实现
type modelConfigWrapper struct {
	config *config.ModelConfig
}

func (mcw *modelConfigWrapper) GetModel() string {
	return mcw.config.Model
}

func (mcw *modelConfigWrapper) GetMaxTokens() int {
	return mcw.config.MaxTokens
}

func (mcw *modelConfigWrapper) GetTemperature() float64 {
	return mcw.config.Temperature
}

func (mcw *modelConfigWrapper) GetTopP() float64 {
	return mcw.config.TopP
}

func (mcw *modelConfigWrapper) GetTopK() int {
	return mcw.config.TopK
}

func (mcw *modelConfigWrapper) GetParallelToolCalls() bool {
	return mcw.config.ParallelToolCalls
}

func (mcw *modelConfigWrapper) GetMaxRetries() int {
	return mcw.config.MaxRetries
}

func (mcw *modelConfigWrapper) GetSupportsToolCalling() bool {
	return mcw.config.SupportsToolCalling
}

func (mcw *modelConfigWrapper) GetAPIKey() string {
	if mcw.config.ResolvedProvider != nil {
		return mcw.config.ResolvedProvider.APIKey
	}
	return ""
}

func (mcw *modelConfigWrapper) GetBaseURL() string {
	if mcw.config.ResolvedProvider != nil {
		return mcw.config.ResolvedProvider.BaseURL
	}
	return ""
}

func (mcw *modelConfigWrapper) GetAPIVersion() string {
	if mcw.config.ResolvedProvider != nil {
		return mcw.config.ResolvedProvider.APIVersion
	}
	return ""
}
