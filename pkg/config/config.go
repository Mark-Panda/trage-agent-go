package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigError 配置错误
type ConfigError struct {
	Message string
}

func (e ConfigError) Error() string {
	return e.Message
}

// ModelProvider 模型提供商配置
type ModelProvider struct {
	APIKey     string `yaml:"api_key" json:"api_key"`
	Provider   string `yaml:"provider" json:"provider"`
	BaseURL    string `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	APIVersion string `yaml:"api_version,omitempty" json:"api_version,omitempty"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	Model               string   `yaml:"model" json:"model"`
	ModelProvider       string   `yaml:"model_provider" json:"model_provider"`
	MaxTokens           int      `yaml:"max_tokens" json:"max_tokens"`
	Temperature         float64  `yaml:"temperature" json:"temperature"`
	TopP                float64  `yaml:"top_p" json:"top_p"`
	TopK                int      `yaml:"top_k" json:"top_k"`
	ParallelToolCalls   bool     `yaml:"parallel_tool_calls" json:"parallel_tool_calls"`
	MaxRetries          int      `yaml:"max_retries" json:"max_retries"`
	SupportsToolCalling bool     `yaml:"supports_tool_calling" json:"supports_tool_calling"`
	CandidateCount      *int     `yaml:"candidate_count,omitempty" json:"candidate_count,omitempty"`
	StopSequences       []string `yaml:"stop_sequences,omitempty" json:"stop_sequences,omitempty"`

	// 解析后的提供商信息
	ResolvedProvider *ModelProvider `yaml:"-" json:"-"`
}

// ResolveConfigValues 解析配置值，支持命令行和环境变量覆盖
func (m *ModelConfig) ResolveConfigValues(
	modelProviders map[string]ModelProvider,
	provider *string,
	model *string,
	modelBaseURL *string,
	apiKey *string,
) error {
	if model != nil {
		m.Model = *model
	}

	if provider != nil {
		m.ModelProvider = *provider
	}

	// 解析提供商信息
	if m.ModelProvider != "" {
		if mp, exists := modelProviders[m.ModelProvider]; exists {
			m.ResolvedProvider = &mp
		} else {
			// 创建新的提供商配置
			newProvider := ModelProvider{
				Provider: m.ModelProvider,
			}
			if apiKey != nil {
				newProvider.APIKey = *apiKey
			}
			if modelBaseURL != nil {
				newProvider.BaseURL = *modelBaseURL
			}
			m.ResolvedProvider = &newProvider
		}
	}

	// 从环境变量解析API密钥和基础URL
	if m.ResolvedProvider != nil {
		envVarAPIKey := strings.ToUpper(m.ResolvedProvider.Provider) + "_API_KEY"
		envVarBaseURL := strings.ToUpper(m.ResolvedProvider.Provider) + "_BASE_URL"

		if resolvedAPIKey := resolveConfigValue(apiKey, m.ResolvedProvider.APIKey, envVarAPIKey); resolvedAPIKey != "" {
			m.ResolvedProvider.APIKey = resolvedAPIKey
		}

		if resolvedBaseURL := resolveConfigValue(modelBaseURL, m.ResolvedProvider.BaseURL, envVarBaseURL); resolvedBaseURL != "" {
			m.ResolvedProvider.BaseURL = resolvedBaseURL
		}
	}

	return nil
}

// AgentConfig 代理配置
type AgentConfig struct {
	EnableLakeview bool     `yaml:"enable_lakeview" json:"enable_lakeview"`
	Model          string   `yaml:"model" json:"model"`
	MaxSteps       int      `yaml:"max_steps" json:"max_steps"`
	Tools          []string `yaml:"tools" json:"tools"`
}

// LakeviewConfig Lakeview配置
type LakeviewConfig struct {
	MaxLines int `yaml:"max_lines" json:"max_lines"`
}

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args" json:"args"`
}

