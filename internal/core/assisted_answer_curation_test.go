package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func TestBuildAssistedCuratedAnswerSummaryPreservesAnswersDecisionsAndOpenQuestions(t *testing.T) {
	sessionID := "session-1"
	now := time.Date(2026, 5, 21, 18, 0, 0, 0, time.UTC)
	events := append(
		testAssistedContextTranscriptEvents(t, sessionID),
		testAssistedAnswerCurationResultEvent(sessionID, now),
	)

	summary := BuildAssistedCuratedAnswerSummary(sessionID, events, now)

	if len(summary.Answers) != 1 {
		t.Fatalf("answer count = %d, want 1", len(summary.Answers))
	}

	if summary.Answers[0].Answer != "The assisted authoring flow." {
		t.Fatalf("answer = %q", summary.Answers[0].Answer)
	}

	if len(summary.Decisions) != 1 {
		t.Fatalf("decision count = %d, want 1", len(summary.Decisions))
	}

	if !strings.Contains(summary.Decisions[0], "Enough context was collected.") {
		t.Fatalf("decision missing summary: %#v", summary.Decisions)
	}

	if len(summary.OpenQuestions) != 1 {
		t.Fatalf("open question count = %d, want 1", len(summary.OpenQuestions))
	}

	if summary.OpenQuestions[0].QuestionID != "question-2" {
		t.Fatalf("open question = %q, want question-2", summary.OpenQuestions[0].QuestionID)
	}
}

func TestCurateAssistedAnswersWritesArtifactAndUpdatesSessionProvenance(t *testing.T) {
	dir := t.TempDir()
	repository, session, paths := setupAssistedAnswerCurationTest(t, dir)
	now := time.Date(2026, 5, 21, 19, 0, 0, 0, time.UTC)
	events := testAssistedContextTranscriptEvents(t, session.ID)
	response := testAssistedAnswerCurationModelResponse(session.ID, now.Add(time.Minute))

	result, err := CurateAssistedAnswers(context.Background(), AssistedAnswerCurationOptions{
		Session:          session,
		SessionPaths:     paths,
		TranscriptEvents: events,
		Repository:       repository,
		Adapter:          &FakeAssistedModelAdapter{Responses: []AssistedModelResponse{response}},
		Now:              func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("CurateAssistedAnswers returned error: %v", err)
	}

	content := readAssistedAnswerCurationFile(t, result.ArtifactPath)
	for _, want := range []string{
		"Curated Assisted Answers",
		"The user wants a concise assisted authoring workflow.",
		"The assisted authoring flow.",
		"question-2",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("curated artifact missing %q:\n%s", want, content)
		}
	}

	updated, err := repository.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}

	if len(updated.Outputs) != 1 {
		t.Fatalf("session output count = %d, want 1", len(updated.Outputs))
	}

	if len(updated.Provenance) != 1 {
		t.Fatalf("session provenance count = %d, want 1", len(updated.Provenance))
	}

	if result.Session.Outputs[0].ID != updated.Outputs[0].ID {
		t.Fatalf("result session output = %q, updated output = %q", result.Session.Outputs[0].ID, updated.Outputs[0].ID)
	}
}

func TestCurateAssistedAnswersRejectsArtifactCollision(t *testing.T) {
	dir := t.TempDir()
	repository, session, paths := setupAssistedAnswerCurationTest(t, dir)
	writeTestFile(t, filepath.Join(paths.SessionDir, "curated-summary.md"), "existing")

	_, err := CurateAssistedAnswers(context.Background(), AssistedAnswerCurationOptions{
		Session:          session,
		SessionPaths:     paths,
		TranscriptEvents: testAssistedContextTranscriptEvents(t, session.ID),
		Repository:       repository,
		Adapter: &FakeAssistedModelAdapter{Responses: []AssistedModelResponse{
			testAssistedAnswerCurationModelResponse(session.ID, time.Date(2026, 5, 21, 20, 0, 0, 0, time.UTC)),
		}},
		Now: func() time.Time { return time.Date(2026, 5, 21, 19, 0, 0, 0, time.UTC) },
	})
	if err == nil {
		t.Fatal("CurateAssistedAnswers returned nil error")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("collision error = %v", err)
	}
}

func TestRenderAssistedCuratedAnswerSummaryIsDeterministic(t *testing.T) {
	now := time.Date(2026, 5, 21, 21, 0, 0, 0, time.UTC)
	summary := BuildAssistedCuratedAnswerSummary(
		"session-1",
		testAssistedContextTranscriptEvents(t, "session-1"),
		now,
	)
	summary.ModelSummary = "Stable summary."

	first := RenderAssistedCuratedAnswerSummary(summary)
	second := RenderAssistedCuratedAnswerSummary(summary)

	if first != second {
		t.Fatalf("rendered summary is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func setupAssistedAnswerCurationTest(
	t *testing.T,
	dir string,
) (AuthoringSessionRepository, authoringcore.Session, AuthoringSessionPaths) {
	t.Helper()

	repository := NewAuthoringSessionRepository(dir)
	session := testAssistedContextSession(dir)

	paths, err := repository.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	return repository, session, paths
}

func testAssistedAnswerCurationModelResponse(sessionID string, createdAt time.Time) AssistedModelResponse {
	return AssistedModelResponse{
		ID:        "curation-response-1",
		RequestID: "curation-" + sessionID,
		SessionID: sessionID,
		TurnID:    "curation",
		Provider:  "fake",
		Status:    AssistedModelResponseReady,
		Question: AssistedAuthoringQuestion{
			ID:       "curation-ready",
			Prompt:   "Curation is ready.",
			Required: false,
		},
		RawText:   "The user wants a concise assisted authoring workflow.",
		CreatedAt: createdAt,
	}
}

func testAssistedAnswerCurationResultEvent(sessionID string, createdAt time.Time) AssistedTranscriptEvent {
	return AssistedTranscriptEvent{
		ID:        "turn-3-result",
		SessionID: sessionID,
		TurnID:    "turn-3",
		TurnIndex: 3,
		Type:      AssistedTranscriptEventResult,
		Intent:    AssistedAuthoringIntentCurate,
		Status:    AssistedAuthoringTurnCompleted,
		Result: AssistedAuthoringTurnResult{
			CompletionReason: AssistedAuthoringCompletionEnoughContext,
			Summary:          "Enough context was collected.",
		},
		CreatedAt: createdAt,
	}
}

func readAssistedAnswerCurationFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}
