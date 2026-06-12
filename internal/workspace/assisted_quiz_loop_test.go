package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAskNextAssistedQuizQuestionInvokesModelAndPersistsTurn(t *testing.T) {
	dir := t.TempDir()
	repository, session, paths, pkg := setupAssistedQuizLoopTest(t, dir)
	now := time.Date(2026, 5, 21, 13, 0, 0, 0, time.UTC)
	response := testAssistedQuizModelResponse("request-001", session.ID, "turn-001", now.Add(time.Minute))

	result, err := AskNextAssistedQuizQuestion(context.Background(), AssistedQuizLoopOptions{
		Config:         testAssistedQuizConfig(session),
		Session:        session,
		SessionPaths:   paths,
		ContextPackage: pkg,
		Repository:     repository,
		Adapter:        &FakeAssistedModelAdapter{Responses: []AssistedModelResponse{response}},
		Now:            func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("AskNextAssistedQuizQuestion returned error: %v", err)
	}

	if result.Turn.Status != AssistedAuthoringTurnAsked {
		t.Fatalf("turn status = %q, want asked", result.Turn.Status)
	}

	events, err := repository.ListTranscriptEvents(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("ListTranscriptEvents returned error: %v", err)
	}

	if len(events) != 1 || events[0].Type != AssistedTranscriptEventQuestion {
		t.Fatalf("transcript events = %#v, want one question event", events)
	}

	assertAssistedQuizFileContains(t, result.InvocationPaths.RequestPath, "Assisted Model Request")
	assertAssistedQuizFileContains(t, result.InvocationPaths.ResponsePath, "What is still unclear?")
}

func TestAnswerAssistedQuizQuestionPersistsAnswer(t *testing.T) {
	dir := t.TempDir()
	repository, session, _, _ := setupAssistedQuizLoopTest(t, dir)
	now := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)
	question := AssistedAuthoringQuestion{
		ID:       "question-1",
		Prompt:   "What is still unclear?",
		Required: true,
	}

	result, err := AnswerAssistedQuizQuestion(context.Background(), AssistedQuizAnswerOptions{
		SessionID:  session.ID,
		TurnID:     "turn-001",
		TurnIndex:  1,
		Intent:     AssistedAuthoringIntentDiscover,
		Question:   question,
		AnswerText: "The deployment boundary.",
		Repository: repository,
		Now:        func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("AnswerAssistedQuizQuestion returned error: %v", err)
	}

	if result.Turn.Status != AssistedAuthoringTurnAnswered {
		t.Fatalf("turn status = %q, want answered", result.Turn.Status)
	}

	events, err := repository.ListTranscriptEvents(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("ListTranscriptEvents returned error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("transcript event count = %d, want 2", len(events))
	}

	if events[1].Type != AssistedTranscriptEventAnswer {
		t.Fatalf("second event type = %q, want answer", events[1].Type)
	}

	if events[1].Answer.Text != "The deployment boundary." {
		t.Fatalf("answer text = %q", events[1].Answer.Text)
	}
}

func TestAnswerAssistedQuizQuestionCanStopOnUserRequest(t *testing.T) {
	dir := t.TempDir()
	repository, session, _, _ := setupAssistedQuizLoopTest(t, dir)

	result, err := AnswerAssistedQuizQuestion(context.Background(), AssistedQuizAnswerOptions{
		SessionID:  session.ID,
		TurnID:     "turn-001",
		TurnIndex:  1,
		Intent:     AssistedAuthoringIntentDiscover,
		Question:   AssistedAuthoringQuestion{ID: "question-1", Prompt: "Continue?", Required: true},
		AnswerText: "Stop here.",
		UserStop:   true,
		Repository: repository,
		Now:        func() time.Time { return time.Date(2026, 5, 21, 15, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("AnswerAssistedQuizQuestion returned error: %v", err)
	}

	if result.StopReason != AssistedQuizLoopStopUser {
		t.Fatalf("stop reason = %q, want user_stopped", result.StopReason)
	}
}

func TestAskNextAssistedQuizQuestionStopsAtTurnLimit(t *testing.T) {
	dir := t.TempDir()
	repository, session, paths, pkg := setupAssistedQuizLoopTest(t, dir)
	now := time.Date(2026, 5, 21, 16, 0, 0, 0, time.UTC)
	priorEvents := testAssistedContextTranscriptEvents(t, session.ID)

	result, err := AskNextAssistedQuizQuestion(context.Background(), AssistedQuizLoopOptions{
		Config: AssistedAuthoringConfig{
			SessionID:    session.ID,
			ProjectID:    session.ProjectID,
			Target:       session.Target,
			Intent:       AssistedAuthoringIntentDiscover,
			MaxTurns:     2,
			MaxFollowUps: 4,
		},
		Session:          session,
		SessionPaths:     paths,
		ContextPackage:   pkg,
		TranscriptEvents: priorEvents,
		Repository:       repository,
		Adapter:          &FakeAssistedModelAdapter{},
		Now:              func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("AskNextAssistedQuizQuestion returned error: %v", err)
	}

	if result.StopReason != AssistedQuizLoopStopTurnLimit {
		t.Fatalf("stop reason = %q, want turn_limit", result.StopReason)
	}

	if result.Turn.Result.CompletionReason != AssistedAuthoringCompletionTurnLimit {
		t.Fatalf("completion reason = %q", result.Turn.Result.CompletionReason)
	}
}

func TestAskNextAssistedQuizQuestionStopsOnBlockedModelResponse(t *testing.T) {
	dir := t.TempDir()
	repository, session, paths, pkg := setupAssistedQuizLoopTest(t, dir)
	now := time.Date(2026, 5, 21, 17, 0, 0, 0, time.UTC)
	response := AssistedModelResponse{
		ID:        "response-1",
		RequestID: "request-001",
		SessionID: session.ID,
		TurnID:    "turn-001",
		Provider:  "fake",
		Status:    AssistedModelResponseBlocked,
		Result: AssistedAuthoringTurnResult{
			BlockedReason: AssistedAuthoringBlockedMissingContext,
			Summary:       "More project context is required.",
		},
		RawText:   "blocked",
		CreatedAt: now.Add(time.Minute),
	}

	result, err := AskNextAssistedQuizQuestion(context.Background(), AssistedQuizLoopOptions{
		Config:         testAssistedQuizConfig(session),
		Session:        session,
		SessionPaths:   paths,
		ContextPackage: pkg,
		Repository:     repository,
		Adapter:        &FakeAssistedModelAdapter{Responses: []AssistedModelResponse{response}},
		Now:            func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("AskNextAssistedQuizQuestion returned error: %v", err)
	}

	if result.StopReason != AssistedQuizLoopStopBlocked {
		t.Fatalf("stop reason = %q, want blocked", result.StopReason)
	}

	if result.Turn.Status != AssistedAuthoringTurnBlocked {
		t.Fatalf("turn status = %q, want blocked", result.Turn.Status)
	}
}

func setupAssistedQuizLoopTest(
	t *testing.T,
	dir string,
) (AuthoringSessionRepository, AuthoringSession, AuthoringSessionPaths, AssistedAuthoringContextPackage) {
	t.Helper()

	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Workspace Protocol\n")
	writeTestFile(t, filepath.Join(dir, "docs", "reference", "authoring.md"), "# Authoring Reference\n")

	repository := NewAuthoringSessionRepository(dir)
	session := testAssistedContextSession(dir)

	paths, err := repository.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	pkg, err := LoadAssistedAuthoringContextPackage(AssistedAuthoringContextPackageOptions{
		ProjectDir:   dir,
		Session:      session,
		SessionPaths: paths,
		TargetPath:   filepath.Join(dir, "docs", "reference", "authoring.md"),
	})
	if err != nil {
		t.Fatalf("LoadAssistedAuthoringContextPackage returned error: %v", err)
	}

	return repository, session, paths, pkg
}

func testAssistedQuizConfig(session AuthoringSession) AssistedAuthoringConfig {
	return AssistedAuthoringConfig{
		SessionID:    session.ID,
		ProjectID:    session.ProjectID,
		Target:       session.Target,
		Intent:       AssistedAuthoringIntentDiscover,
		MaxTurns:     4,
		MaxFollowUps: 3,
	}
}

func testAssistedQuizModelResponse(
	requestID string,
	sessionID string,
	turnID string,
	createdAt time.Time,
) AssistedModelResponse {
	return AssistedModelResponse{
		ID:        "response-1",
		RequestID: requestID,
		SessionID: sessionID,
		TurnID:    turnID,
		Provider:  "fake",
		Status:    AssistedModelResponseReady,
		Question: AssistedAuthoringQuestion{
			ID:        "question-1",
			Prompt:    "What is still unclear?",
			Rationale: "The loop needs the next missing detail.",
			Required:  true,
		},
		RawText:   "Ask what is still unclear.",
		CreatedAt: createdAt,
	}
}

func assertAssistedQuizFileContains(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if !strings.Contains(string(content), want) {
		t.Fatalf("%s missing %q:\n%s", path, want, string(content))
	}
}
