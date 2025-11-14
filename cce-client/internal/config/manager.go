package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProxyConfig 代理配置（简化版，与 proxy/internal/models/config.go 对应）
type ProxyConfig struct {
	Port             int `yaml:"port"`
	ReadTimeout      int `yaml:"read_timeout"`
	WriteTimeout     int `yaml:"write_timeout"`
	IdleTimeout      int `yaml:"idle_timeout"`
	RequestTimeout   int `yaml:"request_timeout"`
	EvaluatorTimeout int `yaml:"evaluator_timeout"`
}

// Service 服务配置
type Service struct {
	ID               string `yaml:"id"`
	Name             string `yaml:"name"`
	URL              string `yaml:"url"`
	APIKey           string `yaml:"api_key"`
	Role             string `yaml:"role"` // "evaluator" or "executor"
	SupportsThinking *bool  `yaml:"supports_thinking,omitempty"`
}

// EvaluatorConfig 评估器配置
type EvaluatorConfig struct {
	Model            string `yaml:"model"`
	MaxTokens        int    `yaml:"max_tokens"`
	IncludeHistory   bool   `yaml:"include_history"`
	MaxHistoryRounds int    `yaml:"max_history_rounds"`
	PromptTemplate   string `yaml:"prompt_template"`
}

// Features 功能开关
type Features struct {
	EvaluatorFallback bool `yaml:"evaluator_fallback"`
	ServiceAutoSwitch bool `yaml:"service_auto_switch"`
	RequestLogging    bool `yaml:"request_logging"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	OutputPath string `yaml:"output_path"`
}

// Config 完整配置
type Config struct {
	Proxy             ProxyConfig            `yaml:"proxy"`
	Services          []Service              `yaml:"services"`
	DifficultyMapping map[string]string      `yaml:"difficulty_mapping"`
	Evaluator         EvaluatorConfig        `yaml:"evaluator"`
	Features          Features               `yaml:"features"`
	Logging           LoggingConfig          `yaml:"logging"`
}

// Manager 配置管理器
type Manager struct {
	configPath string
	config     *Config
}

// NewManager 创建配置管理器
func NewManager() (*Manager, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %w", err)
	}

	manager := &Manager{
		configPath: configPath,
	}

	// 加载配置（如果文件不存在，创建默认配置）
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		manager.config = getDefaultConfig()
		if err := manager.Save(); err != nil {
			return nil, fmt.Errorf("保存默认配置失败: %w", err)
		}
	} else {
		if err := manager.Load(); err != nil {
			return nil, fmt.Errorf("加载配置失败: %w", err)
		}
	}

	return manager, nil
}

// Load 加载配置
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	m.config = &Config{}
	if err := yaml.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// Save 保存配置
func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetConfig 获取配置对象
func (m *Manager) GetConfig() *Config {
	return m.config
}

// SetConfig 设置配置对象
func (m *Manager) SetConfig(config *Config) {
	m.config = config
}

// GetConfigPath 获取配置文件路径
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// GetProxyPort 获取代理端口
func (m *Manager) GetProxyPort() int {
	return m.config.Proxy.Port
}

// GetLogsPath 获取日志目录路径
func (m *Manager) GetLogsPath() string {
	if m.config.Logging.OutputPath != "" {
		return m.config.Logging.OutputPath
	}
	// 默认日志路径
	configDir := filepath.Dir(m.configPath)
	return filepath.Join(configDir, "logs")
}

// Validate 验证配置
func (m *Manager) Validate() error {
	// 检查服务列表
	if len(m.config.Services) == 0 {
		return fmt.Errorf("至少需要配置一个服务")
	}

	// 检查是否有 evaluator
	hasEvaluator := false
	hasExecutor := false
	serviceIDs := make(map[string]bool)

	for _, svc := range m.config.Services {
		if svc.ID == "" {
			return fmt.Errorf("服务 ID 不能为空")
		}
		if serviceIDs[svc.ID] {
			return fmt.Errorf("服务 ID 重复: %s", svc.ID)
		}
		serviceIDs[svc.ID] = true

		if svc.Role == "evaluator" {
			hasEvaluator = true
		} else if svc.Role == "executor" {
			hasExecutor = true
		}
	}

	if !hasEvaluator {
		return fmt.Errorf("至少需要一个 evaluator 服务")
	}
	if !hasExecutor {
		return fmt.Errorf("至少需要一个 executor 服务")
	}

	// 检查难度映射
	for level := 1; level <= 5; level++ {
		levelStr := fmt.Sprintf("%d", level)
		serviceID, exists := m.config.DifficultyMapping[levelStr]
		if !exists {
			return fmt.Errorf("难度级别 %d 未配置", level)
		}
		if !serviceIDs[serviceID] {
			return fmt.Errorf("难度级别 %d 映射的服务 %s 不存在", level, serviceID)
		}
	}

	return nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, "Library", "Application Support", "CCE")
	return filepath.Join(configDir, "config.yaml"), nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	trueVal := true
	return &Config{
		Proxy: ProxyConfig{
			Port:             27015,
			ReadTimeout:      1800,
			WriteTimeout:     1800,
			IdleTimeout:      300,
			RequestTimeout:   1800,
			EvaluatorTimeout: 30,
		},
		Services: []Service{
			{
				ID:               "evaluator-main",
				Name:             "主决策者服务",
				URL:              "https://api.anthropic.com/v1/messages",
				APIKey:           "sk-your-api-key-here",
				Role:             "evaluator",
				SupportsThinking: &trueVal,
			},
			{
				ID:               "haiku-service",
				Name:             "Haiku 服务",
				URL:              "https://api.anthropic.com/v1/messages",
				APIKey:           "sk-your-api-key-here",
				Role:             "executor",
				SupportsThinking: &trueVal,
			},
		},
		DifficultyMapping: map[string]string{
			"1": "haiku-service",
			"2": "haiku-service",
			"3": "haiku-service",
			"4": "haiku-service",
			"5": "haiku-service",
		},
		Evaluator: EvaluatorConfig{
			Model:            "claude-3-haiku-20240307",
			MaxTokens:        100,
			IncludeHistory:   true,
			MaxHistoryRounds: 3,
			PromptTemplate:   "你是任务复杂度评估专家...",
		},
		Features: Features{
			EvaluatorFallback: false,
			ServiceAutoSwitch: false,
			RequestLogging:    true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			OutputPath: "./logs",
		},
	}
}
