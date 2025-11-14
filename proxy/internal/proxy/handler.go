package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ethan/claude-proxy/internal/config"
	"github.com/ethan/claude-proxy/internal/evaluator"
	"github.com/ethan/claude-proxy/internal/logger"
	"github.com/ethan/claude-proxy/internal/models"
)

// Handler 代理处理器
type Handler struct {
	evaluatorClient *evaluator.Client
}

// NewHandler 创建代理处理器
func NewHandler() *Handler {
	return &Handler{
		evaluatorClient: evaluator.NewClient(),
	}
}

// ProxyMiddleware 代理中间件
func (h *Handler) ProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// 检查是否是需要代理的请求
		if !h.shouldProxy(c.Request) {
			c.Next()
			return
		}
		
		// 处理代理请求
		if err := h.handleProxyRequest(c, startTime); err != nil {
			logger.LogError("代理请求失败", err,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "代理请求失败",
				"details": err.Error(),
			})
		}
	}
}

// shouldProxy 判断是否需要代理
func (h *Handler) shouldProxy(req *http.Request) bool {
	// 代理所有 Claude API 请求
	// 检查路径是否匹配 Claude API 端点
	path := req.URL.Path
	
	// 支持的 Claude API 路径
	claudePaths := []string{
		"/api/v1/messages",
		"/v1/messages",
		"/anthropic/v1/messages",
		"/api/anthropic/v1/messages",
	}
	
	for _, apiPath := range claudePaths {
		if strings.HasPrefix(path, apiPath) {
			return true
		}
	}
	
	return false
}

// handleProxyRequest 处理代理请求
func (h *Handler) handleProxyRequest(c *gin.Context, startTime time.Time) error {
	// 读取请求体
	var requestBody []byte
	if c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return fmt.Errorf("读取请求体失败: %v", err)
		}
		requestBody = body
		// 恢复请求体以便后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	
	// 解析Claude请求
	var claudeReq models.ClaudeRequest
	if err := json.Unmarshal(requestBody, &claudeReq); err != nil {
		return fmt.Errorf("解析请求体失败: %v", err)
	}

	// 提取用户信息
	userID, sessionID := models.ExtractUserInfo(claudeReq.Metadata)

	// 检测是否为 Warmup 请求
	if models.IsWarmupRequest(&claudeReq) {
		logger.LogInfo("检测到 Warmup 请求，执行广播式预热",
			"user_id", userID,
			"session_id", sessionID,
		)
		// 调用专门的 Warmup 处理函数
		return h.handleWarmupRequest(c, &claudeReq, requestBody, userID, sessionID, startTime)
	}
	
	// 调用决策者服务评估难度
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Cfg.Proxy.EvaluatorTimeout)*time.Second)
	defer cancel()
	
	evalResponse, err := h.evaluatorClient.EvaluateDifficulty(ctx, &claudeReq)
	if err != nil {
		logger.LogError("决策者服务评估失败", err,
			"user_id", userID,
			"session_id", sessionID,
		)
		return fmt.Errorf("决策者服务评估失败: %v", err)
	}
	
	// 记录决策结果
	if config.Cfg.Features.RequestLogging {
		logger.LogEvaluatorRequest(userID, sessionID, evalResponse.DifficultyLevel, evalResponse.Reasoning, time.Since(startTime))
	}
	
	// 根据难度等级获取目标服务
	targetServiceID, ok := config.Cfg.DifficultyMapping[fmt.Sprintf("%d", evalResponse.DifficultyLevel)]
	if !ok {
		return fmt.Errorf("未配置难度等级 %d 的服务映射", evalResponse.DifficultyLevel)
	}
	
	targetService, err := config.GetServiceByID(targetServiceID)
	if err != nil {
		return fmt.Errorf("获取目标服务失败: %v", err)
	}
	
	// 转发请求到目标服务
	if claudeReq.Stream {
		// 处理流式响应
		return h.handleStreamingProxy(c, targetService, requestBody, userID, sessionID, startTime)
	} else {
		// 处理普通响应
		return h.handleNormalProxy(c, targetService, requestBody, userID, sessionID, startTime)
	}
}

