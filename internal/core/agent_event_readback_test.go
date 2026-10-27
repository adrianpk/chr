package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestSummarizeAgentEventStreamKeepsLatestKnownState(t *testing.T) {
	started := validAgentEvent()
	started.LoopID = "loop-1"
	started.Phase = executioncore.AgentEventPhaseLoopStarted
	started.Message = "loop started"

	blocked := validAgentEvent()
	blocked.Timestamp = started.Timestamp.Add(time.Minute)
	blocked.LoopID = "loop-1"
	blocked.TicketKey = "CHR-1"
	blocked.ExecutionID = "exec-1"
	blocked.Phase = executioncore.AgentEventPhaseWorkBlocked
	blocked.Level = executioncore.AgentEventLevelWarn
	blocked.Message = "CHR-1 blocked"

	summary := executioncore.SummarizeAgentEventStream(executioncore.AgentEventStream{
		Path:   ".chr/default/runtime/runs/loop-1/events.jsonl",
		Events: []executioncore.AgentEvent{started, blocked},
	})

	if summary.Status != executioncore.AgentEventReadbackStatusBlocked {
		t.Fatalf("Status = %q, want blocked", summary.Status)
	}

	if summary.LatestLoopID != "loop-1" {
		t.Fatalf("LatestLoopID = %q, want loop-1", summary.LatestLoopID)
	}

	if summary.LatestTicketKey != "CHR-1" {
		t.Fatalf("LatestTicketKey = %q, want CHR-1", summary.LatestTicketKey)
	}

	if summary.LatestExecutionID != "exec-1" {
		t.Fatalf("LatestExecutionID = %q, want exec-1", summary.LatestExecutionID)
	}

	if summary.LatestPhase != executioncore.AgentEventPhaseWorkBlocked {
		t.Fatalf("LatestPhase = %q, want work_blocked", summary.LatestPhase)
	}
}

func TestSummarizeAgentEventStreamMarksMalformedReadbackDegraded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	validLine := mustJSONLine(t, validAgentEvent())
	content := strings.Join([]string{validLine, "{not-json"}, "\n")

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write event stream: %v", err)
	}

	stream, err := executioncore.ReadAgentEventStream(context.Background(), path)
	if err != nil {
		t.Fatalf("ReadAgentEventStream returned error: %v", err)
	}

	summary := executioncore.SummarizeAgentEventStream(stream)
	if !summary.Degraded {
		t.Fatal("Degraded = false, want true")
	}

	if summary.MalformedCount != 1 {
		t.Fatalf("MalformedCount = %d, want 1", summary.MalformedCount)
	}

	if summary.Status != executioncore.AgentEventReadbackStatusRunning {
		t.Fatalf("Status = %q, want running from latest valid event", summary.Status)
	}
}

func TestAgentEventReadbackStatusForEventMapsTerminalPhases(t *testing.T) {
	tests := []struct {
		name  string
		phase executioncore.AgentEventPhase
		level executioncore.AgentEventLevel
		want  executioncore.AgentEventReadbackStatus
	}{
		{name: "loop succeeded", phase: executioncore.AgentEventPhaseLoopFinished, level: executioncore.AgentEventLevelInfo, want: executioncore.AgentEventReadbackStatusSucceeded},
		{name: "loop blocked", phase: executioncore.AgentEventPhaseLoopFinished, level: executioncore.AgentEventLevelWarn, want: executioncore.AgentEventReadbackStatusBlocked},
		{name: "loop failed", phase: executioncore.AgentEventPhaseLoopFinished, level: executioncore.AgentEventLevelError, want: executioncore.AgentEventReadbackStatusFailed},
		{name: "work blocked", phase: executioncore.AgentEventPhaseWorkBlocked, level: executioncore.AgentEventLevelWarn, want: executioncore.AgentEventReadbackStatusBlocked},
		{name: "work failed", phase: executioncore.AgentEventPhaseWorkFailed, level: executioncore.AgentEventLevelError, want: executioncore.AgentEventReadbackStatusFailed},
		{name: "work stopped", phase: executioncore.AgentEventPhaseWorkStopped, level: executioncore.AgentEventLevelWarn, want: executioncore.AgentEventReadbackStatusStopped},
		{name: "validation failed", phase: executioncore.AgentEventPhaseValidationFinished, level: executioncore.AgentEventLevelError, want: executioncore.AgentEventReadbackStatusFailed},
		{name: "validation passed still running", phase: executioncore.AgentEventPhaseValidationFinished, level: executioncore.AgentEventLevelInfo, want: executioncore.AgentEventReadbackStatusRunning},
		{name: "ticket updated", phase: executioncore.AgentEventPhaseTicketStatusUpdated, level: executioncore.AgentEventLevelInfo, want: executioncore.AgentEventReadbackStatusSucceeded},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := validAgentEvent()
			event.Phase = test.phase
			event.Level = test.level

			got := executioncore.AgentEventReadbackStatusForEvent(event)
			if got != test.want {
				t.Fatalf("AgentEventReadbackStatusForEvent() = %q, want %q", got, test.want)
			}
		})
	}
}
