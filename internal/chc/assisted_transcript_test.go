package chc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAssistedTranscriptEventsFromTurnCreatesExpectedEvents(t *testing.T) {
	turn := validAssistedAuthoringTurn(AssistedAuthoringTurnAnswered)

	events, err := AssistedTranscriptEventsFromTurn(turn)
	if err != nil {
		t.Fatalf("AssistedTranscriptEventsFromTurn returned error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}

	if events[0].Type != AssistedTranscriptEventQuestion {
		t.Fatalf("events[0].Type = %q, want question", events[0].Type)
	}

	if events[1].Type != AssistedTranscriptEventAnswer {
		t.Fatalf("events[1].Type = %q, want answer", events[1].Type)
	}
}

func TestRenderAndParseAssistedTranscriptEvent(t *testing.T) {
	event := validAssistedTranscriptEvent(AssistedTranscriptEventAnswer)

	rendered, err := RenderAssistedTranscriptEvent(event)
	if err != nil {
		t.Fatalf("RenderAssistedTranscriptEvent returned error: %v", err)
	}

	events, err := ParseAssistedTranscriptEvents("# Transcript\n\n" + rendered)
	if err != nil {
		t.Fatalf("ParseAssistedTranscriptEvents returned error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}

	if events[0].ID != event.ID {
		t.Fatalf("ID = %q, want %q", events[0].ID, event.ID)
	}

	if events[0].Answer.Text != event.Answer.Text {
		t.Fatalf("Answer.Text = %q, want %q", events[0].Answer.Text, event.Answer.Text)
	}
}

func TestValidateAssistedTranscriptEventRejectsInvalidEvent(t *testing.T) {
	for _, test := range invalidAssistedTranscriptEventTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAssistedTranscriptEvent(test.event)
			if err == nil {
				t.Fatal("ValidateAssistedTranscriptEvent returned nil error")
			}
		})
	}
}

func invalidAssistedTranscriptEventTests() []struct {
	name  string
	event AssistedTranscriptEvent
} {
	return []struct {
		name  string
		event AssistedTranscriptEvent
	}{
		{name: "id", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.ID = "" })},
		{name: "session id", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.SessionID = "" })},
		{name: "turn id", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.TurnID = "" })},
		{name: "turn index", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.TurnIndex = 0 })},
		{name: "type", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.Type = AssistedTranscriptEventType("unknown") })},
		{name: "intent", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.Intent = AssistedAuthoringIntent("unknown") })},
		{name: "status", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.Status = AssistedAuthoringTurnStatus("unknown") })},
		{name: "created at", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.CreatedAt = time.Time{} })},
		{name: "answer payload", event: assistedTranscriptEventWith(func(event *AssistedTranscriptEvent) { event.Answer.Text = "" })},
	}
}

func TestAuthoringSessionRepositoryAppendsTranscriptEvents(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAssistedTranscriptSession()

	_, err := repository.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	events, err := AssistedTranscriptEventsFromTurn(validAssistedAuthoringTurn(AssistedAuthoringTurnAnswered))
	if err != nil {
		t.Fatalf("AssistedTranscriptEventsFromTurn returned error: %v", err)
	}

	err = repository.AppendTranscriptEvents(ctx, session.ID, events)
	if err != nil {
		t.Fatalf("AppendTranscriptEvents returned error: %v", err)
	}

	loaded, err := repository.ListTranscriptEvents(ctx, session.ID)
	if err != nil {
		t.Fatalf("ListTranscriptEvents returned error: %v", err)
	}

	if len(loaded) != len(events) {
		t.Fatalf("len(loaded) = %d, want %d", len(loaded), len(events))
	}
}

func TestAuthoringSessionRepositoryDoesNotRewriteSessionWhenAppendingTranscript(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAssistedTranscriptSession()

	paths, err := repository.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	before := readTestFile(t, paths.SessionPath)
	events := []AssistedTranscriptEvent{validAssistedTranscriptEvent(AssistedTranscriptEventQuestion)}

	err = repository.AppendTranscriptEvents(ctx, session.ID, events)
	if err != nil {
		t.Fatalf("AppendTranscriptEvents returned error: %v", err)
	}

	after := readTestFile(t, paths.SessionPath)
	if after != before {
		t.Fatal("session.md changed after appending transcript events")
	}
}

func TestAuthoringSessionRepositoryRejectsMissingTranscriptSession(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	events := []AssistedTranscriptEvent{validAssistedTranscriptEvent(AssistedTranscriptEventQuestion)}

	err := repository.AppendTranscriptEvents(ctx, "missing", events)
	if !errors.Is(err, ErrAuthoringSessionNotFound) {
		t.Fatalf("AppendTranscriptEvents error = %v, want ErrAuthoringSessionNotFound", err)
	}
}

func validAssistedTranscriptEvent(eventType AssistedTranscriptEventType) AssistedTranscriptEvent {
	turn := validAssistedAuthoringTurn(AssistedAuthoringTurnAnswered)

	events, err := AssistedTranscriptEventsFromTurn(turn)
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		if event.Type == eventType {
			return event
		}
	}

	return events[0]
}

func assistedTranscriptEventWith(update func(*AssistedTranscriptEvent)) AssistedTranscriptEvent {
	event := validAssistedTranscriptEvent(AssistedTranscriptEventAnswer)
	update(&event)

	return event
}

func validAssistedTranscriptSession() AuthoringSession {
	now := time.Date(2026, 5, 21, 17, 0, 0, 0, time.UTC)

	return AuthoringSession{
		ID:        "auth-1",
		ProjectID: "project-1",
		Status:    AuthoringSessionStatusDraft,
		Mode:      AuthoringSessionModeAssisted,
		Target:    AuthoringTargetSpec,
		Topic:     "Retry behavior",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	return string(content)
}

func TestParseAssistedTranscriptEventsIgnoresEmptyTranscript(t *testing.T) {
	events, err := ParseAssistedTranscriptEvents("# Transcript\n\n")
	if err != nil {
		t.Fatalf("ParseAssistedTranscriptEvents returned error: %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("len(events) = %d, want 0", len(events))
	}
}

func TestRenderAssistedTranscriptEventIncludesReadableSections(t *testing.T) {
	event := validAssistedTranscriptEvent(AssistedTranscriptEventAnswer)

	rendered, err := RenderAssistedTranscriptEvent(event)
	if err != nil {
		t.Fatalf("RenderAssistedTranscriptEvent returned error: %v", err)
	}

	for _, want := range []string{"### Question", "### Answer", "What should this spec capture?", "It should capture the retry behavior."} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered event does not contain %q:\n%s", want, rendered)
		}
	}
}
