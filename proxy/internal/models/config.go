package models

// Config 代理服务的配置结构
type Config struct {
	// 代理监听配置
	Proxy ProxyConfig `json:"proxy" mapstructure:"proxy"`

	// 服务列表
	Services []Service `json:"services" mapstructure:"services"`

	// 难度等级映射
	DifficultyMapping map[string]string `json:"difficulty_mapping" mapstructure:"difficulty_mapping"`

	// 决策者配置
	Evaluator EvaluatorConfig `json:"evaluator" mapstructure:"evaluator"`

	// 功能开关
	Features FeatureFlags `json:"features" mapstructure:"features"`

	// 日志配置
	Logging LogConfig `json:"logging" mapstructure:"logging"`
}

// ProxyConfig 代理服务配置
type ProxyConfig struct {
	Port int `json:"port" mapstructure:"port" default:"27015"`

	// 超时配置（单位：秒）
	ReadTimeout       int `json:"read_timeout" mapstructure:"read_timeout" default:"1800"`           // 读取超时，默认30分钟
	WriteTimeout      int `json:"write_timeout" mapstructure:"write_timeout" default:"1800"`         // 写入超时，默认30分钟
	IdleTimeout       int `json:"idle_timeout" mapstructure:"idle_timeout" default:"300"`            // 空闲超时，默认5分钟
	RequestTimeout    int `json:"request_timeout" mapstructure:"request_timeout" default:"1800"`     // 转发请求超时，默认30分钟
	EvaluatorTimeout  int `json:"evaluator_timeout" mapstructure:"evaluator_timeout" default:"30"`   // 评估器超时，默认30秒
}

// Service 服务配置
type Service struct {
	ID              string `json:"id" mapstructure:"id"`
	Name            string `json:"name" mapstructure:"name"`
	URL             string `json:"url" mapstructure:"url"`                             // 包含域名/IP和路径
	APIKey          string `json:"api_key" mapstructure:"api_key"`                     // Bearer token
	Role            string `json:"role" mapstructure:"role"`                           // "evaluator" 或 "executor"
	SupportsThinking bool   `json:"supports_thinking" mapstructure:"supports_thinking"` // 是否支持thinking模式（默认true）
}

// EvaluatorConfig 决策者配置
type EvaluatorConfig struct {
	// Prompt模板，支持变量替换
	PromptTemplate string `json:"prompt_template" mapstructure:"prompt_template"`

	// 是否包含历史上下文
	IncludeHistory bool `json:"include_history" mapstructure:"include_history" default:"true"`

	// 历史上下文的最大轮数
	MaxHistoryRounds int `json:"max_history_rounds" mapstructure:"max_history_rounds" default:"3"`

	// 评估模型（使用什么模型进行评估）
	Model string `json:"model" mapstructure:"model" default:"claude-3-haiku-20240307"`

	// 最大Token数
	MaxTokens int `json:"max_tokens" mapstructure:"max_tokens" default:"100"`
}

// FeatureFlags 功能开关
type FeatureFlags struct {
	// 决策者服务不可用时使用备选服务
	EvaluatorFallback bool `json:"evaluator_fallback" mapstructure:"evaluator_fallback" default:"false"`
	
	// 目标服务不可用时自动切换
	ServiceAutoSwitch bool `json:"service_auto_switch" mapstructure:"service_auto_switch" default:"false"`
	
	// 记录请求日志
	RequestLogging bool `json:"request_logging" mapstructure:"request_logging" default:"true"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level" mapstructure:"level" default:"info"`
	OutputPath string `json:"output_path" mapstructure:"output_path" default:"./logs"`
}
