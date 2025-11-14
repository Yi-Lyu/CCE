package tests

import (
	"testing"
	
	"github.com/ethan/claude-proxy/internal/models"
)

func TestExtractUserInfo(t *testing.T) {
	tests := []struct {
		name          string
		metadata      models.RequestMetadata
		expectedUser  string
		expectedSession string
	}{
		{
			name: "标准格式",
			metadata: models.RequestMetadata{
				UserID: "user_4d9e1ae2fbecbcb2af13c108249fe9dcd2c3dc9f9bb8a482196b2fea322b71d9_account__session_88b74551-e948-440a-94a2-ebea22189fa9",
			},
			expectedUser:    "4d9e1ae2fbecbcb2af13c108249fe9dcd2c3dc9f9bb8a482196b2fea322b71d9",
			expectedSession: "88b74551-e948-440a-94a2-ebea22189fa9",
		},
		{
			name: "空用户ID",
			metadata: models.RequestMetadata{
				UserID: "",
			},
			expectedUser:    "",
			expectedSession: "",
		},
		{
			name: "不完整格式",
			metadata: models.RequestMetadata{
				UserID: "user_12345",
			},
			expectedUser:    "",
			expectedSession: "",
		},
		{
			name: "只有用户没有会话",
			metadata: models.RequestMetadata{
				UserID: "user_4d9e1ae2fbecbcb2af13c108249fe9dcd2c3dc9f9bb8a482196b2fea322b71d9_account",
			},
			expectedUser:    "4d9e1ae2fbecbcb2af13c108249fe9dcd2c3dc9f9bb8a482196b2fea322b71d9",
			expectedSession: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, sessionID := models.ExtractUserInfo(tt.metadata)
			
			if userID != tt.expectedUser {
				t.Errorf("ExtractUserInfo() userID = %v, want %v", userID, tt.expectedUser)
			}
			if sessionID != tt.expectedSession {
				t.Errorf("ExtractUserInfo() sessionID = %v, want %v", sessionID, tt.expectedSession)
			}
		})
	}
}