// handleWarmupRequest 处理 Warmup 预热请求（广播式）
// 将 Warmup 请求并发发送到所有 executor 服务，确保所有服务都完成预热
func (h *Handler) handleWarmupRequest(c *gin.Context, claudeReq *models.ClaudeRequest, requestBody []byte, userID, sessionID string, startTime time.Time) error {
	// 获取所有 executor 服务
	executors, err := config.GetAllExecutorServices()
	if err != nil {
		logger.LogError("获取执行者服务列表失败", err)
		return fmt.Errorf("获取执行者服务列表失败: %v", err)
	}

	logger.LogInfo("开始广播式预热",
		"service_count", len(executors),
		"user_id", userID,
		"session_id", sessionID,
	)

	// 创建一个用于接收第一个成功响应的 channel
	type warmupResult struct {
		service  *models.Service
		response *http.Response
		err      error
	}

	resultChan := make(chan warmupResult, len(executors))
	var wg sync.WaitGroup

	// 并发向所有服务发送 Warmup 请求
	for _, service := range executors {
		wg.Add(1)
		go func(svc *models.Service) {
			defer wg.Done()

			// 创建目标请求
			req, err := h.createTargetRequest(c.Request, svc, requestBody)
			if err != nil {
				logger.LogError("创建 Warmup 请求失败", err, "service", svc.Name)
				resultChan <- warmupResult{service: svc, err: err}
				return
			}

			// 发送请求（设置10秒超时）
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				logger.LogWarn("Warmup 请求失败", "service", svc.Name, "error", err)
				resultChan <- warmupResult{service: svc, err: err}
				return
			}

			logger.LogInfo("Warmup 请求成功", "service", svc.Name, "status", resp.StatusCode)
			resultChan <- warmupResult{service: svc, response: resp, err: nil}
		}(service)
	}

	// 等待所有请求完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var firstSuccessResponse *http.Response
	var firstSuccessService *models.Service
	successCount := 0
	failCount := 0

	for result := range resultChan {
		if result.err == nil && result.response != nil {
			successCount++
			// 保存第一个成功的响应用于返回给客户端
			if firstSuccessResponse == nil {
				firstSuccessResponse = result.response
				firstSuccessService = result.service
			} else {
				// 关闭其他成功的响应体
				result.response.Body.Close()
			}
		} else {
			failCount++
		}
	}

	// 记录统计信息
	logger.LogInfo("Warmup 预热完成",
		"total", len(executors),
		"success", successCount,
		"failed", failCount,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	// 如果没有任何服务成功，返回错误
	if firstSuccessResponse == nil {
		return fmt.Errorf("所有服务的 Warmup 请求都失败了")
	}

	defer firstSuccessResponse.Body.Close()

	// 根据请求类型返回响应
	if claudeReq.Stream {
		// 流式响应：复制响应头并流式传输
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		// 复制其他响应头
		for key, values := range firstSuccessResponse.Header {
			if key != "Content-Type" && key != "Cache-Control" && key != "Connection" {
				for _, value := range values {
					c.Header(key, value)
				}
			}
		}

		c.Status(firstSuccessResponse.StatusCode)

		// 流式传输
		w := c.Writer
		flusher, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("响应写入器不支持Flush")
		}

		scanner := bufio.NewScanner(firstSuccessResponse.Body)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(w, "%s\n", line)
			if line == "" {
				flusher.Flush()
			}
		}

		if err := scanner.Err(); err != nil {
			logger.LogError("读取 Warmup 流式响应失败", err, "service", firstSuccessService.Name)
			return fmt.Errorf("读取流式响应失败: %v", err)
		}

		flusher.Flush()
	} else {
		// 普通响应：复制响应头和响应体
		for key, values := range firstSuccessResponse.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(firstSuccessResponse.StatusCode)

		_, err = io.Copy(c.Writer, firstSuccessResponse.Body)
		if err != nil {
			return fmt.Errorf("复制响应体失败: %v", err)
		}
	}

	// 记录请求日志
	if config.Cfg.Features.RequestLogging {
		logger.LogRequest(userID, sessionID, "WARMUP", c.Request.URL.Path, string(requestBody), firstSuccessResponse.StatusCode, time.Since(startTime))
	}

	return nil
}

