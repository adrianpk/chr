package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Event tails remain bounded and preserve structured context. A truncated
// opening line does not fabricate activity, and cancelled reads terminate.
func TestRecentWorkEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 300*1024)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	event := executioncore.AgentEvent{SchemaVersion: 1, ProjectID: "sample", ExecutionID: "one", TicketKey: "PRJ-1", Phase: executioncore.AgentEventPhaseValidationStarted, Level: executioncore.AgentEventLevelInfo, Message: "Running validation", Timestamp: time.Now().UTC()}
	for range workflowstate.PageLimit + 2 {
		if err := executioncore.AppendAgentEvent(t.Context(), path, event); err != nil {
			t.Fatal(err)
		}
	}

	observed, err := executioncore.ReadRecentAgentEvents(t.Context(), path)
	if err != nil || len(observed.Events) != workflowstate.PageLimit || len(observed.Malformed) != 0 {
		t.Fatal(len(observed.Events), observed.Malformed, err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err = executioncore.ReadRecentAgentEvents(ctx, path); err == nil {
		t.Fatal("cancelled observation continued")
	}
}
