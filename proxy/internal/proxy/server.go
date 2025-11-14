package proxy

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/ethan/claude-proxy/internal/config"
	"github.com/ethan/claude-proxy/internal/logger"
)

// Server 代理服务器
type Server struct {
	router  *gin.Engine
	handler *Handler
	srv     *http.Server
}

// NewServer 创建代理服务器
func NewServer() *Server {
	// 设置Gin模式
	if config.Cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	handler := NewHandler()
	
	return &Server{
		router:  router,
		handler: handler,
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 添加日志中间件
	s.router.Use(s.loggerMiddleware())
	
	// 添加恢复中间件
	s.router.Use(gin.Recovery())
	
	// 添加CORS中间件
	s.router.Use(s.corsMiddleware())
	
	// 添加代理中间件
	s.router.Use(s.handler.ProxyMiddleware())
	
	// 健康检查端点
	s.router.GET("/health", s.healthCheck)
	
	// 状态端点
	s.router.GET("/status", s.statusCheck)
}

// loggerMiddleware 自定义日志中间件
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Cfg.Features.RequestLogging {
			c.Next()
			return
		}
		
		// 开始时间
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 计算延迟
		latency := time.Since(start)
		
		// 记录请求日志
		logger.LogInfo("HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// corsMiddleware CORS中间件
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// statusCheck 状态检查
func (s *Server) statusCheck(c *gin.Context) {
	// 收集服务状态
	services := make([]gin.H, 0, len(config.Cfg.Services))
	for _, svc := range config.Cfg.Services {
		services = append(services, gin.H{
			"id":   svc.ID,
			"name": svc.Name,
			"role": svc.Role,
			"url":  svc.URL,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
		"config": gin.H{
			"proxy_port":          config.Cfg.Proxy.Port,
			"evaluator_fallback":  config.Cfg.Features.EvaluatorFallback,
			"service_auto_switch": config.Cfg.Features.ServiceAutoSwitch,
			"request_logging":     config.Cfg.Features.RequestLogging,
		},
		"services":           services,
		"difficulty_mapping": config.Cfg.DifficultyMapping,
		"time":              time.Now().Format(time.RFC3339),
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	// 设置路由
	s.setupRoutes()
	
	// 创建HTTP服务器，使用配置的超时值
	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Cfg.Proxy.Port),
		Handler:      s.router,
		ReadTimeout:  time.Duration(config.Cfg.Proxy.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Cfg.Proxy.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.Cfg.Proxy.IdleTimeout) * time.Second,
	}

	// 记录超时配置
	logger.LogInfo("服务器超时配置",
		"read_timeout", config.Cfg.Proxy.ReadTimeout,
		"write_timeout", config.Cfg.Proxy.WriteTimeout,
		"idle_timeout", config.Cfg.Proxy.IdleTimeout,
		"request_timeout", config.Cfg.Proxy.RequestTimeout,
	)
	
	// 启动服务器
	go func() {
		logger.LogInfo("代理服务器启动", "port", config.Cfg.Proxy.Port)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError("服务器启动失败", err)
			os.Exit(1)
		}
	}()
	
	// 等待中断信号
	s.waitForShutdown()
	
	return nil
}

// waitForShutdown 等待关闭信号
func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.LogInfo("正在关闭服务器...")
	
	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 优雅关闭服务器
	if err := s.srv.Shutdown(ctx); err != nil {
		logger.LogError("服务器关闭失败", err)
		os.Exit(1)
	}
	
	logger.LogInfo("服务器已关闭")
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if s.srv == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	return s.srv.Shutdown(ctx)
}
