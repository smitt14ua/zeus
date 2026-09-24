package cmd

import (
	"testing"

	"github.com/smitt14ua/zeus/internal/protocol"
)

// Every command reachable through a scope must have a registered handler.
func TestBuildExecutor_RegistersAllScopedCommands(t *testing.T) {
	exec := buildExecutor()
	allowed := protocol.ResolveScopes(append([]string{protocol.ScopeAll}, protocol.AllScopes...))
	if !allowed[protocol.CmdSystemReboot] {
		t.Fatalf("expected %s among scoped commands", protocol.CmdSystemReboot)
	}
	for cmd := range allowed {
		if !exec.Has(cmd) {
			t.Errorf("command %q is in a scope but has no registered handler", cmd)
		}
	}
}
