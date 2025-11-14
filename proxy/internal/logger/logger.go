package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/ethan/claude-proxy/internal/models"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// SugarLogger 语法糖日志实例
	SugarLogger *zap.SugaredLogger
)

// InitLogger 初始化日志
func InitLogger(cfg *models.LogConfig) error {
	// 创建日志目录
	if err := os.MkdirAll(cfg.OutputPath, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	// 设置日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("解析日志级别失败: %v", err)
	}
	
	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	
	// 创建核心
	cores := []zapcore.Core{
		// 控制台输出
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		),
	}
	
	// 文件输出
	logFile := filepath.Join(cfg.OutputPath, fmt.Sprintf("claude-proxy-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			level,
		))
	}
	
	// 创建日志器
	core := zapcore.NewTee(cores...)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	SugarLogger = Logger.Sugar()
	
	return nil
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// LogRequest 记录请求信息
func LogRequest(userID, sessionID, method, path string, requestBody interface{}, statusCode int, duration time.Duration) {
	if SugarLogger == nil {
		return
	}
	
	SugarLogger.Infow("API Request",
		"user_id", userID,
		"session_id", sessionID,
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
		"request_body", requestBody,
	)
}

// LogEvaluatorRequest 记录决策请求
func LogEvaluatorRequest(userID, sessionID string, difficultyLevel int, reasoning string, duration time.Duration) {
	if SugarLogger == nil {
		return
	}
	
	SugarLogger.Infow("Evaluator Decision",
		"user_id", userID,
		"session_id", sessionID,
		"difficulty_level", difficultyLevel,
		"reasoning", reasoning,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogError 记录错误
func LogError(message string, err error, fields ...interface{}) {
	if SugarLogger == nil {
		return
	}
	
	allFields := append([]interface{}{"error", err}, fields...)
	SugarLogger.Errorw(message, allFields...)
}

// LogInfo 记录信息
func LogInfo(message string, fields ...interface{}) {
	if SugarLogger == nil {
		return
	}
	
	SugarLogger.Infow(message, fields...)
}

// LogWarn 记录警告
func LogWarn(message string, fields ...interface{}) {
	if SugarLogger == nil {
		return
	}

	SugarLogger.Warnw(message, fields...)
}

// LogDebug 记录调试信息
func LogDebug(message string, fields ...interface{}) {
	if SugarLogger == nil {
		return
	}

	SugarLogger.Debugw(message, fields...)
}

// Close 关闭日志
func Close() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
