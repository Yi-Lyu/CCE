package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"github.com/ethan/claude-proxy/internal/config"
	"github.com/ethan/claude-proxy/internal/logger"
	"github.com/ethan/claude-proxy/internal/models"
)

// ContextManager 管理用户上下文
type ContextManager struct {
	mu       sync.RWMutex
	contexts map[string]*models.UserContext // key: userID_sessionID
	maxHistory int
}

// NewContextManager 创建上下文管理器
func NewContextManager() *ContextManager {
	return &ContextManager{
		contexts: make(map[string]*models.UserContext),
		maxHistory: 10, // 保留最近10条请求历史
	}
}

// GetContext 获取用户上下文
func (cm *ContextManager) GetContext(userID, sessionID string) *models.UserContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	key := fmt.Sprintf("%s_%s", userID, sessionID)
	ctx, exists := cm.contexts[key]
	if !exists {
		return &models.UserContext{
			UserID:    userID,
			SessionID: sessionID,
			RequestHistory: []models.RequestSummary{},
		}
	}
	
	return ctx
}

// UpdateContext 更新用户上下文
func (cm *ContextManager) UpdateContext(userID, sessionID string, summary models.RequestSummary) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	key := fmt.Sprintf("%s_%s", userID, sessionID)
	ctx, exists := cm.contexts[key]
	if !exists {
		ctx = &models.UserContext{
			UserID:    userID,
			SessionID: sessionID,
			RequestHistory: []models.RequestSummary{},
		}
		cm.contexts[key] = ctx
	}
	
	// 添加新的请求历史
	ctx.RequestHistory = append(ctx.RequestHistory, summary)
	
	// 保持历史记录在限制范围内
	if len(ctx.RequestHistory) > cm.maxHistory {
		ctx.RequestHistory = ctx.RequestHistory[len(ctx.RequestHistory)-cm.maxHistory:]
	}
}

// Client 决策者服务客户端
type Client struct {
	httpClient     *http.Client
	contextManager *ContextManager
	maxRetries     int
}

// NewClient 创建决策者客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		contextManager: NewContextManager(),
		maxRetries:     3, // 默认重试3次
	}
}

