package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/smitt14ua/zeus/internal/protocol"
)

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// wsConnPair starts an in-process WebSocket server, connects a client to it,
// and returns the client connection along with a channel that receives every
// raw message the server reads.
func wsConnPair(t *testing.T) (*websocket.Conn, <-chan []byte) {
	t.Helper()
	ch := make(chan []byte, 32)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			ch <- data
		}
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("wsConnPair dial: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	return client, ch
}

func decodeStream(t *testing.T, data []byte) protocol.Stream {
	t.Helper()
	var env protocol.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Type != protocol.TypeStream {
		t.Fatalf("type = %q, want %q", env.Type, protocol.TypeStream)
	}
	var s protocol.Stream
	if err := json.Unmarshal(env.Payload, &s); err != nil {
		t.Fatalf("unmarshal stream payload: %v", err)
	}
	return s
}

func TestStreamWriter_CompleteLines(t *testing.T) {
	conn, msgs := wsConnPair(t)
	mu := &sync.Mutex{}
	w := &streamWriter{conn: conn, mu: mu, id: "cmd-1", fd: protocol.FDStdout}

	n, err := w.Write([]byte("line one\nline two\n"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != 18 {
		t.Errorf("n = %d, want 18", n)
	}

	s1 := decodeStream(t, <-msgs)
	s2 := decodeStream(t, <-msgs)

	if s1.Line != "line one" {
		t.Errorf("line 1 = %q", s1.Line)
	}
	if s1.FD != protocol.FDStdout {
		t.Errorf("fd = %q", s1.FD)
	}
	if s2.Line != "line two" {
		t.Errorf("line 2 = %q", s2.Line)
	}
}

func TestStreamWriter_MultipleWritesSameLine(t *testing.T) {
	conn, msgs := wsConnPair(t)
	mu := &sync.Mutex{}
	w := &streamWriter{conn: conn, mu: mu, id: "cmd-2", fd: protocol.FDStdout}

	w.Write([]byte("par"))
	w.Write([]byte("tial"))

	// Nothing sent yet — no newline.
	select {
	case raw := <-msgs:
		t.Fatalf("unexpected message before newline: %s", raw)
	default:
	}

	w.Write([]byte("\n"))
	s := decodeStream(t, <-msgs)
	if s.Line != "partial" {
		t.Errorf("line = %q, want %q", s.Line, "partial")
	}
}

func TestStreamWriter_FlushPartialOnClose(t *testing.T) {
	conn, msgs := wsConnPair(t)
	mu := &sync.Mutex{}
	w := &streamWriter{conn: conn, mu: mu, id: "cmd-3", fd: protocol.FDStderr}

	w.Write([]byte("no newline here"))

	select {
	case raw := <-msgs:
		t.Fatalf("unexpected early message: %s", raw)
	default:
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s := decodeStream(t, <-msgs)
	if s.Line != "no newline here" {
		t.Errorf("line = %q", s.Line)
	}
	if s.FD != protocol.FDStderr {
		t.Errorf("fd = %q", s.FD)
	}
}

func TestStreamWriter_CloseEmptyBuffer(t *testing.T) {
	conn, msgs := wsConnPair(t)
	mu := &sync.Mutex{}
	w := &streamWriter{conn: conn, mu: mu, id: "cmd-4", fd: protocol.FDStdout}

	if err := w.Close(); err != nil {
		t.Fatalf("Close on empty buffer: %v", err)
	}
	select {
	case raw := <-msgs:
		t.Fatalf("unexpected message on empty close: %s", raw)
	default:
	}
}

func TestStreamWriter_IDAndFDPropagated(t *testing.T) {
	conn, msgs := wsConnPair(t)
	mu := &sync.Mutex{}
	w := &streamWriter{conn: conn, mu: mu, id: "my-id", fd: protocol.FDStderr}

	w.Write([]byte("hello\n"))

	raw := <-msgs
	var env protocol.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.ID != "my-id" {
		t.Errorf("ID = %q, want %q", env.ID, "my-id")
	}
	s := decodeStream(t, raw)
	if s.FD != protocol.FDStderr {
		t.Errorf("fd = %q, want %q", s.FD, protocol.FDStderr)
	}
}
