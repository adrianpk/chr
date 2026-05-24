package chcapp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceStartQuizCreatesScopedRuntimeThroughModelBoundary(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.StartQuiz(QuizStartRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	if result.Message != "Quiz started." {
		t.Fatalf("Message = %q, want Quiz started.", result.Message)
	}

	if result.State.Status != chc.QuizRuntimeStatusActive {
		t.Fatalf("State.Status = %q, want active", result.State.Status)
	}

	if result.State.Mode != chc.QuizModeDeep {
		t.Fatalf("State.Mode = %q, want deep", result.State.Mode)
	}

	if result.State.CaseSlug != "blog" {
		t.Fatalf("State.CaseSlug = %q, want blog", result.State.CaseSlug)
	}

	if result.State.ActiveQuestion == nil || result.State.ActiveQuestion.ID != "question-1" {
		t.Fatalf("ActiveQuestion = %#v, want question-1", result.State.ActiveQuestion)
	}

	runtimePath := filepath.Join(projectDir, ".chc", "tmp", "quiz-runtime", "blog.json")
	if content := readServiceWorkspaceFile(t, runtimePath); !strings.Contains(content, `"CaseSlug": "blog"`) {
		t.Fatalf("runtime content missing blog case:\n%s", content)
	}

	if len(model.requests) != 1 {
		t.Fatalf("model requests = %d, want 1", len(model.requests))
	}

	assertQuizModelRequestSections(t, model.requests[0], []string{
		"agent_protocol",
		"project",
		"use_case",
		"project_dump",
		"use_case_dump",
	})
}

func TestServiceAnswerQuizRecordsRawCaptureAndAdvances(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponse("response-2", "question-2", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.StartQuiz(QuizStartRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
		Mode:  chc.QuizModeFast,
	})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	result, err := operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  chc.QuizScope{CaseSlug: "blog"},
		Answer: "The authoring flow needs editorial review before publication.",
	})
	if err != nil {
		t.Fatalf("AnswerQuiz returned error: %v", err)
	}

	if result.State.AnswersRecorded != 1 {
		t.Fatalf("AnswersRecorded = %d, want 1", result.State.AnswersRecorded)
	}

	if result.State.QuestionsAsked != 2 {
		t.Fatalf("QuestionsAsked = %d, want 2", result.State.QuestionsAsked)
	}

	if result.State.ActiveQuestion == nil || result.State.ActiveQuestion.ID != "question-2" {
		t.Fatalf("ActiveQuestion = %#v, want question-2", result.State.ActiveQuestion)
	}

	rawPath := filepath.Join(projectDir, ".chc", "quizzes", "blog-quiz-raw.md")
	rawContent := readServiceWorkspaceFile(t, rawPath)

	for _, want := range []string{
		"question-1",
		"The authoring flow needs editorial review before publication.",
		"status: answered",
	} {
		if !strings.Contains(rawContent, want) {
			t.Fatalf("raw capture missing %q:\n%s", want, rawContent)
		}
	}
}

