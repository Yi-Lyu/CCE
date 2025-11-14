package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/ethan/claude-proxy/internal/models"
)

// MockEvaluatorServer 模拟决策者服务
func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	
	// 决策端点
	r.POST("/api/v1/messages/evaluate-difficulty", handleEvaluate)
	r.POST("/evaluate-difficulty", handleEvaluate)
	
	// 模拟 Claude API 端点（用于测试完整流程）
	r.POST("/api/v1/messages", handleMessages)
	
	port := ":8081"
	fmt.Printf("模拟决策者服务启动在 %s\n", port)
	log.Fatal(r.Run(port))
}

// handleEvaluate 处理决策请求
func handleEvaluate(c *gin.Context) {
	var req models.EvaluatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	// 模拟决策逻辑
	difficulty := evaluateDifficulty(&req)
	
	c.JSON(200, models.EvaluatorResponse{
		DifficultyLevel: difficulty,
		Reasoning:      fmt.Sprintf("基于模型 %s 和 %d 条消息的分析", req.OriginalRequest.Model, len(req.OriginalRequest.Messages)),
	})
}

// evaluateDifficulty 模拟难度评估
func evaluateDifficulty(req *models.EvaluatorRequest) int {
	// 简单的模拟逻辑
	messageCount := len(req.OriginalRequest.Messages)
	model := req.OriginalRequest.Model
	
	// 基于模型名称
	if strings.Contains(model, "haiku") {
		return 1
	} else if strings.Contains(model, "sonnet") {
		return 3
	} else if strings.Contains(model, "opus") {
		return 5
	}
	
	// 基于消息数量
	if messageCount <= 1 {
		return 1
	} else if messageCount <= 3 {
		return 2
	} else if messageCount <= 5 {
		return 3
	} else if messageCount <= 10 {
		return 4
	}
	
	return 5
}

// handleMessages 模拟 Claude API 响应
func handleMessages(c *gin.Context) {
	var req models.ClaudeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	// 检查认证
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(401, gin.H{"error": "未授权"})
		return
	}
	
	if req.Stream {
		// 流式响应
		handleStreamResponse(c, &req)
	} else {
		// 普通响应
		handleNormalResponse(c, &req)
	}
}

// handleNormalResponse 处理普通响应
func handleNormalResponse(c *gin.Context, req *models.ClaudeRequest) {
	response := gin.H{
		"id":      "msg_test_" + fmt.Sprintf("%d", time.Now().Unix()),
		"type":    "message",
		"role":    "assistant",
		"model":   req.Model,
		"content": []gin.H{
			{
				"type": "text",
				"text": fmt.Sprintf("这是来自模拟服务的响应。收到 %d 条消息。", len(req.Messages)),
			},
		},
		"stop_reason": "end_turn",
		"usage": gin.H{
			"input_tokens":  100,
			"output_tokens": 50,
		},
	}
	
	c.JSON(200, response)
}

// handleStreamResponse 处理流式响应
func handleStreamResponse(c *gin.Context, req *models.ClaudeRequest) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	
	w := c.Writer
	flusher, _ := w.(http.Flusher)
	
	// 发送开始事件
	fmt.Fprintf(w, "event: message_start\n")
	fmt.Fprintf(w, "data: %s\n\n", `{"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","model":"`+req.Model+`","content":[],"stop_reason":null}}`)
	flusher.Flush()
	
	// 发送内容
	messages := []string{
		"这是",
		"来自模拟服务的",
		"流式响应。",
		fmt.Sprintf("收到 %d 条消息。", len(req.Messages)),
	}
	
	for i, msg := range messages {
		fmt.Fprintf(w, "event: content_block_delta\n")
		data := map[string]interface{}{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]string{
				"type": "text_delta",
				"text": msg,
			},
		}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
		
		// 模拟延迟
		time.Sleep(100 * time.Millisecond)
	}
	
	// 发送结束事件
	fmt.Fprintf(w, "event: message_stop\n")
	fmt.Fprintf(w, "data: %s\n\n", `{"type":"message_stop"}`)
	flusher.Flush()
}
