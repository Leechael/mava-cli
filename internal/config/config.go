package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	BaseURL      = "https://gateway.mava.app"
	WSURL        = "wss://socket.mava.app/socket.io/?EIO=4&transport=websocket"
	DashboardURL = "https://dashboard.mava.app/dashboard/ticket?id="
)

func GetToken() (string, error) {
	token := os.Getenv("MAVA_TOKEN")
	if token == "" {
		return "", fmt.Errorf("MAVA_TOKEN environment variable is not set")
	}
	return token, nil
}

func GetCurrentUserID() string {
	token, err := GetToken()
	if err != nil {
		return ""
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload := parts[1]
	// Add base64 padding if needed
	if m := len(payload) % 4; m != 0 {
		payload += strings.Repeat("=", 4-m)
	}
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal(decoded, &data); err != nil {
		return ""
	}
	if id, ok := data["clientMemberId"].(string); ok {
		return id
	}
	return ""
}
