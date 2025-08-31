package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	configContent := `
agents:
  trae_agent:
    enable_lakeview: true
    model: test_model
    max_steps: 100
    tools:
      - bash
      - edit_file

model_providers:
  test_provider:
    api_key: "test_key"
    provider: "test_provider"
    base_url: "https://test.com"
    api_version: "v1"

models:
  test_model:
    model_provider: test_provider
    model: "test-model"
    max_tokens: 2048
    temperature: 0.7
    top_p: 0.9
    top_k: 1
    parallel_tool_calls: true
    max_retries: 2
    supports_tool_calling: true
`

	// 写入临时文件
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	// 加载配置
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证代理配置
	if len(config.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(config.Agents))
	}

	agent, exists := config.Agents["trae_agent"]
	if !exists {
		t.Fatal("Expected trae_agent to exist")
	}

	if !agent.EnableLakeview {
		t.Error("Expected EnableLakeview to be true")
	}

	if agent.Model != "test_model" {
		t.Errorf("Expected model 'test_model', got '%s'", agent.Model)
	}

	if agent.MaxSteps != 100 {
		t.Errorf("Expected max_steps 100, got %d", agent.MaxSteps)
	}

	if len(agent.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(agent.Tools))
	}

	// 验证模型提供商配置
	if len(config.ModelProviders) != 1 {
		t.Errorf("Expected 1 model provider, got %d", len(config.ModelProviders))
	}

	provider, exists := config.ModelProviders["test_provider"]
	if !exists {
		t.Fatal("Expected test_provider to exist")
	}

	if provider.APIKey != "test_key" {
		t.Errorf("Expected API key 'test_key', got '%s'", provider.APIKey)
	}

	if provider.Provider != "test_provider" {
		t.Errorf("Expected provider 'test_provider', got '%s'", provider.Provider)
	}

	// 验证模型配置
	if len(config.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(config.Models))
	}

	model, exists := config.Models["test_model"]
	if !exists {
		t.Fatal("Expected test_model to exist")
	}

	if model.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", model.Model)
	}

	if model.ModelProvider != "test_provider" {
		t.Errorf("Expected model provider 'test_provider', got '%s'", model.ModelProvider)
	}

	if model.MaxTokens != 2048 {
		t.Errorf("Expected max_tokens 2048, got %d", model.MaxTokens)
	}

	if model.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", model.Temperature)
	}
}

func TestConfig_Validate(t *testing.T) {
	config := &Config{
		Agents: map[string]AgentConfig{
			"test_agent": {
				Model: "test_model",
			},
		},
		ModelProviders: map[string]ModelProvider{
			"test_provider": {
				APIKey:   "test_key",
				Provider: "test_provider",
			},
		},
		Models: map[string]ModelConfig{
			"test_model": {
				Model:          "test-model",
				ModelProvider:  "test_provider",
				MaxTokens:      2048,
				Temperature:    0.7,
				TopP:          0.9,
				TopK:          1,
				ParallelToolCalls: true,
				MaxRetries:    2,
				SupportsToolCalling: true,
			},
		},
	}

	// 验证有效配置
	if err := config.Validate(); err != nil {
		t.Errorf("Expected no validation error, got %v", err)
	}

	// 测试缺少代理
	config.Agents = nil
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for missing agents")
	}

	// 恢复配置
	config.Agents = map[string]AgentConfig{
		"test_agent": {
			Model: "test_model",
		},
	}

	// 测试缺少模型提供商
	config.ModelProviders = nil
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for missing model providers")
	}

	// 恢复配置
	config.ModelProviders = map[string]ModelProvider{
		"test_provider": {
			APIKey:   "test_key",
			Provider: "test_provider",
		},
	}

	// 测试缺少模型
	config.Models = nil
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for missing models")
	}
}

func TestConfig_GetTraeAgentConfig(t *testing.T) {
	config := &Config{
		Agents: map[string]AgentConfig{
			"trae_agent": {
				Model: "test_model",
			},
		},
	}

	agentConfig, err := config.GetTraeAgentConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if agentConfig.Model != "test_model" {
		t.Errorf("Expected model 'test_model', got '%s'", agentConfig.Model)
	}

	// 测试缺少trae_agent
	config.Agents = map[string]AgentConfig{}
	_, err = config.GetTraeAgentConfig()
	if err == nil {
		t.Error("Expected error for missing trae_agent")
	}
}

func TestConfig_GetModelConfig(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"test_model": {
				Model: "test-model",
			},
		},
	}

	modelConfig, err := config.GetModelConfig("test_model")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if modelConfig.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", modelConfig.Model)
	}

	// 测试不存在的模型
	_, err = config.GetModelConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent model")
	}
}
