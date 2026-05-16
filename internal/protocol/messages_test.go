package protocol

import (
	"encoding/json"
	"testing"
)

func TestEncode_RoundTrip(t *testing.T) {
	payload := Result{Success: true, ExitCode: 0}
	raw, err := Encode(TypeResult, "abc", payload)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("Unmarshal envelope: %v", err)
	}
	if env.Type != TypeResult {
		t.Errorf("Type = %q, want %q", env.Type, TypeResult)
	}
	if env.ID != "abc" {
		t.Errorf("ID = %q, want %q", env.ID, "abc")
	}

	var got Result
	if err := json.Unmarshal(env.Payload, &got); err != nil {
		t.Fatalf("Unmarshal payload: %v", err)
	}
	if !got.Success || got.ExitCode != 0 {
		t.Errorf("payload round-trip mismatch: %+v", got)
	}
}

func TestEncode_NoID(t *testing.T) {
	raw, err := Encode(TypeHeartbeat, "", Heartbeat{Agent: "srv"})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if env.ID != "" {
		t.Errorf("ID should be omitted, got %q", env.ID)
	}
}