// handleNormalProxy 处理普通响应的代理
func (h *Handler) handleNormalProxy(c *gin.Context, service *models.Service, requestBody []byte, userID, sessionID string, startTime time.Time) error {
	// 创建目标请求
	req, err := h.createTargetRequest(c.Request, service, requestBody)
	if err != nil {
		return err
	}
	
	// 发送请求
	client := &http.Client{Timeout: time.Duration(config.Cfg.Proxy.RequestTimeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// 如果启用了服务自动切换，这里可以实现切换逻辑
		if config.Cfg.Features.ServiceAutoSwitch {
			logger.LogWarn("目标服务不可用，尝试切换", "service_id", service.ID, "error", err)
			// TODO: 实现服务切换逻辑
		}
		return fmt.Errorf("请求目标服务失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	
	// 设置状态码
	c.Status(resp.StatusCode)
	
	// 复制响应体
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return fmt.Errorf("复制响应体失败: %v", err)
	}
	
	// 记录请求日志
	if config.Cfg.Features.RequestLogging {
		logger.LogRequest(userID, sessionID, c.Request.Method, c.Request.URL.Path, string(requestBody), resp.StatusCode, time.Since(startTime))
	}
	
	return nil
}

// handleStreamingProxy 处理流式响应的代理
func (h *Handler) handleStreamingProxy(c *gin.Context, service *models.Service, requestBody []byte, userID, sessionID string, startTime time.Time) error {
	// 创建目标请求
	req, err := h.createTargetRequest(c.Request, service, requestBody)
	if err != nil {
		return err
	}
	
	// 发送请求
	client := &http.Client{Timeout: time.Duration(config.Cfg.Proxy.RequestTimeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if config.Cfg.Features.ServiceAutoSwitch {
			logger.LogWarn("目标服务不可用，尝试切换", "service_id", service.ID, "error", err)
			// TODO: 实现服务切换逻辑
		}
		return fmt.Errorf("请求目标服务失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	
	// 复制其他响应头
	for key, values := range resp.Header {
		if key != "Content-Type" && key != "Cache-Control" && key != "Connection" {
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}
	
	// 设置状态码
	c.Status(resp.StatusCode)
	
	// 创建带缓冲的写入器
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("响应写入器不支持Flush")
	}
	
	// 实时转发流式数据
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		
		// 写入数据
		fmt.Fprintf(w, "%s\n", line)
		
		// 如果是空行，表示一个事件结束，立即刷新
		if line == "" {
			flusher.Flush()
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流式响应失败: %v", err)
	}
	
	// 最后刷新
	flusher.Flush()
	
	// 记录请求日志
	if config.Cfg.Features.RequestLogging {
		logger.LogRequest(userID, sessionID, c.Request.Method, c.Request.URL.Path, string(requestBody), resp.StatusCode, time.Since(startTime))
	}
	
	return nil
}

// sanitizeRequestForExecutor 清理请求体，移除executor可能不支持的字段
// 主要移除thinking相关字段，因为第三方Claude兼容API可能不支持
func (h *Handler) sanitizeRequestForExecutor(body []byte) ([]byte, error) {
	// 解析为map以便修改
	var reqMap map[string]interface{}
	if err := json.Unmarshal(body, &reqMap); err != nil {
		return nil, fmt.Errorf("解析请求体失败: %v", err)
	}

	// 移除thinking字段（Claude Opus的特性，第三方API可能不支持）
	delete(reqMap, "thinking")

	// 重新序列化
	sanitizedBody, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("序列化清理后的请求体失败: %v", err)
	}

	return sanitizedBody, nil
}

// createTargetRequest 创建目标服务的请求
func (h *Handler) createTargetRequest(originalReq *http.Request, service *models.Service, body []byte) (*http.Request, error) {
	// 解析服务URL
	targetURL, err := url.Parse(service.URL)
	if err != nil {
		return nil, fmt.Errorf("解析目标服务URL失败: %v", err)
	}

	// 保持原始请求的查询参数
	targetURL.RawQuery = originalReq.URL.RawQuery

	// 根据服务配置决定是否清理请求体
	// 如果服务不支持thinking模式，则移除相关字段
	sanitizedBody := body
	if !service.SupportsThinking {
		cleaned, err := h.sanitizeRequestForExecutor(body)
		if err != nil {
			logger.LogWarn("清理请求体失败，使用原始请求体", "error", err, "service", service.Name)
		} else {
			sanitizedBody = cleaned
			logger.LogDebug("已清理请求体，移除了thinking字段", "service", service.Name)
		}
	}

	// 创建新请求
	req, err := http.NewRequest(originalReq.Method, targetURL.String(), bytes.NewReader(sanitizedBody))
	if err != nil {
		return nil, fmt.Errorf("创建目标请求失败: %v", err)
	}

	// 复制原始请求头
	for key, values := range originalReq.Header {
		// 跳过Host和Authorization头
		if key == "Host" || key == "Authorization" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 设置正确的Host头
	req.Host = targetURL.Host

	// 设置目标服务的认证头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", service.APIKey))

	return req, nil
}
