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

func TestEncode_MarshalError(t *testing.T) {
	// channels cannot be JSON-marshaled
	_, err := Encode(TypeStream, "id", make(chan int))
	if err == nil {
		t.Fatal("expected error for un-marshalable payload")
	}
}

func TestResolveScopes_Nil_AllowsAll(t *testing.T) {
	if ResolveScopes(nil) != nil {
		t.Error("nil scopes should return nil (allow all)")
	}
}

func TestResolveScopes_Empty_AllowsAll(t *testing.T) {
	if ResolveScopes([]string{}) != nil {
		t.Error("empty scopes should return nil (allow all)")
	}
}

func TestResolveScopes_ScopeAll_AllowsAll(t *testing.T) {
	if ResolveScopes([]string{ScopeAll}) != nil {
		t.Error("ScopeAll should return nil (allow all)")
	}
}

func TestResolveScopes_AllInMixed_AllowsAll(t *testing.T) {
	if ResolveScopes([]string{ScopeView, ScopeAll}) != nil {
		t.Error("presence of ScopeAll should return nil regardless of other scopes")
	}
}

func TestResolveScopes_View(t *testing.T) {
	allowed := ResolveScopes([]string{ScopeView})
	if !allowed[CmdProfileList] {
		t.Error("profile.list should be allowed by view scope")
	}
	if !allowed[CmdProfileInfo] {
		t.Error("profile.info should be allowed by view scope")
	}
	if allowed[CmdProfileStart] {
		t.Error("profile.start should not be allowed by view scope")
	}
	if allowed[CmdUpdate] {
		t.Error("update should not be allowed by view scope")
	}
}

func TestResolveScopes_Control(t *testing.T) {
	allowed := ResolveScopes([]string{ScopeControl})
	if !allowed[CmdProfileStart] {
		t.Error("profile.start should be allowed by control scope")
	}
	if !allowed[CmdProfileStop] {
		t.Error("profile.stop should be allowed by control scope")
	}
	if allowed[CmdProfileList] {
		t.Error("profile.list should not be allowed by control scope")
	}
	if allowed[CmdProfileAdd] {
		t.Error("profile.add should not be allowed by control scope")
	}
}

func TestResolveScopes_Manage(t *testing.T) {
	allowed := ResolveScopes([]string{ScopeManage})
	if !allowed[CmdProfileAdd] {
		t.Error("profile.add should be allowed by manage scope")
	}
	if !allowed[CmdProfileRm] {
		t.Error("profile.rm should be allowed by manage scope")
	}
	if !allowed[CmdMissionsPull] {
		t.Error("missions.pull should be allowed by manage scope")
	}
	if allowed[CmdProfileStart] {
		t.Error("profile.start should not be allowed by manage scope")
	}
}

func TestResolveScopes_Update(t *testing.T) {
	allowed := ResolveScopes([]string{ScopeUpdate})
	if !allowed[CmdUpdate] {
		t.Error("update should be allowed by update scope")
	}
	if allowed[CmdProfileList] {
		t.Error("profile.list should not be allowed by update scope")
	}
}

func TestResolveScopes_MultipleScopes(t *testing.T) {
	allowed := ResolveScopes([]string{ScopeView, ScopeControl})
	if !allowed[CmdProfileList] {
		t.Error("profile.list should be allowed")
	}
	if !allowed[CmdProfileStart] {
		t.Error("profile.start should be allowed")
	}
	if allowed[CmdUpdate] {
		t.Error("update should not be allowed")
	}
	if allowed[CmdProfileAdd] {
		t.Error("profile.add should not be allowed")
	}
}

func TestResolveScopes_UnknownScope_EmptyAllowed(t *testing.T) {
	allowed := ResolveScopes([]string{"unknown"})
	if len(allowed) != 0 {
		t.Errorf("unknown scope should produce empty allowed set, got %v", allowed)
	}
}
