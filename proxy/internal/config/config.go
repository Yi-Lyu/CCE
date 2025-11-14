package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/ethan/claude-proxy/internal/models"
	"github.com/spf13/viper"
)

var (
	// Cfg 全局配置实例
	Cfg *models.Config
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	if configPath == "" {
		// 默认查找配置文件路径
		configPath = "./configs/config.yaml"
	}
	
	// 获取配置文件的目录和文件名
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	
	viper.AddConfigPath(dir)
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	
	// 设置环境变量前缀
	viper.SetEnvPrefix("CLAUDE_PROXY")
	viper.AutomaticEnv()
	
	// 设置默认值
	setDefaults()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，创建默认配置
			return createDefaultConfig(configPath)
		}
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	// 解析配置到结构体
	Cfg = &models.Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	// 为服务设置默认值
	// SupportsThinking 默认为true（支持thinking模式）
	for i := range Cfg.Services {
		// 如果配置中未显式设置，则默认支持thinking
		// 由于viper.Unmarshal后bool的零值是false，我们需要检查是否被显式设置
		// 这里采用简单方法：如果是evaluator或官方API，默认支持
		if Cfg.Services[i].Role == "evaluator" {
			// Evaluator服务默认支持thinking
			if !viper.IsSet(fmt.Sprintf("services.%d.supports_thinking", i)) {
				Cfg.Services[i].SupportsThinking = true
			}
		} else if Cfg.Services[i].Role == "executor" {
			// Executor服务需要显式配置，默认支持thinking以保持兼容性
			if !viper.IsSet(fmt.Sprintf("services.%d.supports_thinking", i)) {
				Cfg.Services[i].SupportsThinking = true
			}
		}
	}

	// 验证配置
	if err := validateConfig(Cfg); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	return nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 代理配置
	viper.SetDefault("proxy.port", 27015)
	viper.SetDefault("proxy.read_timeout", 1800)      // 30分钟
	viper.SetDefault("proxy.write_timeout", 1800)     // 30分钟
	viper.SetDefault("proxy.idle_timeout", 300)       // 5分钟
	viper.SetDefault("proxy.request_timeout", 1800)   // 30分钟
	viper.SetDefault("proxy.evaluator_timeout", 30)   // 30秒

	// 功能开关
	viper.SetDefault("features.evaluator_fallback", false)
	viper.SetDefault("features.service_auto_switch", false)
	viper.SetDefault("features.request_logging", true)

	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.output_path", "./logs")

	// 决策者默认配置
	viper.SetDefault("evaluator.include_history", true)
	viper.SetDefault("evaluator.max_history_rounds", 3)
	viper.SetDefault("evaluator.model", "claude-3-haiku-20240307")
	viper.SetDefault("evaluator.max_tokens", 100)

	// 决策者默认Prompt模板
	defaultPrompt := `你是一个任务复杂度评估专家。请分析以下 Claude API 请求中【当前这一步具体任务】的复杂度，并返回 JSON 格式的结果。

重要说明：
- 请评估【当前这一步操作】的复杂度，而非整体项目的复杂度
- 例如：如果整体任务是"开发复杂电商系统"，但当前步骤是"创建一个配置文件"，应评估为简单任务（1-2级）
- 请聚焦于当前需要执行的具体操作，不要被项目整体规模影响判断

当前任务信息：
- 模型: {{.Model}}
- 消息数量: {{.MessageCount}}
- 当前任务: {{.CurrentTask}}{{.HistoryContext}}

评估标准：
1 级（非常简单）：简单查询、基础问答、信息查找、单行代码、创建简单文件
2 级（简单）：基础分析、简单总结、格式转换、简单函数编写、修改配置
3 级（中等）：代码编写、数据分析、文档生成、模块开发、多文件修改
4 级（复杂）：架构设计、复杂重构、深度分析、多模块集成、算法实现
5 级（非常复杂）：系统设计、多步骤规划任务、创新性解决方案、大型重构

请严格按照以下 JSON 格式返回，不要包含任何其他内容：
{
  "difficulty_level": 数字（必须是1-5之间的整数）
}

你的评估：`
	viper.SetDefault("evaluator.prompt_template", defaultPrompt)
}

