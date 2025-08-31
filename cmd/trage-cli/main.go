package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"trage-agent-go/pkg/agent"
	"trage-agent-go/pkg/config"
	"trage-agent-go/pkg/tools"

	"github.com/spf13/cobra"
)

var (
	// 命令行参数
	configFile     string
	provider       string
	model          string
	modelBaseURL   string
	apiKey         string
	maxSteps       int
	workingDir     string
	mustPatch      bool
	trajectoryFile string
	patchPath      string
	consoleType    string
	agentType      string
	task           string
	filePath       string
	interactive    bool
)

// 根命令
var rootCmd = &cobra.Command{
	Use:   "trage-cli",
	Short: "Trae Agent - LLM-based agent for software engineering tasks",
	Long: `Trae Agent 是一个基于LLM的代理，专门用于处理软件工程任务。

它提供了强大的CLI接口，可以理解自然语言指令并使用各种工具和LLM提供商执行复杂的软件工程工作流。

主要特性：
- 多LLM支持（OpenAI、Anthropic、Google Gemini、OpenRouter、Ollama等）
- 丰富的工具生态系统（文件编辑、bash执行、结构化思考等）
- 交互式模式，支持迭代开发
- 轨迹记录，详细记录所有代理操作
- 灵活的配置系统，支持YAML配置和环境变量
`,
	Version: "0.1.0",
}

// run命令
var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "执行任务",
	Long:  "执行指定的软件工程任务",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTask,
}

// show-config命令
var showConfigCmd = &cobra.Command{
	Use:   "show-config",
	Short: "显示配置",
	Long:  "显示当前加载的配置信息",
	RunE:  showConfig,
}

// interactive命令
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "交互式模式",
	Long:  "启动交互式模式，支持多轮对话",
	RunE:  startInteractive,
}

