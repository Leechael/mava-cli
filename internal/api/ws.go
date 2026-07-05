package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
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

type socketIOConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	SetReadDeadline(time.Time) error
	Close() error
}

type socketIODialer func(token string) (socketIOConn, error)

// ErrWebSocketAckTimeout is returned when the event was sent but no matching ack arrived before the read deadline.
var ErrWebSocketAckTimeout = errors.New("websocket timeout waiting for response")

func defaultSocketIODialer(token string) (socketIOConn, error) {
	header := http.Header{}
	header.Set("Cookie", "x-auth-token="+token)
	// Match the browser dashboard socket handshake more closely. The Mava
	// socket server can intermittently drop acks for bare non-browser writes.
	header.Set("Origin", "https://dashboard.mava.app")

	conn, _, err := websocket.DefaultDialer.Dial(config.WSURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}
	return conn, nil
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

func completeSocketIOHandshake(conn socketIOConn) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed reading open: %w", err)
		}
		msgType, _ := parseSocketIOMessage(string(msg))
		switch msgType {
		case "ping":
			if err := conn.WriteMessage(websocket.TextMessage, []byte("3")); err != nil {
				return err
			}
		case "open":
			if err := conn.WriteMessage(websocket.TextMessage, []byte("40")); err != nil {
				return err
			}
			return waitForSocketIOConnect(conn)
		default:
			return fmt.Errorf("expected open packet, got %s", msgType)
		}
	}
}

func waitForSocketIOConnect(conn socketIOConn) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed reading connect ack: %w", err)
		}
		msgType, _ := parseSocketIOMessage(string(msg))
		switch msgType {
		case "ping":
			if err := conn.WriteMessage(websocket.TextMessage, []byte("3")); err != nil {
				return err
			}
		case "connect":
			return nil
		}
	}
}

func joinDefaultSocketIORooms(conn socketIOConn) error {
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`42["joinRoom"]`)); err != nil {
		return err
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`42["joinClientMemberNotificationRoom"]`)); err != nil {
		return err
	}
	return nil
}

func sendSocketIOEventWithAck(conn socketIOConn, eventName string, payload interface{}, ackID int, timeout time.Duration) (map[string]interface{}, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	sendData := fmt.Sprintf(`42%d["%s",%s]`, ackID, eventName, string(payloadBytes))
	if err := conn.WriteMessage(websocket.TextMessage, []byte(sendData)); err != nil {
		return nil, err
	}

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	wantAckID := strconv.Itoa(ackID)
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if isTimeoutError(err) {
				return nil, fmt.Errorf("%w: %v", ErrWebSocketAckTimeout, err)
			}
			return nil, fmt.Errorf("websocket read failed waiting for response: %w", err)
		}
		mt, result := parseSocketIOMessage(string(raw))
		switch mt {
		case "ping":
			if err := conn.WriteMessage(websocket.TextMessage, []byte("3")); err != nil {
				return nil, err
			}
		case "ack":
			m, ok := result.(map[string]interface{})
			if !ok {
				continue
			}
			if gotAckID, _ := m["id"].(string); gotAckID != wantAckID {
				continue
			}
			return m, nil
		}
	}
}

func wsSendAndWait(token string, eventName string, payload interface{}, ackID int, timeout time.Duration, dial socketIODialer, attempts int) (map[string]interface{}, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		result, err := wsSendAndWaitOnce(token, eventName, payload, ackID, timeout, dial)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !errors.Is(err, ErrWebSocketAckTimeout) {
			return nil, err
		}
	}
	return nil, lastErr
}

func isTimeoutError(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func wsSendAndWaitOnce(token string, eventName string, payload interface{}, ackID int, timeout time.Duration, dial socketIODialer) (map[string]interface{}, error) {
	conn, err := dial(token)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	if err := completeSocketIOHandshake(conn); err != nil {
		return nil, err
	}
	if err := joinDefaultSocketIORooms(conn); err != nil {
		return nil, err
	}
	return sendSocketIOEventWithAck(conn, eventName, payload, ackID, timeout)
}

// WsSendAndWait connects via Socket.IO, sends an event, and waits for the ack.
func WsSendAndWait(eventName string, payload interface{}, ackID int, timeout time.Duration) (map[string]interface{}, error) {
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}
	return wsSendAndWait(token, eventName, payload, ackID, timeout, defaultSocketIODialer, 1)
}

// WsSendAndWaitRetryOnAckTimeout retries once after an ack timeout.
// Only use this for operations that are safe to repeat or verify with a read-back.
func WsSendAndWaitRetryOnAckTimeout(eventName string, payload interface{}, ackID int, timeout time.Duration) (map[string]interface{}, error) {
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}
	return wsSendAndWait(token, eventName, payload, ackID, timeout, defaultSocketIODialer, 2)
}
