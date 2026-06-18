package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSummarizeAgentEventStreamKeepsLatestKnownState(t *testing.T) {
	started := validAgentEvent()
	started.LoopID = "loop-1"
	started.Phase = AgentEventPhaseLoopStarted
	started.Message = "loop started"

	blocked := validAgentEvent()
	blocked.Timestamp = started.Timestamp.Add(time.Minute)
	blocked.LoopID = "loop-1"
	blocked.TicketKey = "CHC-1"
	blocked.ExecutionID = "exec-1"
	blocked.Phase = AgentEventPhaseWorkBlocked
	blocked.Level = AgentEventLevelWarn
	blocked.Message = "CHC-1 blocked"

	summary := SummarizeAgentEventStream(AgentEventStream{
		Path:   ".chc/user/runtime/runs/loop-1/events.jsonl",
		Events: []AgentEvent{started, blocked},
	})

	if summary.Status != AgentEventReadbackStatusBlocked {
		t.Fatalf("Status = %q, want blocked", summary.Status)
	}

	if summary.LatestLoopID != "loop-1" {
		t.Fatalf("LatestLoopID = %q, want loop-1", summary.LatestLoopID)
	}

	if summary.LatestTicketKey != "CHC-1" {
		t.Fatalf("LatestTicketKey = %q, want CHC-1", summary.LatestTicketKey)
	}

	if summary.LatestExecutionID != "exec-1" {
		t.Fatalf("LatestExecutionID = %q, want exec-1", summary.LatestExecutionID)
	}

	if summary.LatestPhase != AgentEventPhaseWorkBlocked {
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

	stream, err := ReadAgentEventStream(context.Background(), path)
	if err != nil {
		t.Fatalf("ReadAgentEventStream returned error: %v", err)
	}

	summary := SummarizeAgentEventStream(stream)
	if !summary.Degraded {
		t.Fatal("Degraded = false, want true")
	}

	if summary.MalformedCount != 1 {
		t.Fatalf("MalformedCount = %d, want 1", summary.MalformedCount)
	}

	if summary.Status != AgentEventReadbackStatusRunning {
		t.Fatalf("Status = %q, want running from latest valid event", summary.Status)
	}
}

func TestAgentEventReadbackStatusForEventMapsTerminalPhases(t *testing.T) {
	tests := []struct {
		name  string
		phase AgentEventPhase
		level AgentEventLevel
		want  AgentEventReadbackStatus
	}{
		{name: "loop succeeded", phase: AgentEventPhaseLoopFinished, level: AgentEventLevelInfo, want: AgentEventReadbackStatusSucceeded},
		{name: "loop blocked", phase: AgentEventPhaseLoopFinished, level: AgentEventLevelWarn, want: AgentEventReadbackStatusBlocked},
		{name: "loop failed", phase: AgentEventPhaseLoopFinished, level: AgentEventLevelError, want: AgentEventReadbackStatusFailed},
		{name: "work blocked", phase: AgentEventPhaseWorkBlocked, level: AgentEventLevelWarn, want: AgentEventReadbackStatusBlocked},
		{name: "work failed", phase: AgentEventPhaseWorkFailed, level: AgentEventLevelError, want: AgentEventReadbackStatusFailed},
		{name: "work stopped", phase: AgentEventPhaseWorkStopped, level: AgentEventLevelWarn, want: AgentEventReadbackStatusStopped},
		{name: "validation failed", phase: AgentEventPhaseValidationFinished, level: AgentEventLevelError, want: AgentEventReadbackStatusFailed},
		{name: "validation passed still running", phase: AgentEventPhaseValidationFinished, level: AgentEventLevelInfo, want: AgentEventReadbackStatusRunning},
		{name: "ticket updated", phase: AgentEventPhaseTicketStatusUpdated, level: AgentEventLevelInfo, want: AgentEventReadbackStatusSucceeded},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := validAgentEvent()
			event.Phase = test.phase
			event.Level = test.level

			got := AgentEventReadbackStatusForEvent(event)
			if got != test.want {
				t.Fatalf("AgentEventReadbackStatusForEvent() = %q, want %q", got, test.want)
			}
		})
	}
}