func init() {
	// 设置根命令
	rootCmd.AddCommand(runCmd, showConfigCmd, interactiveCmd)

	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "c", "trae_config.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", "LLM提供商")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "特定模型")
	rootCmd.PersistentFlags().StringVar(&modelBaseURL, "model-base-url", "", "模型API的基础URL")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "API密钥")
	rootCmd.PersistentFlags().IntVar(&maxSteps, "max-steps", 0, "最大执行步数")
	rootCmd.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", "", "代理的工作目录")
	rootCmd.PersistentFlags().BoolVarP(&mustPatch, "must-patch", "x", false, "是否必须生成补丁")
	rootCmd.PersistentFlags().StringVarP(&trajectoryFile, "trajectory-file", "t", "", "轨迹文件保存路径")
	rootCmd.PersistentFlags().StringVarP(&patchPath, "patch-path", "j", "", "补丁文件路径")
	rootCmd.PersistentFlags().StringVarP(&consoleType, "console-type", "o", "simple", "控制台类型（simple或rich）")
	rootCmd.PersistentFlags().StringVarP(&agentType, "agent-type", "g", "trae_agent", "代理类型")

	// run命令标志
	runCmd.Flags().StringVarP(&filePath, "file", "f", "", "包含任务描述的文件路径")
	runCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "交互式模式")

	// 绑定环境变量
	rootCmd.PersistentFlags().Lookup("config-file").Value.Set(os.Getenv("TRAE_CONFIG_FILE"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runTask 运行任务
func runTask(cmd *cobra.Command, args []string) error {
	// 获取任务描述
	var taskDescription string
	if len(args) > 0 {
		taskDescription = args[0]
	} else if filePath != "" {
		// 从文件读取任务描述
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read task file: %v", err)
		}
		taskDescription = string(content)
	} else {
		return fmt.Errorf("must provide either a task description or a file path")
	}

	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	// 解析命令行参数覆盖配置
	if err := parseCommandLineOverrides(cfg); err != nil {
		return fmt.Errorf("failed to parse command line overrides: %v", err)
	}

	// 创建代理工厂
	factory := agent.NewAgentFactory()

	// 创建代理
	agentInstance, err := factory.CreateAgent(agent.AgentType(agentType), cfg, trajectoryFile)
	if err != nil {
		return fmt.Errorf("failed to create agent: %v", err)
	}

	// 注册工具
	registerTools(agentInstance)

	// 设置工作目录
	if workingDir != "" {
		if err := os.Chdir(workingDir); err != nil {
			return fmt.Errorf("failed to change working directory: %v", err)
		}
	}

	// 构建额外参数
	extraArgs := buildExtraArgs()

	// 运行代理
	ctx := context.Background()
	execution, err := agentInstance.Run(ctx, taskDescription, extraArgs, nil)
	if err != nil {
		return fmt.Errorf("agent execution failed: %v", err)
	}

	// 输出结果
	if execution.Success {
		fmt.Printf("✅ 任务执行成功！\n")
		fmt.Printf("输出: %s\n", execution.Output)
	} else {
		fmt.Printf("❌ 任务执行失败！\n")
		fmt.Printf("错误: %s\n", execution.Error)
	}

	fmt.Printf("执行时间: %v\n", execution.Duration)
	fmt.Printf("执行步数: %d\n", len(execution.Steps))

	return nil
}

// showConfig 显示配置
func showConfig(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	fmt.Println("cfg", *cfg)
	// 验证配置
	if err := cfg.Validate(); err != nil {
		fmt.Printf("⚠️  配置验证警告: %v\n", err)
	}

	fmt.Println("=== Trae Agent 配置 ===")

	// 显示代理配置
	fmt.Println("\n代理配置:")
	for name, agentCfg := range cfg.Agents {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    模型: %s\n", agentCfg.Model)
		fmt.Printf("    最大步数: %d\n", agentCfg.MaxSteps)
		fmt.Printf("    启用Lakeview: %t\n", agentCfg.EnableLakeview)
		fmt.Printf("    工具: %s\n", strings.Join(agentCfg.Tools, ", "))
	}

	// 显示模型提供商配置
	fmt.Println("\n模型提供商配置:")
	for name, provider := range cfg.ModelProviders {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    提供商: %s\n", provider.Provider)
		if provider.BaseURL != "" {
			fmt.Printf("    基础URL: %s\n", provider.BaseURL)
		}
		if provider.APIVersion != "" {
			fmt.Printf("    API版本: %s\n", provider.APIVersion)
		}
		if provider.APIKey != "" {
			fmt.Printf("    API密钥: %s...\n", provider.APIKey[:min(8, len(provider.APIKey))])
		}
	}

	// 显示模型配置
	fmt.Println("\n模型配置:")
	for name, modelCfg := range cfg.Models {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    模型: %s\n", modelCfg.Model)
		fmt.Printf("    提供商: %s\n", modelCfg.ModelProvider)
		if modelCfg.ResolvedProvider != nil {
			fmt.Printf("    解析的提供商: %s\n", modelCfg.ResolvedProvider.Provider)
		}
		fmt.Printf("    最大令牌数: %d\n", modelCfg.MaxTokens)
		fmt.Printf("    温度: %.2f\n", modelCfg.Temperature)
	}

	// 显示环境变量
	fmt.Println("\n环境变量:")
	envVars := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GOOGLE_API_KEY", "OPENROUTER_API_KEY", "DOUBAO_API_KEY"}
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			fmt.Printf("  %s: %s...\n", envVar, value[:min(8, len(value))])
		}
	}

	return nil
}

// startInteractive 启动交互式模式
func startInteractive(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 启动 Trae Agent 交互式模式")
	fmt.Println("输入 'help' 查看可用命令，输入 'exit' 退出")
	fmt.Println()

	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 创建代理工厂
	factory := agent.NewAgentFactory()

	// 创建代理
	agentInstance, err := factory.CreateAgent(agent.AgentTypeTraeAgent, cfg, trajectoryFile)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// 注册工具
	registerTools(agentInstance)

	// 启动交互式循环
	return runInteractiveLoop(agentInstance, cfg)
}

// runInteractiveLoop 运行交互式循环
func runInteractiveLoop(agentInstance agent.Agent, cfg *config.Config) error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("✅ 交互式模式已启动！")
	fmt.Println("可用命令: help, status, clear, exit/quit")
	fmt.Println()

	for {
		fmt.Print("trae-agent> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "help":
			showHelp()
		case "status":
			showStatus(agentInstance, cfg)
		case "clear":
			clearScreen()
		case "exit", "quit":
			fmt.Println("👋 再见！")
			return nil
		default:
			// 执行任务
			if err := executeTask(agentInstance, input); err != nil {
				fmt.Printf("❌ 任务执行失败: %v\n", err)
			}
		}

		fmt.Println()
	}

	return scanner.Err()
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("📖 可用命令:")
	fmt.Println("• 输入任何任务描述来执行任务")
	fmt.Println("• 'help' - 显示此帮助信息")
	fmt.Println("• 'status' - 显示代理状态")
	fmt.Println("• 'clear' - 清屏")
	fmt.Println("• 'exit' 或 'quit' - 退出会话")
}