// Config 主配置结构
type Config struct {
	Agents          map[string]AgentConfig     `yaml:"agents" json:"agents"`
	ModelProviders  map[string]ModelProvider   `yaml:"model_providers" json:"model_providers"`
	Models          map[string]ModelConfig     `yaml:"models" json:"models"`
	Lakeview        LakeviewConfig             `yaml:"lakeview" json:"lakeview"`
	MCPServers      map[string]MCPServerConfig `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	AllowMCPServers []string                   `yaml:"allow_mcp_servers,omitempty" json:"allow_mcp_servers,omitempty"`
}

// LoadConfig 加载配置文件
func LoadConfig(configFile string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 使用YAML库解析配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	fmt.Println("config", config)

	// 解析模型配置中的提供商信息
	if err := config.resolveModelProviders(); err != nil {
		return nil, fmt.Errorf("failed to resolve model providers: %w", err)
	}

	return &config, nil
}

// GetTraeAgentConfig 获取TraeAgent配置
func (c *Config) GetTraeAgentConfig() (*AgentConfig, error) {
	if agentConfig, exists := c.Agents["trae_agent"]; exists {
		return &agentConfig, nil
	}
	return nil, &ConfigError{Message: "trae_agent configuration not found"}
}

// GetModelConfig 获取模型配置
func (c *Config) GetModelConfig(modelName string) (*ModelConfig, error) {
	if modelConfig, exists := c.Models[modelName]; exists {
		return &modelConfig, nil
	}
	return nil, &ConfigError{Message: fmt.Sprintf("model configuration '%s' not found", modelName)}
}

// GetModelProvider 获取模型提供商配置
func (c *Config) GetModelProvider(providerName string) (*ModelProvider, error) {
	if providerConfig, exists := c.ModelProviders[providerName]; exists {
		return &providerConfig, nil
	}
	return nil, &ConfigError{Message: fmt.Sprintf("model provider '%s' not found", providerName)}
}

// resolveConfigValue 解析配置值，优先级：命令行 > 环境变量 > 配置文件
func resolveConfigValue(cliValue *string, configValue string, envVar string) string {
	if cliValue != nil && *cliValue != "" {
		return *cliValue
	}

	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}

	return configValue
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证必需的配置项
	if len(c.Agents) == 0 {
		return &ConfigError{Message: "at least one agent must be configured"}
	}

	if len(c.ModelProviders) == 0 {
		return &ConfigError{Message: "at least one model provider must be configured"}
	}

	if len(c.Models) == 0 {
		return &ConfigError{Message: "at least one model must be configured"}
	}

	// 验证每个代理配置
	for agentName, agentConfig := range c.Agents {
		if agentConfig.Model == "" {
			return &ConfigError{Message: fmt.Sprintf("agent '%s' must specify a model", agentName)}
		}

		if _, exists := c.Models[agentConfig.Model]; !exists {
			return &ConfigError{Message: fmt.Sprintf("agent '%s' references undefined model '%s'", agentName, agentConfig.Model)}
		}
	}

	// 验证每个模型配置
	for modelName, modelConfig := range c.Models {
		if modelConfig.Model == "" {
			return &ConfigError{Message: fmt.Sprintf("model '%s' must specify a model name", modelName)}
		}

		if modelConfig.ModelProvider == "" {
			return &ConfigError{Message: fmt.Sprintf("model '%s' must specify a provider", modelName)}
		}

		if _, exists := c.ModelProviders[modelConfig.ModelProvider]; !exists {
			return &ConfigError{Message: fmt.Sprintf("model '%s' references undefined provider '%s'", modelName, modelConfig.ModelProvider)}
		}

		// 检查API密钥
		provider := c.ModelProviders[modelConfig.ModelProvider]
		if provider.APIKey == "" {
			// 检查环境变量
			envVar := strings.ToUpper(provider.Provider) + "_API_KEY"
			if os.Getenv(envVar) == "" {
				return &ConfigError{Message: fmt.Sprintf("model '%s' has no API key configured and no environment variable '%s' found", modelName, envVar)}
			}
		}
	}

	return nil
}

// GetEnv 获取环境变量值
func (c *Config) GetEnv(key string) string {
	return os.Getenv(key)
}

// GetEnvWithDefault 获取环境变量值，如果不存在则返回默认值
func (c *Config) GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取环境变量整数值
func (c *Config) GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool 获取环境变量布尔值
func (c *Config) GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// resolveModelProviders 解析模型配置中的提供商信息
func (c *Config) resolveModelProviders() error {
	for modelName, modelConfig := range c.Models {
		if modelConfig.ModelProvider != "" {
			if provider, exists := c.ModelProviders[modelConfig.ModelProvider]; exists {
				// 获取结构体，修改它，然后重新赋值给map
				updatedConfig := modelConfig
				updatedConfig.ResolvedProvider = &provider
				c.Models[modelName] = updatedConfig
			} else {
				return &ConfigError{Message: fmt.Sprintf("model '%s' references undefined provider '%s'", modelName, modelConfig.ModelProvider)}
			}
		}
	}
	return nil
}
