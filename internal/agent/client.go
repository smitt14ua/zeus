package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/smitt14ua/zeus/internal/protocol"
	"github.com/smitt14ua/zeus/internal/process"
	"github.com/smitt14ua/zeus/internal/storage"
)

// StatusFunc returns the current running state of all profiles.
// Client calls it for each heartbeat.
type StatusFunc func() []protocol.ProfileStatus

// Client connects to a zeus web panel, authenticates with a token, and
// dispatches incoming commands to the Executor. It reconnects automatically
// on disconnect.
type Client struct {
	URL               string
	Token             string
	Name              string
	Version           string
	HeartbeatInterval time.Duration
	ReconnectDelay    time.Duration
	Executor          *Executor
	AllowedScopes     []string // nil = all scopes; sent in hello/heartbeat so the panel knows

	mu   sync.Mutex // guards all WebSocket writes
	conn *websocket.Conn
}

// Run connects and blocks until ctx is cancelled.
func (c *Client) Run(ctx context.Context) error {
	for {
		if err := c.connect(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			fmt.Printf("zeus agent: disconnected: %v — reconnecting in %s\n", err, c.ReconnectDelay)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.ReconnectDelay):
		}
	}
}

func (c *Client) connect(ctx context.Context) error {
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+c.Token)

	url := c.URL + "?name=" + c.Name
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, hdr)
	if err != nil {
		return err
	}
	defer conn.Close()

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	fmt.Printf("zeus agent: connected to %s as %q\n", c.URL, c.Name)

	if err := c.sendHello(); err != nil {
		return err
	}

	connCtx, cancelConn := context.WithCancel(ctx)
	hbDone := make(chan struct{})
	defer func() { <-hbDone }() // wait for heartbeat to stop (registered first, runs second)
	defer cancelConn()           // stop heartbeat goroutine (registered second, runs first)
	go func() {
		defer close(hbDone)
		c.heartbeatLoop(connCtx)
	}()

	return c.readLoop(ctx, conn)
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var env protocol.Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			continue
		}

		switch env.Type {
		case protocol.TypePing:
			c.writeRaw(protocol.TypePong, env.ID, struct{}{})
		case protocol.TypeCommand:
			var cmd protocol.Command
			if err := json.Unmarshal(env.Payload, &cmd); err != nil {
				c.sendResult(env.ID, false, 1, "malformed command payload")
				continue
			}
			go c.runCommand(ctx, env.ID, cmd)
		}
	}
}

func (c *Client) runCommand(ctx context.Context, id string, cmd protocol.Command) {
	w := &streamWriter{
		conn: c.conn,
		mu:   &c.mu,
		id:   id,
		fd:   protocol.FDStdout,
	}
	err := c.Executor.Dispatch(ctx, cmd.Cmd, cmd.Args, w)
	_ = w.Close()

	shouldRestart := errors.Is(err, ErrRestartRequested)
	if err != nil && !shouldRestart {
		c.sendResult(id, false, 1, err.Error())
	} else {
		c.sendResult(id, true, 0, "")
	}
	c.sendHeartbeat()

	if shouldRestart {
		restartSelf()
	}
}

func (c *Client) sendResult(id string, success bool, exitCode int, errMsg string) {
	c.writeRaw(protocol.TypeResult, id, protocol.Result{
		Success:  success,
		ExitCode: exitCode,
		Error:    errMsg,
	})
}

func (c *Client) sendHello() error {
	return c.writeRaw(protocol.TypeHello, "", protocol.Hello{
		Agent:         c.Name,
		ZeusVersion:   c.Version,
		AllowedScopes: c.AllowedScopes,
	})
}

func (c *Client) heartbeatLoop(ctx context.Context) {
	c.sendHeartbeat()
	ticker := time.NewTicker(c.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
		}
	}
}

func (c *Client) sendHeartbeat() {
	c.writeRaw(protocol.TypeHeartbeat, "", protocol.Heartbeat{
		Agent:         c.Name,
		ZeusVersion:   c.Version,
		Profiles:      collectProfileStatus(),
		AllowedScopes: c.AllowedScopes,
	})
}

func (c *Client) writeRaw(msgType, id string, payload any) error {
	data, err := protocol.Encode(msgType, id, payload)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// collectProfileStatus reads saved profiles and checks which are running.
func collectProfileStatus() []protocol.ProfileStatus {
	repo := storage.ProfileRepository{}
	profiles, err := repo.List()
	if err != nil {
		return nil
	}

	mgr := process.Manager{}
	running, _ := mgr.List()
	runningMap := make(map[string]int, len(running))
	for _, e := range running {
		runningMap[e.Name] = e.PID
	}

	statuses := make([]protocol.ProfileStatus, 0, len(profiles))
	for _, p := range profiles {
		pid := runningMap[p.Name]
		statuses = append(statuses, protocol.ProfileStatus{
			Name:    p.Name,
			Running: pid != 0,
			PID:     pid,
		})
	}
	return statuses
}
