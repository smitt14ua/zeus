// Package protocol defines the WebSocket message types shared between the
// zeus agent and any compatible web server implementation.
// It has no internal dependencies and can be extracted into its own module.
package protocol

import "encoding/json"

// Message types (Envelope.Type).
const (
	TypeHello     = "hello"
	TypeHeartbeat = "heartbeat"
	TypeCommand   = "command"
	TypeStream    = "stream"
	TypeResult    = "result"
	TypeEvent     = "event"
	TypePing      = "ping"
	TypePong      = "pong"
)

// Command names (Command.Cmd).
const (
	CmdProfileNew   = "profile.new"
	CmdProfileAdd   = "profile.add"
	CmdProfileList  = "profile.list"
	CmdProfileInfo  = "profile.info"
	CmdProfileStart = "profile.start"
	CmdProfileStop  = "profile.stop"
	CmdProfileRm    = "profile.rm"
	CmdMissionsPull = "missions.pull"
	CmdUpdate       = "update"
)

// Event names (Event.Event).
const (
	EventProfileStarted = "profile.started"
	EventProfileStopped = "profile.stopped"
)

// FD labels for Stream messages.
const (
	FDStdout = "stdout"
	FDStderr = "stderr"
)

// Envelope is the top-level wrapper for every message in both directions.
type Envelope struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Command is the payload for TypeCommand messages (server → agent).
type Command struct {
	Cmd  string         `json:"cmd"`
	Args map[string]any `json:"args,omitempty"`
}

// Stream is the payload for TypeStream messages (agent → server).
// One message per output line while a command is running.
type Stream struct {
	Line string `json:"line"`
	FD   string `json:"fd"` // FDStdout or FDStderr
}

// Result is the payload for TypeResult messages (agent → server).
// Sent once when a command finishes.
type Result struct {
	Success  bool   `json:"success"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

// Hello is the payload for TypeHello messages (agent → server, sent on connect).
type Hello struct {
	Agent       string `json:"agent"`
	ZeusVersion string `json:"zeus_version"`
}

// Heartbeat is the payload for TypeHeartbeat messages (agent → server, periodic).
type Heartbeat struct {
	Agent       string          `json:"agent"`
	ZeusVersion string          `json:"zeus_version"`
	Profiles    []ProfileStatus `json:"profiles"`
}

// ProfileStatus is the per-profile snapshot inside Heartbeat.
type ProfileStatus struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
	PID     int    `json:"pid,omitempty"`
}

// Event is the payload for TypeEvent messages (agent → server, spontaneous).
type Event struct {
	Event  string `json:"event"`
	Name   string `json:"name,omitempty"`
	PID    int    `json:"pid,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Encode marshals a payload into an Envelope ready for sending.
func Encode(msgType, id string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{ID: id, Type: msgType, Payload: raw})
}
