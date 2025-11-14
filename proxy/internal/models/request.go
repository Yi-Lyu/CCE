package models

import (
	"encoding/json"
	"time"
)

// ClaudeRequest Claude API 请求结构
type ClaudeRequest struct {
	Model    string            `json:"model"`
	Messages []Message         `json:"messages"`
	System   []SystemMessage   `json:"system,omitempty"`
	Tools    []interface{}     `json:"tools,omitempty"`
	Metadata RequestMetadata   `json:"metadata,omitempty"`
	MaxTokens int              `json:"max_tokens,omitempty"`
	Stream   bool              `json:"stream,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"-"` // 使用自定义解析
}

// UnmarshalJSON 自定义 JSON 解析，支持 content 字段的两种格式
// 1. 字符串格式: "content": "hello"
// 2. 数组格式: "content": [{"type": "text", "text": "hello"}]
func (m *Message) UnmarshalJSON(data []byte) error {
	// 使用辅助结构体避免递归调用
	type Alias Message
	aux := &struct {
		Content json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	// 解析 JSON
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 尝试将 content 解析为数组格式
	var contentBlocks []ContentBlock
	if err := json.Unmarshal(aux.Content, &contentBlocks); err == nil {
		m.Content = contentBlocks
		return nil
	}

	// 如果数组格式失败，尝试解析为字符串格式
	var contentStr string
	if err := json.Unmarshal(aux.Content, &contentStr); err != nil {
		return err
	}

	// 将字符串转换为 ContentBlock 数组
	m.Content = []ContentBlock{
		{
			Type: "text",
			Text: contentStr,
		},
	}

	return nil
}

// MarshalJSON 自定义 JSON 序列化，确保 content 字段被正确输出
func (m Message) MarshalJSON() ([]byte, error) {
	// 使用辅助结构体避免递归调用
	type Alias Message
	return json.Marshal(&struct {
		Role    string         `json:"role"`
		Content []ContentBlock `json:"content"`
	}{
		Role:    m.Role,
		Content: m.Content,
	})
}

// SystemMessage 系统消息
type SystemMessage struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl CacheControl  `json:"cache_control,omitempty"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text,omitempty"`
	CacheControl CacheControl  `json:"cache_control,omitempty"`
}

// CacheControl 缓存控制
type CacheControl struct {
	Type string `json:"type"`
}

// RequestMetadata 请求元数据
type RequestMetadata struct {
	UserID string `json:"user_id"`
}

// UserContext 用户上下文信息
type UserContext struct {
	UserID         string
	SessionID      string
	RequestHistory []RequestSummary
}

// RequestSummary 请求摘要，用于维护历史记录
type RequestSummary struct {
	Timestamp      time.Time
	Model          string
	MessageCount   int
	TokenCount     int
	DifficultyLevel int
	ResponseTime   time.Duration
}

// EvaluatorRequest 发送给决策者服务的请求
type EvaluatorRequest struct {
	OriginalRequest ClaudeRequest `json:"original_request"`
	UserContext     UserContext   `json:"user_context"`
}

// EvaluatorResponse 决策者服务的响应
type EvaluatorResponse struct {
	DifficultyLevel int    `json:"difficulty_level"` // 1-5
	Reasoning       string `json:"reasoning,omitempty"`
}

// ExtractUserInfo 从 metadata 中提取用户ID和会话ID
func ExtractUserInfo(metadata RequestMetadata) (userID, sessionID string) {
	// 示例: user_4d9e1ae2fbecbcb2af13c108249fe9dcd2c3dc9f9bb8a482196b2fea322b71d9_account__session_88b74551-e948-440a-94a2-ebea22189fa9
	userIDStr := metadata.UserID
	if userIDStr == "" {
		return "", ""
	}
	
	// 解析 user_id
	if len(userIDStr) > 5 && userIDStr[:5] == "user_" {
		parts := []rune(userIDStr)
		// 查找第一个 "_account" 的位置
		for i := 0; i < len(parts)-8; i++ {
			if string(parts[i:i+8]) == "_account" {
				// user_id 是从 user_ 之后到 _account 之前
				userID = string(parts[5:i])
				
				// 查找 session_id
				sessionPrefix := "__session_"
				for j := i + 8; j < len(parts)-len(sessionPrefix); j++ {
					if string(parts[j:j+len(sessionPrefix)]) == sessionPrefix {
						sessionID = string(parts[j+len(sessionPrefix):])
						break
					}
				}
				break
			}
		}
	}
	
	return userID, sessionID
}

// IsWarmupRequest 检测是否为 Warmup 请求
// Warmup 请求是 Claude Code 用于预热服务端缓存的特殊请求
func IsWarmupRequest(req *ClaudeRequest) bool {
	// 检查是否有消息
	if len(req.Messages) == 0 {
		return false
	}

	// 检查第一条消息是否为用户消息
	firstMsg := req.Messages[0]
	if firstMsg.Role != "user" {
		return false
	}

	// 检查内容是否包含 "Warmup" 文本
	for _, content := range firstMsg.Content {
		if content.Type == "text" && content.Text == "Warmup" {
			return true
		}
	}

	return false
}