func TestServiceQuizLifecycleOperations(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponse("response-2", "question-2", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	assertServiceQuizSwitchAndSkip(t, operations)
	assertServiceQuizResumeListAndComplete(t, operations)
	assertServiceQuizCancel(t, operations)
}

func assertServiceQuizSwitchAndSkip(t *testing.T, operations service) {
	t.Helper()

	_, err := operations.StartQuiz(QuizStartRequest{})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	modeResult, err := operations.SwitchQuizMode(QuizModeRequest{
		Mode: chc.QuizModeFast,
	})
	if err != nil {
		t.Fatalf("SwitchQuizMode returned error: %v", err)
	}

	if modeResult.State.Mode != chc.QuizModeFast {
		t.Fatalf("Mode = %q, want fast", modeResult.State.Mode)
	}

	skipResult, err := operations.SkipQuiz(QuizSkipRequest{})
	if err != nil {
		t.Fatalf("SkipQuiz returned error: %v", err)
	}

	if skipResult.State.SkipsRecorded != 1 {
		t.Fatalf("SkipsRecorded = %d, want 1", skipResult.State.SkipsRecorded)
	}
}

func assertServiceQuizResumeListAndComplete(t *testing.T, operations service) {
	t.Helper()

	resumeResult, err := operations.ResumeQuiz(QuizResumeRequest{})
	if err != nil {
		t.Fatalf("ResumeQuiz returned error: %v", err)
	}

	if resumeResult.State.QuestionsAsked != 2 {
		t.Fatalf("QuestionsAsked = %d, want 2", resumeResult.State.QuestionsAsked)
	}

	listResult, err := operations.ListQuizzes()
	if err != nil {
		t.Fatalf("ListQuizzes returned error: %v", err)
	}

	if len(listResult.States) != 1 {
		t.Fatalf("States = %#v, want one runtime state", listResult.States)
	}

	completeResult, err := operations.CompleteQuiz(QuizCompleteRequest{})
	if err != nil {
		t.Fatalf("CompleteQuiz returned error: %v", err)
	}

	if completeResult.State.Status != chc.QuizRuntimeStatusDone {
		t.Fatalf("complete status = %q, want done", completeResult.State.Status)
	}
}

func assertServiceQuizCancel(t *testing.T, operations service) {
	t.Helper()

	_, err := operations.StartQuiz(QuizStartRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("StartQuiz case returned error: %v", err)
	}

	cancelResult, err := operations.CancelQuiz(QuizCancelRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("CancelQuiz returned error: %v", err)
	}

	if cancelResult.State.Status != chc.QuizRuntimeStatusCancelled {
		t.Fatalf("cancel status = %q, want cancelled", cancelResult.State.Status)
	}
}

type fakeServiceQuizModel struct {
	requests  []chc.QuizModelRequest
	responses []chc.QuizModelResponse
}

func (m *fakeServiceQuizModel) InvokeQuizModel(
	_ context.Context,
	request chc.QuizModelRequest,
) (chc.QuizModelResponse, error) {
	m.requests = append(m.requests, request)

	if len(m.responses) == 0 {
		return serviceQuizModelResponse("response-fallback", "question-fallback", request.CreatedAt), nil
	}

	response := m.responses[0]
	m.responses = m.responses[1:]
	response.RequestID = request.ID

	return response, nil
}

func writeServiceQuizFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "AGENTS.md"), "Agent protocol context.\n")
	writeServiceWorkspaceFile(t, chc.ProjectMetadataPath(projectDir), chc.RenderProjectMetadata(chc.ProjectMetadata{
		Name:        "Quiz Project",
		Description: "Project metadata for quiz context.",
	}))
	writeServiceWorkspaceFile(t, chc.CaseMetadataPath(projectDir, "blog"), chc.RenderCaseMetadata(chc.CaseMetadata{
		Name:        "Blog Publishing",
		Description: "Publish articles.",
	}))
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "project-context.md"), `---
title: "Project Context"
status: "raw"
---

# Project Context

Project-wide dump body.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "blog-context.md"), `---
title: "Blog Context"
status: "raw"
case: "blog"
---

# Blog Context

Use-case dump body.
`)
}

func serviceQuizModelResponse(
	responseID string,
	questionID string,
	createdAt time.Time,
) chc.QuizModelResponse {
	return chc.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    chc.QuizModelResponseReady,
		Question: chc.QuizQuestion{
			ID:        questionID,
			Prompt:    "What should the workflow preserve?",
			Required:  true,
			Topic:     "workflow",
			Depth:     1,
			TrunkID:   questionID,
			CreatedAt: createdAt,
		},
		RawText:   "raw model text",
		CreatedAt: createdAt,
	}
}

func serviceQuizClock() func() time.Time {
	return func() time.Time {
		return time.Date(2026, 5, 24, 8, 0, 0, 0, time.UTC)
	}
}

func assertQuizModelRequestSections(
	t *testing.T,
	request chc.QuizModelRequest,
	names []string,
) {
	t.Helper()

	seen := map[string]bool{}
	for _, section := range request.ContextSections {
		seen[section.Name] = true
	}

	for _, name := range names {
		if !seen[name] {
			t.Fatalf("context section %q missing from %#v", name, request.ContextSections)
		}
	}
}