// EvaluateDifficulty 评估请求难度
func (c *Client) EvaluateDifficulty(ctx context.Context, request *models.ClaudeRequest) (*models.EvaluatorResponse, error) {
	// 提取用户信息
	userID, sessionID := models.ExtractUserInfo(request.Metadata)
	
	// 获取用户上下文
	userContext := c.contextManager.GetContext(userID, sessionID)
	
	// 构建评估请求
	evalReq := &models.EvaluatorRequest{
		OriginalRequest: *request,
		UserContext:     *userContext,
	}
	
	// 获取决策者服务配置
	evaluatorService, err := config.GetEvaluatorService()
	if err != nil {
		return nil, fmt.Errorf("获取决策者服务失败: %v", err)
	}
	
	// 执行请求（带重试）
	var response *models.EvaluatorResponse
	var lastErr error
	
	for i := 0; i < c.maxRetries; i++ {
		response, lastErr = c.doRequest(ctx, evaluatorService, evalReq)
		if lastErr == nil {
			break
		}
		
		logger.LogWarn("决策者服务请求失败，准备重试",
			"attempt", i+1,
			"max_retries", c.maxRetries,
			"error", lastErr,
		)
		
		// 指数退避
		if i < c.maxRetries-1 {
			backoff := time.Duration(1<<uint(i)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	
	if lastErr != nil {
		// 如果启用了备选服务，这里可以实现备选逻辑
		if config.Cfg.Features.EvaluatorFallback {
			logger.LogWarn("决策者服务不可用，使用默认难度等级", "default_level", 3)
			return &models.EvaluatorResponse{
				DifficultyLevel: 3,
				Reasoning:       "决策者服务不可用，使用默认中等难度",
			}, nil
		}
		
		return nil, fmt.Errorf("决策者服务请求失败: %v", lastErr)
	}
	
	// 更新用户上下文
	c.contextManager.UpdateContext(userID, sessionID, models.RequestSummary{
		Timestamp:       time.Now(),
		Model:           request.Model,
		MessageCount:    len(request.Messages),
		DifficultyLevel: response.DifficultyLevel,
	})
	
	return response, nil
}

// doRequest 执行单次请求
func (c *Client) doRequest(ctx context.Context, service *models.Service, evalReq *models.EvaluatorRequest) (*models.EvaluatorResponse, error) {
	// 构建评估 prompt
	prompt := c.buildEvaluationPrompt(evalReq)
	
	// 从配置中获取evaluator设置
	evalModel := config.Cfg.Evaluator.Model
	if evalModel == "" {
		evalModel = "claude-3-haiku-20240307" // 默认使用haiku
	}

	maxTokens := config.Cfg.Evaluator.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 100 // 默认100 tokens
	}

	// 构建 Claude API 请求
	claudeReq := &models.ClaudeRequest{
		Model: evalModel,
		Messages: []models.Message{
			{
				Role: "user",
				Content: []models.ContentBlock{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
		MaxTokens: maxTokens,
		Stream:    false,
	}
	
	// 序列化请求
	requestBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}
	
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", service.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", service.APIKey))
	req.Header.Set("anthropic-version", "2023-06-01")
	
	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("决策者服务返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}
	
	// 解析 Claude API 响应
	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, body: %s", err, string(body))
	}
	
	// 提取难度等级
	if len(claudeResp.Content) == 0 || claudeResp.Content[0].Text == "" {
		return nil, fmt.Errorf("响应内容为空")
	}
	
	responseText := claudeResp.Content[0].Text
	difficultyLevel, reasoning := c.extractDifficultyLevel(responseText)
	if difficultyLevel < 1 || difficultyLevel > 5 {
		return nil, fmt.Errorf("无效的难度等级: %d, 响应: %s", difficultyLevel, responseText)
	}
	
	return &models.EvaluatorResponse{
		DifficultyLevel: difficultyLevel,
		Reasoning:       reasoning,
	}, nil
}

// GetContextManager 获取上下文管理器（用于测试）
func (c *Client) GetContextManager() *ContextManager {
	return c.contextManager
}

// buildEvaluationPrompt 构建评估任务复杂度的 prompt
// 使用配置文件中的prompt模板，支持变量替换
func (c *Client) buildEvaluationPrompt(evalReq *models.EvaluatorRequest) string {
	// 获取配置
	cfg := config.Cfg.Evaluator

	// 构建历史上下文信息
	var contextInfo string
	if cfg.IncludeHistory && len(evalReq.UserContext.RequestHistory) > 0 {
		// 限制历史轮数
		maxRounds := cfg.MaxHistoryRounds
		if maxRounds <= 0 {
			maxRounds = 3
		}

		historyCount := len(evalReq.UserContext.RequestHistory)
		startIdx := 0
		if historyCount > maxRounds {
			startIdx = historyCount - maxRounds
		}

		contextInfo = fmt.Sprintf("\n\n用户最近的请求历史（%d条）：", historyCount-startIdx)
		for i := startIdx; i < historyCount; i++ {
			hist := evalReq.UserContext.RequestHistory[i]
			contextInfo += fmt.Sprintf("\n%d. 模型: %s, 难度: %d, 消息数: %d, 耗时: %dms",
				i-startIdx+1, hist.Model, hist.DifficultyLevel, hist.MessageCount, hist.ResponseTime.Milliseconds())
		}
	}

	// 提取请求的关键信息
	messageCount := len(evalReq.OriginalRequest.Messages)
	model := evalReq.OriginalRequest.Model

	// 智能提取最新的用户任务内容
	currentTask := c.extractUserIntent(evalReq.OriginalRequest.Messages)

	// 如果没有提取到有效内容，尝试提取最近几轮对话的简要摘要
	if currentTask == "" {
		currentTask = c.extractRecentContext(evalReq.OriginalRequest.Messages, 3)
	}

	// 限制长度避免过长
	if len(currentTask) > 500 {
		currentTask = currentTask[:500] + "..."
	}

	// 准备模板变量
	templateData := map[string]interface{}{
		"Model":          model,
		"MessageCount":   messageCount,
		"CurrentTask":    currentTask,
		"HistoryContext": contextInfo,
	}

	// 使用模板渲染prompt
	prompt := c.renderTemplate(cfg.PromptTemplate, templateData)

	return prompt
}

// renderTemplate 渲染模板，替换变量
// 支持 {{.VarName}} 格式的变量
func (c *Client) renderTemplate(template string, data map[string]interface{}) string {
	result := template

	// 替换所有变量
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		var replacement string

		switch v := value.(type) {
		case string:
			replacement = v
		case int:
			replacement = fmt.Sprintf("%d", v)
		default:
			replacement = fmt.Sprintf("%v", v)
		}

		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}

// extractUserIntent 智能提取用户的真实意图
// 过滤掉system-reminder、tool_result、命令输出等辅助内容
func (c *Client) extractUserIntent(messages []models.Message) string {
	// 从最后一条user消息开始，向前查找
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "user" {
			continue
		}

		var userTexts []string
		for _, content := range messages[i].Content {
			if content.Type != "text" {
				continue
			}

			text := strings.TrimSpace(content.Text)

			// 过滤掉辅助内容
			if c.isAuxiliaryContent(text) {
				continue
			}

			// 收集真正的用户文本
			if text != "" {
				userTexts = append(userTexts, text)
			}
		}

		// 如果找到了有效的用户文本，返回
		if len(userTexts) > 0 {
			return strings.Join(userTexts, " ")
		}
	}

	return ""
}

// isAuxiliaryContent 判断文本是否为辅助内容（system-reminder、tool_result等）
func (c *Client) isAuxiliaryContent(text string) bool {
	// 检查是否为system-reminder
	if strings.Contains(text, "<system-reminder>") {
		return true
	}

	// 检查是否为tool_result
	if strings.Contains(text, "<tool_result>") || strings.Contains(text, "tool_result") {
		return true
	}

	// 检查是否为命令输出
	if strings.Contains(text, "<command-name>") || strings.Contains(text, "<local-command-stdout>") {
		return true
	}

	// 检查是否为User has answered提示
	if strings.HasPrefix(text, "User has answered your questions:") {
		return true
	}

	// 检查是否为工具调用结果提示
	if strings.HasPrefix(text, "File created successfully") ||
	   strings.HasPrefix(text, "Todos have been modified") ||
	   strings.Contains(text, "tool_use_id") {
		return true
	}

	return false
}

// extractRecentContext 提取最近几轮对话的简要摘要（作为备选方案）
func (c *Client) extractRecentContext(messages []models.Message, recentCount int) string {
	var contextParts []string
	userMsgCount := 0

	// 从后向前遍历，收集最近N轮的user消息
	for i := len(messages) - 1; i >= 0 && userMsgCount < recentCount; i-- {
		if messages[i].Role != "user" {
			continue
		}

		userMsgCount++

		// 提取该user消息的有效文本
		for _, content := range messages[i].Content {
			if content.Type == "text" && !c.isAuxiliaryContent(content.Text) {
				text := strings.TrimSpace(content.Text)
				if text != "" && len(text) < 200 { // 只取较短的文本
					contextParts = append([]string{text}, contextParts...) // 保持时间顺序
					break
				}
			}
		}
	}

	if len(contextParts) > 0 {
		return strings.Join(contextParts, " → ")
	}

	return "无法提取有效的用户任务内容"
}

// extractDifficultyLevel 从 AI 响应中提取难度等级
// 返回难度等级和原始响应文本
func (c *Client) extractDifficultyLevel(response string) (int, string) {
	response = strings.TrimSpace(response)
	originalResponse := response
	
	// 首先尝试解析 JSON 格式
	var jsonResp struct {
		DifficultyLevel int `json:"difficulty_level"`
	}
	
	// 尝试提取 JSON 部分（可能包含 markdown 代码块）
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &jsonResp); err == nil {
			if jsonResp.DifficultyLevel >= 1 && jsonResp.DifficultyLevel <= 5 {
				return jsonResp.DifficultyLevel, originalResponse
			}
		}
	}
	
	// 如果 JSON 解析失败，尝试直接解析整个响应
	if err := json.Unmarshal([]byte(response), &jsonResp); err == nil {
		if jsonResp.DifficultyLevel >= 1 && jsonResp.DifficultyLevel <= 5 {
			return jsonResp.DifficultyLevel, originalResponse
		}
	}
	
	// 尝试查找 JSON 中的 difficulty_level 字段
	if strings.Contains(response, "difficulty_level") {
		// 使用正则或字符串查找
		parts := strings.Split(response, "difficulty_level")
		if len(parts) > 1 {
			valuePart := parts[1]
			// 查找数字
			for _, char := range valuePart {
				if char >= '1' && char <= '5' {
					return int(char - '0'), originalResponse
				}
			}
		}
	}
	
	// 最后尝试查找第一个出现的 1-5 的数字
	for _, char := range response {
		if char >= '1' && char <= '5' {
			level := int(char - '0')
			logger.LogWarn("从响应中提取到难度等级（非JSON格式）", "level", level, "response", response)
			return level, originalResponse
		}
	}
	
	// 如果没有找到，返回默认值 3
	logger.LogWarn("无法从响应中提取难度等级，使用默认值", "response", response)
	return 3, originalResponse
}
