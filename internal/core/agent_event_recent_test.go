package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Event tails remain bounded and preserve structured context. A truncated
// opening line does not fabricate activity, and cancelled reads terminate.
func TestRecentWorkEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 300*1024)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	event := AgentEvent{SchemaVersion: 1, ProjectID: "sample", ExecutionID: "one", TicketKey: "PRJ-1", Phase: AgentEventPhaseValidationStarted, Level: AgentEventLevelInfo, Message: "Running validation", Timestamp: time.Now().UTC()}
	for range workflowstate.WorkflowPageLimit + 2 {
		if err := AppendAgentEvent(t.Context(), path, event); err != nil {
			t.Fatal(err)
		}
	}

	observed, err := ReadRecentAgentEvents(t.Context(), path)
	if err != nil || len(observed.Events) != workflowstate.WorkflowPageLimit || len(observed.Malformed) != 0 {
		t.Fatal(len(observed.Events), observed.Malformed, err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err = ReadRecentAgentEvents(ctx, path); err == nil {
		t.Fatal("cancelled observation continued")
	}
}
