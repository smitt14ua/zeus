// Package protocol defines the WebSocket message types shared between the
// zeus agent and any compatible web server implementation.
// It has no internal dependencies and can be extracted into its own module.
package protocol

import (
	"encoding/json"
	"slices"
)

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
	CmdSystemReboot = "system.reboot"
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

// Scope names for the --allow flag on zeus agent.
// Scopes restrict which commands the agent will accept and execute.
const (
	ScopeView    = "view"    // profile.list, profile.info (read-only)
	ScopeControl = "control" // profile.start, profile.stop
	ScopeManage  = "manage"  // profile.new, profile.add, profile.rm, missions.pull, profile.info (write operations)
	ScopeUpdate  = "update"  // update
	ScopeSystem  = "system"  // system.reboot (host machine operations)
	ScopeAll     = "all"     // every scope except OptInScopes (default; currently includes system)
)

// scopeCommands maps each scope to the commands it covers.
var scopeCommands = map[string][]string{
	ScopeView:    {CmdProfileList, CmdProfileInfo},
	ScopeControl: {CmdProfileStart, CmdProfileStop},
	ScopeManage:  {CmdProfileNew, CmdProfileAdd, CmdProfileRm, CmdMissionsPull},
	ScopeUpdate:  {CmdUpdate},
	ScopeSystem:  {CmdSystemReboot},
}

// AllScopes is the ordered list of all named scopes (excluding "all").
var AllScopes = []string{ScopeView, ScopeControl, ScopeManage, ScopeUpdate, ScopeSystem}

// OptInScopes are never granted implicitly: an empty scope list and ScopeAll
// exclude them, so they must be named explicitly (e.g. --allow all,system).
//
// TEMPORARY (v0.7.1): empty, so ScopeSystem is granted by default and by
// ScopeAll. Restore {ScopeSystem: true} in the next release.
var OptInScopes = map[string]bool{}

// ResolveScopes returns the set of allowed command names for the given scope list.
// An empty list or one containing ScopeAll grants every scope except OptInScopes.
func ResolveScopes(scopes []string) map[string]bool {
	all := len(scopes) == 0 || slices.Contains(scopes, ScopeAll)
	allowed := make(map[string]bool)
	for scope, cmds := range scopeCommands {
		if (all && !OptInScopes[scope]) || slices.Contains(scopes, scope) {
			for _, cmd := range cmds {
				allowed[cmd] = true
			}
		}
	}
	return allowed
}

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
	Agent         string   `json:"agent"`
	ZeusVersion   string   `json:"zeus_version"`
	AllowedScopes []string `json:"allowed_scopes,omitempty"` // nil/absent = all scopes except OptInScopes
}

// Heartbeat is the payload for TypeHeartbeat messages (agent → server, periodic).
type Heartbeat struct {
	Agent         string          `json:"agent"`
	ZeusVersion   string          `json:"zeus_version"`
	Profiles      []ProfileStatus `json:"profiles"`
	AllowedScopes []string        `json:"allowed_scopes,omitempty"` // nil/absent = all scopes except OptInScopes
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