// validateConfig 验证配置的有效性
func validateConfig(cfg *models.Config) error {
	// 检查是否有服务配置
	if len(cfg.Services) == 0 {
		return fmt.Errorf("至少需要配置一个服务")
	}
	
	// 检查是否有决策者服务
	hasEvaluator := false
	serviceIDs := make(map[string]bool)
	for _, svc := range cfg.Services {
		if svc.ID == "" {
			return fmt.Errorf("服务ID不能为空")
		}
		if serviceIDs[svc.ID] {
			return fmt.Errorf("服务ID重复: %s", svc.ID)
		}
		serviceIDs[svc.ID] = true
		
		if svc.URL == "" {
			return fmt.Errorf("服务 %s 的URL不能为空", svc.ID)
		}
		if svc.APIKey == "" {
			return fmt.Errorf("服务 %s 的API Key不能为空", svc.ID)
		}
		
		if svc.Role == "evaluator" {
			hasEvaluator = true
		}
	}
	
	if !hasEvaluator {
		return fmt.Errorf("至少需要配置一个决策者服务 (role=evaluator)")
	}
	
	// 检查难度映射
	if len(cfg.DifficultyMapping) == 0 {
		return fmt.Errorf("难度映射配置不能为空")
	}
	
	// 检查难度映射中的服务ID是否存在
	for level, serviceID := range cfg.DifficultyMapping {
		if !serviceIDs[serviceID] {
			return fmt.Errorf("难度等级 %s 映射的服务ID %s 不存在", level, serviceID)
		}
	}
	
	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	defaultConfig := `# Claude 智能代理服务配置

proxy:
  port: 27015

# 服务列表
services:
  - id: "evaluator-1"
    name: "决策者服务"
    url: "https://api.example.com/v1/messages"
    api_key: "cr_your_evaluator_api_key"
    role: "evaluator"

  - id: "simple-service"
    name: "简单任务服务"
    url: "https://simple.example.com/v1/messages"
    api_key: "cr_your_simple_api_key"
    role: "executor"
    supports_thinking: true   # 官方API支持thinking（默认值）

  - id: "medium-service"
    name: "中等任务服务"
    url: "https://medium.example.com/v1/messages"
    api_key: "cr_your_medium_api_key"
    role: "executor"
    supports_thinking: true

  - id: "complex-service"
    name: "复杂任务服务"
    url: "https://complex.example.com/v1/messages"
    api_key: "cr_your_complex_api_key"
    role: "executor"
    supports_thinking: true

# 难度等级映射 (1-5)
difficulty_mapping:
  "1": "simple-service"
  "2": "simple-service"
  "3": "medium-service"
  "4": "complex-service"
  "5": "complex-service"

# 决策者配置
evaluator:
  model: "claude-3-haiku-20240307"
  max_tokens: 100
  include_history: true
  max_history_rounds: 3
  prompt_template: |
    你是一个任务复杂度评估专家。请分析以下 Claude API 请求中【当前这一步具体任务】的复杂度，并返回 JSON 格式的结果。

    重要说明：
    - 请评估【当前这一步操作】的复杂度，而非整体项目的复杂度
    - 请聚焦于当前需要执行的具体操作，不要被项目整体规模影响判断

    当前任务信息：
    - 模型: {{.Model}}
    - 消息数量: {{.MessageCount}}
    - 当前任务: {{.CurrentTask}}{{.HistoryContext}}

    评估标准：
    1 级（非常简单）：简单查询、基础问答、单行代码、创建简单文件
    2 级（简单）：基础分析、简单总结、格式转换、修改配置
    3 级（中等）：代码编写、数据分析、模块开发、多文件修改
    4 级（复杂）：架构设计、复杂重构、多模块集成、算法实现
    5 级（非常复杂）：系统设计、多步骤规划任务、大型重构

    请严格按照以下 JSON 格式返回：
    {
      "difficulty_level": 数字（必须是1-5之间的整数）
    }

    你的评估：

# 功能开关
features:
  evaluator_fallback: false  # 决策者服务不可用时使用备选
  service_auto_switch: false # 目标服务不可用时自动切换
  request_logging: true      # 记录请求日志

# 日志配置
logging:
  level: "info"
  output_path: "./logs"
`
	
	// 创建配置目录
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}
	
	// 写入默认配置
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("创建默认配置文件失败: %v", err)
	}
	
	return fmt.Errorf("已创建默认配置文件: %s，请修改后重新运行", configPath)
}

// GetServiceByID 根据ID获取服务配置
func GetServiceByID(id string) (*models.Service, error) {
	if Cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}
	
	for _, svc := range Cfg.Services {
		if svc.ID == id {
			return &svc, nil
		}
	}
	
	return nil, fmt.Errorf("服务ID不存在: %s", id)
}

// GetEvaluatorService 获取决策者服务
func GetEvaluatorService() (*models.Service, error) {
	if Cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}
	
	for _, svc := range Cfg.Services {
		if svc.Role == "evaluator" {
			return &svc, nil
		}
	}
	
	return nil, fmt.Errorf("未配置决策者服务")
}

// GetAllExecutorServices 获取所有执行者服务
// 返回所有 role="executor" 的服务列表，用于广播式 Warmup 预热
func GetAllExecutorServices() ([]*models.Service, error) {
	if Cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	var executors []*models.Service
	for i := range Cfg.Services {
		if Cfg.Services[i].Role == "executor" {
			executors = append(executors, &Cfg.Services[i])
		}
	}

	if len(executors) == 0 {
		return nil, fmt.Errorf("未配置执行者服务")
	}

	return executors, nil
}
