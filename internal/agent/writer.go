package agent

import (
	"bytes"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/smitt14ua/zeus/internal/protocol"
)

// streamWriter is an io.Writer that buffers bytes, splits on newlines, and
// sends each complete line as a TypeStream WebSocket message tagged with the
// originating command ID. Partial lines are flushed by Close.
type streamWriter struct {
	conn *websocket.Conn
	mu   *sync.Mutex // shared with Client; serialises all WebSocket writes
	id   string
	fd   string
	buf  bytes.Buffer
}

func (w *streamWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		idx := bytes.IndexByte(w.buf.Bytes(), '\n')
		if idx < 0 {
			break
		}
		line := string(w.buf.Next(idx + 1))
		line = line[:len(line)-1] // strip trailing \n
		if err := w.sendLine(line); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// Close flushes any buffered content that did not end with a newline.
func (w *streamWriter) Close() error {
	if w.buf.Len() == 0 {
		return nil
	}
	return w.sendLine(w.buf.String())
}

func (w *streamWriter) sendLine(line string) error {
	msg, err := protocol.Encode(protocol.TypeStream, w.id, protocol.Stream{
		Line: line,
		FD:   w.fd,
	})
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}
