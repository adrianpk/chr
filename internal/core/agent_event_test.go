package core

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestValidateAgentEventAcceptsValidEvent(t *testing.T) {
	event := validAgentEvent()

	err := executioncore.ValidateAgentEvent(event)
	if err != nil {
		t.Fatalf("ValidateAgentEvent returned error: %v", err)
	}
}

func TestValidateAgentEventRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name  string
		event executioncore.AgentEvent
		want  string
	}{
		{
			name:  "missing schema",
			event: func() executioncore.AgentEvent { event := validAgentEvent(); event.SchemaVersion = 0; return event }(),
			want:  "schema version",
		},
		{
			name: "missing timestamp",
			event: func() executioncore.AgentEvent {
				event := validAgentEvent()
				event.Timestamp = time.Time{}

				return event
			}(),
			want: "timestamp",
		},
		{
			name:  "missing project",
			event: func() executioncore.AgentEvent { event := validAgentEvent(); event.ProjectID = ""; return event }(),
			want:  "project id",
		},
		{
			name:  "invalid phase",
			event: func() executioncore.AgentEvent { event := validAgentEvent(); event.Phase = "unknown"; return event }(),
			want:  "phase",
		},
		{
			name:  "invalid level",
			event: func() executioncore.AgentEvent { event := validAgentEvent(); event.Level = "debug"; return event }(),
			want:  "level",
		},
		{
			name:  "missing message",
			event: func() executioncore.AgentEvent { event := validAgentEvent(); event.Message = " "; return event }(),
			want:  "message",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := executioncore.ValidateAgentEvent(test.event)
			if err == nil {
				t.Fatal("ValidateAgentEvent returned nil error")
			}

			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateAgentEvent error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestFileAgentEventSinkAppendsAndReadsEventsInOrder(t *testing.T) {
	path := executioncore.AgentEventStreamPath(filepath.Join(t.TempDir(), "runs", "loop-1"))
	sink := executioncore.NewFileAgentEventSink(path)

	first := validAgentEvent()
	first.LoopID = "loop-1"
	first.Phase = executioncore.AgentEventPhaseLoopStarted
	first.Message = "loop started: 2 ready tickets"

	second := validAgentEvent()
	second.LoopID = "loop-1"
	second.TicketKey = "CHR-1"
	second.Phase = executioncore.AgentEventPhaseTicketSelected
	second.Message = "CHR-1 selected"
	second.Refs = executioncore.AgentEventRefs{
		"ticket_path": ".chr/default/tickets/CHR-1.md",
	}

	if err := sink.EmitAgentEvent(context.Background(), first); err != nil {
		t.Fatalf("EmitAgentEvent first returned error: %v", err)
	}

	if err := sink.EmitAgentEvent(context.Background(), second); err != nil {
		t.Fatalf("EmitAgentEvent second returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read event stream: %v", err)
	}

	if got := strings.Count(string(content), "\n"); got != 2 {
		t.Fatalf("event stream newline count = %d, want 2", got)
	}

	stream, err := executioncore.ReadAgentEventStream(context.Background(), path)
	if err != nil {
		t.Fatalf("ReadAgentEventStream returned error: %v", err)
	}

	assertAgentEventStreamPath(t, stream, path)
	assertAgentEventStreamOrder(t, stream)
}

func TestExecutionAgentEventStreamPathUsesCanonicalEvidenceDirectory(t *testing.T) {
	record := executioncore.Record{Artifacts: executioncore.Artifacts{
		ArtifactID: "00001-ticket-run",
		RunDir:     filepath.Join(t.TempDir(), "ops", "default", "run", "00001-ticket-run"),
	}}

	if got, want := executioncore.RecordEventStreamPath(record), filepath.Join(record.Artifacts.RunDir, "evidence", "events.jsonl"); got != want {
		t.Fatalf("ExecutionAgentEventStreamPath = %q, want %q", got, want)
	}
}

func assertAgentEventStreamPath(t *testing.T, stream executioncore.AgentEventStream, path string) {
	t.Helper()

	if stream.Path != path {
		t.Fatalf("stream path = %q, want %q", stream.Path, path)
	}
}

func assertAgentEventStreamOrder(t *testing.T, stream executioncore.AgentEventStream) {
	t.Helper()

	if len(stream.Malformed) != 0 {
		t.Fatalf("Malformed = %#v, want empty", stream.Malformed)
	}

	if len(stream.Events) != 2 {
		t.Fatalf("Events length = %d, want 2", len(stream.Events))
	}

	if stream.Events[0].Phase != executioncore.AgentEventPhaseLoopStarted {
		t.Fatalf("first phase = %q, want %q", stream.Events[0].Phase, executioncore.AgentEventPhaseLoopStarted)
	}

	if stream.Events[1].TicketKey != "CHR-1" {
		t.Fatalf("second ticket = %q, want CHR-1", stream.Events[1].TicketKey)
	}

	if stream.Events[1].Refs["ticket_path"] != ".chr/default/tickets/CHR-1.md" {
		t.Fatalf("second refs = %#v, want ticket path", stream.Events[1].Refs)
	}
}

func TestReadAgentEventStreamKeepsMalformedLineDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	validLine := mustJSONLine(t, validAgentEvent())
	invalidEvent := validAgentEvent()
	invalidEvent.Message = ""
	invalidEventLine := mustJSONLine(t, invalidEvent)

	content := strings.Join([]string{
		validLine,
		"{not-json",
		invalidEventLine,
		"",
	}, "\n")

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write event stream: %v", err)
	}

	stream, err := executioncore.ReadAgentEventStream(context.Background(), path)
	if err != nil {
		t.Fatalf("ReadAgentEventStream returned error: %v", err)
	}

	if len(stream.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(stream.Events))
	}

	if len(stream.Malformed) != 2 {
		t.Fatalf("Malformed length = %d, want 2", len(stream.Malformed))
	}

	if stream.Malformed[0].Line != 2 || !strings.Contains(stream.Malformed[0].Error, "invalid character") {
		t.Fatalf("first malformed = %#v, want JSON diagnostic on line 2", stream.Malformed[0])
	}

	if stream.Malformed[1].Line != 3 || !strings.Contains(stream.Malformed[1].Error, "message") {
		t.Fatalf("second malformed = %#v, want validation diagnostic on line 3", stream.Malformed[1])
	}
}

func TestReadAgentEventStreamReturnsEmptyForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "events.jsonl")

	stream, err := executioncore.ReadAgentEventStream(context.Background(), path)
	if err != nil {
		t.Fatalf("ReadAgentEventStream returned error: %v", err)
	}

	if stream.Path != path {
		t.Fatalf("stream path = %q, want %q", stream.Path, path)
	}

	if len(stream.Events) != 0 || len(stream.Malformed) != 0 {
		t.Fatalf("stream = %#v, want empty", stream)
	}
}

func validAgentEvent() executioncore.AgentEvent {
	return executioncore.AgentEvent{
		SchemaVersion: executioncore.AgentEventSchemaVersion,
		Timestamp:     time.Date(2026, 6, 18, 11, 32, 1, 0, time.UTC),
		ProjectID:     "project-1",
		Phase:         executioncore.AgentEventPhaseWorkerStarted,
		Level:         executioncore.AgentEventLevelInfo,
		Message:       "worker started",
	}
}

func mustJSONLine(t *testing.T, event executioncore.AgentEvent) string {
	t.Helper()

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	return string(data)
}
