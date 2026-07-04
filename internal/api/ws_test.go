package api

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type fakeSocketIOConn struct {
	reads  []fakeRead
	writes []string
	closed bool
}

type fakeRead struct {
	data string
	err  error
}

func (f *fakeSocketIOConn) ReadMessage() (int, []byte, error) {
	if len(f.reads) == 0 {
		return websocket.TextMessage, nil, errors.New("no more frames")
	}
	next := f.reads[0]
	f.reads = f.reads[1:]
	if next.err != nil {
		return websocket.TextMessage, nil, next.err
	}
	return websocket.TextMessage, []byte(next.data), nil
}

func (f *fakeSocketIOConn) WriteMessage(_ int, data []byte) error {
	f.writes = append(f.writes, string(data))
	return nil
}

func (f *fakeSocketIOConn) SetReadDeadline(time.Time) error { return nil }
func (f *fakeSocketIOConn) Close() error                    { f.closed = true; return nil }

func TestCompleteSocketIOHandshakeWaitsForNamespaceConnect(t *testing.T) {
	conn := &fakeSocketIOConn{reads: []fakeRead{
		{data: `0{"sid":"engine"}`},
		{data: `2`},
		{data: `40{"sid":"namespace"}`},
	}}

	if err := completeSocketIOHandshake(conn); err != nil {
		t.Fatalf("completeSocketIOHandshake() error = %v", err)
	}

	want := []string{"40", "3"}
	if !reflect.DeepEqual(conn.writes, want) {
		t.Fatalf("writes = %#v, want %#v", conn.writes, want)
	}
}

func TestSendSocketIOEventWithAckIgnoresMismatchedAckAndPongs(t *testing.T) {
	conn := &fakeSocketIOConn{reads: []fakeRead{
		{data: `431[{"status":500}]`},
		{data: `2`},
		{data: `432[{"status":204}]`},
	}}

	got, err := sendSocketIOEventWithAck(conn, "ticketUpdate", map[string]string{"ticketId": "abc"}, 2, time.Second)
	if err != nil {
		t.Fatalf("sendSocketIOEventWithAck() error = %v", err)
	}

	data, ok := got["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("ack data = %#v, want one response object", got["data"])
	}
	first, ok := data[0].(map[string]interface{})
	if !ok || int(first["status"].(float64)) != 204 {
		t.Fatalf("ack first = %#v, want status 204", data[0])
	}

	if len(conn.writes) < 2 || !strings.HasPrefix(conn.writes[0], `422["ticketUpdate",`) || conn.writes[1] != "3" {
		t.Fatalf("writes = %#v, want ticketUpdate ack id 2 followed by pong", conn.writes)
	}
}

func TestWsSendAndWaitRetriesAfterAckTimeout(t *testing.T) {
	attempts := 0
	dialer := func(token string) (socketIOConn, error) {
		attempts++
		if token != "token-for-test" {
			t.Fatalf("token = %q, want token-for-test", token)
		}
		if attempts == 1 {
			return &fakeSocketIOConn{reads: []fakeRead{
				{data: `0{"sid":"engine-a"}`},
				{data: `40{"sid":"namespace-a"}`},
				{err: errors.New("read timeout")},
			}}, nil
		}
		return &fakeSocketIOConn{reads: []fakeRead{
			{data: `0{"sid":"engine-b"}`},
			{data: `40{"sid":"namespace-b"}`},
			{data: `431[{"status":204}]`},
		}}, nil
	}

	got, err := wsSendAndWait("token-for-test", "ticketUpdate", map[string]string{"ticketId": "abc"}, 1, time.Second, dialer, 2)
	if err != nil {
		t.Fatalf("wsSendAndWait() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	data := got["data"].([]interface{})
	status := int(data[0].(map[string]interface{})["status"].(float64))
	if status != 204 {
		t.Fatalf("status = %d, want 204", status)
	}
}
