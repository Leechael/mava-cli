package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/phalahq/mava-api/internal/config"
)

// WsAckResult holds a Socket.IO acknowledgement response.
type WsAckResult struct {
	ID   string        `json:"id"`
	Data []interface{} `json:"data"`
}

// parseSocketIOMessage parses Engine.IO / Socket.IO framed messages.
func parseSocketIOMessage(data string) (string, interface{}) {
	if len(data) == 0 {
		return "empty", nil
	}
	switch {
	case data[0] == '0':
		var v interface{}
		json.Unmarshal([]byte(data[1:]), &v)
		return "open", v
	case data[0] == '2':
		return "ping", nil
	case data[0] == '3':
		return "pong", nil
	case strings.HasPrefix(data, "40"):
		if len(data) > 2 {
			var v interface{}
			json.Unmarshal([]byte(data[2:]), &v)
			return "connect", v
		}
		return "connect", nil
	case strings.HasPrefix(data, "42"):
		var v interface{}
		json.Unmarshal([]byte(data[2:]), &v)
		return "event", v
	case strings.HasPrefix(data, "43"):
		idx := strings.Index(data, "[")
		if idx > 0 {
			ackID := data[2:idx]
			var payload interface{}
			json.Unmarshal([]byte(data[idx:]), &payload)
			return "ack", map[string]interface{}{"id": ackID, "data": payload}
		}
		return "ack", data[2:]
	default:
		return "unknown", data
	}
}

// WsSendAndWait connects via Socket.IO, sends an event, and waits for the ack.
func WsSendAndWait(eventName string, payload interface{}, ackID int, timeout time.Duration) (map[string]interface{}, error) {
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set("Cookie", "x-auth-token="+token)

	conn, _, err := websocket.DefaultDialer.Dial(config.WSURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}
	defer conn.Close()

	// Receive Engine.IO open
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed reading open: %w", err)
	}
	msgType, _ := parseSocketIOMessage(string(msg))
	if msgType != "open" {
		return nil, fmt.Errorf("expected open packet, got %s", msgType)
	}

	// Send Socket.IO connect
	if err := conn.WriteMessage(websocket.TextMessage, []byte("40")); err != nil {
		return nil, err
	}

	// Receive connect ack
	_, _, err = conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed reading connect ack: %w", err)
	}

	// Join rooms
	conn.WriteMessage(websocket.TextMessage, []byte(`42["joinRoom"]`))
	conn.WriteMessage(websocket.TextMessage, []byte(`42["joinClientMemberNotificationRoom"]`))
	time.Sleep(300 * time.Millisecond)

	// Send event with ack ID
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	sendData := fmt.Sprintf(`42%d["%s",%s]`, ackID, eventName, string(payloadBytes))
	if err := conn.WriteMessage(websocket.TextMessage, []byte(sendData)); err != nil {
		return nil, err
	}

	// Wait for ack
	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("websocket timeout waiting for response")
		}
		mt, result := parseSocketIOMessage(string(raw))
		switch mt {
		case "ping":
			conn.WriteMessage(websocket.TextMessage, []byte("3"))
		case "ack":
			if m, ok := result.(map[string]interface{}); ok {
				return m, nil
			}
			return map[string]interface{}{"data": result}, nil
		}
	}
}