// showStatus 显示代理状态
func showStatus(agentInstance agent.Agent, cfg *config.Config) {
	fmt.Println("📊 代理状态:")

	agentConfig := agentInstance.GetConfig()
	if agentConfig != nil {
		fmt.Printf("• 模型: %s\n", agentConfig.Model)
		fmt.Printf("• 最大步数: %d\n", agentConfig.MaxSteps)
		fmt.Printf("• 工具数量: %d\n", len(agentConfig.Tools))
	}

	fmt.Printf("• 配置文件: %s\n", configFile)
	if workingDir, err := os.Getwd(); err == nil {
		fmt.Printf("• 工作目录: %s\n", workingDir)
	}
}

// clearScreen 清屏
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// executeTask 执行任务
func executeTask(agentInstance agent.Agent, task string) error {
	fmt.Printf("🚀 执行任务: %s\n", task)

	ctx := context.Background()
	extraArgs := buildExtraArgs()

	// 执行任务
	execution, err := agentInstance.Run(ctx, task, extraArgs, nil)
	if err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	// 显示结果
	if execution.Success {
		fmt.Printf("✅ 任务执行成功！\n")
		if execution.Output != "" {
			fmt.Printf("输出: %s\n", execution.Output)
		}
	} else {
		fmt.Printf("❌ 任务执行失败！\n")
		if execution.Error != "" {
			fmt.Printf("错误: %s\n", execution.Error)
		}
	}

	fmt.Printf("执行时间: %v\n", execution.Duration)
	fmt.Printf("执行步数: %d\n", len(execution.Steps))

	return nil
}

// parseCommandLineOverrides 解析命令行参数覆盖配置
func parseCommandLineOverrides(cfg *config.Config) error {
	// 如果指定了提供商，更新配置
	if provider != "" {
		// 这里需要实现具体的配置覆盖逻辑
		fmt.Printf("使用提供商: %s\n", provider)
	}

	// 如果指定了模型，更新配置
	if model != "" {
		fmt.Printf("使用模型: %s\n", model)
	}

	// 如果指定了API密钥，更新配置
	if apiKey != "" {
		fmt.Printf("使用API密钥: %s...\n", apiKey[:min(8, len(apiKey))])
	}

	return nil
}

// registerTools 注册工具
func registerTools(agentInstance agent.Agent) {
	// 创建工具实例
	bashTool := tools.NewBashTool()
	editTool := tools.NewEditTool()
	sequentialThinkingTool := tools.NewSequentialThinkingTool()
	taskDoneTool := tools.NewTaskDoneTool()

	// 检查代理类型并注册工具
	switch ag := agentInstance.(type) {
	case *agent.BaseAgent:
		ag.AddTool(bashTool)
		ag.AddTool(editTool)
		ag.AddTool(sequentialThinkingTool)
		ag.AddTool(taskDoneTool)
		fmt.Printf("已注册工具: %s\n", strings.Join(ag.GetToolRegistry().ListTools(), ", "))
	case *agent.TraeAgent:
		ag.AddTool(bashTool)
		ag.AddTool(editTool)
		ag.AddTool(sequentialThinkingTool)
		ag.AddTool(taskDoneTool)
		fmt.Printf("已注册工具: %s\n", strings.Join(ag.GetToolRegistry().ListTools(), ", "))
	default:
		fmt.Printf("警告: 未知的代理类型 %T，无法注册工具\n", agentInstance)
	}
}

// buildExtraArgs 构建额外参数
func buildExtraArgs() map[string]string {
	extraArgs := make(map[string]string)

	if workingDir != "" {
		extraArgs["working_dir"] = workingDir
	}

	if mustPatch {
		extraArgs["must_patch"] = "true"
	}

	if patchPath != "" {
		extraArgs["patch_path"] = patchPath
	}

	return extraArgs
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
