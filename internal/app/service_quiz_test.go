package app

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func TestServiceStartQuizCreatesScopedRuntimeThroughModelBoundary(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.StartQuiz(QuizStartRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	if result.Message != "Quiz started." {
		t.Fatalf("Message = %q, want Quiz started.", result.Message)
	}

	if result.State.Status != core.QuizRuntimeStatusActive {
		t.Fatalf("State.Status = %q, want active", result.State.Status)
	}

	if result.State.Mode != core.QuizModeDeep {
		t.Fatalf("State.Mode = %q, want deep", result.State.Mode)
	}

	if result.State.CaseSlug != "blog" {
		t.Fatalf("State.CaseSlug = %q, want blog", result.State.CaseSlug)
	}

	if result.State.ActiveQuestion == nil || result.State.ActiveQuestion.ID != "question-1" {
		t.Fatalf("ActiveQuestion = %#v, want question-1", result.State.ActiveQuestion)
	}

	runtimePath := filepath.Join(projectDir, ".chr", "default", "tmp", "quiz-runtime", "blog.json")
	if content := readServiceWorkspaceFile(t, runtimePath); !strings.Contains(content, `"CaseSlug": "blog"`) {
		t.Fatalf("runtime content missing blog case:\n%s", content)
	}

	if len(model.requests) != 1 {
		t.Fatalf("model requests = %d, want 1", len(model.requests))
	}

	assertQuizModelRequestSections(t, model.requests[0], []string{
		"project",
		"quiz",
		"recent_entries",
		"current_question",
	})

	assertServiceQuizRedWrightPayload(t, model.requests[0])
}

func TestServiceAnswerQuizRecordsRawCaptureAndAdvances(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponseWithCandidates("response-1", "question-1", clock(), []core.QuizCandidate{serviceQuizCandidate("question-2", "Which constraint matters next?", clock())}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.StartQuiz(QuizStartRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
		Mode:  core.QuizModeFast,
	})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	result, err := operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  core.QuizScope{CaseSlug: "blog"},
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

	rawPath := filepath.Join(projectDir, "ops", "default", "quiz", "00000-blog-quiz-raw", "artifact.md")
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

func TestServiceAnswerQuizReusesPersistedPromptIntent(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponseWithQuestion("response-1", serviceQuizQuestion(core.QuizQuestion{
				ID:        "question-1",
				Prompt:    "What should the workflow preserve?",
				Topic:     "workflow",
				Depth:     1,
				TrunkID:   "question-1",
				Rationale: `{"kind":"quiz_question","metadata":{"follow_up_depth":2,"completion_recommended":false,"routing_state":{"phase":"follow_up"}}}`,
				CreatedAt: clock(),
			})),
			serviceQuizModelResponse("response-2", "question-2", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.StartQuiz(QuizStartRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	_, err = operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  core.QuizScope{CaseSlug: "blog"},
		Answer: "The workflow must preserve editorial review.",
	})
	if err != nil {
		t.Fatalf("AnswerQuiz returned error: %v", err)
	}

	if len(model.requests) != 2 {
		t.Fatalf("model requests = %d, want two", len(model.requests))
	}

	metadata, ok := model.requests[1].Payload.CurrentQuestion.PromptIntent["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("prompt intent = %#v, want metadata map", model.requests[1].Payload.CurrentQuestion.PromptIntent)
	}

	if metadata["follow_up_depth"] != float64(2) {
		t.Fatalf("follow_up_depth = %#v, want persisted value 2", metadata["follow_up_depth"])
	}

	rawPath := filepath.Join(projectDir, "ops", "default", "quiz", "00000-blog-quiz-raw", "artifact.md")
	rawContent := readServiceWorkspaceFile(t, rawPath)

	if !strings.Contains(rawContent, `"routing_state":{"phase":"follow_up"}`) {
		t.Fatalf("raw capture missing prompt intent:\n%s", rawContent)
	}
}

func TestQuizModelCurrentQuestionPayloadUsesActiveQuestionPromptIntent(t *testing.T) {
	payload := quizModelQuestionPayloadFromQuestion(core.QuizQuestion{
		ID:        "question-1",
		Prompt:    "What should the workflow preserve?",
		Required:  true,
		Topic:     "workflow",
		Depth:     1,
		Rationale: `{"kind":"quiz_question","metadata":{"follow_up_depth":3,"completion_recommended":false,"routing_state":{"phase":"follow_up"}}}`,
		CreatedAt: serviceQuizClock()(),
	})

	metadata, ok := payload.PromptIntent["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("prompt intent = %#v, want metadata map", payload.PromptIntent)
	}

	if metadata["follow_up_depth"] != float64(3) {
		t.Fatalf("follow_up_depth = %#v, want persisted value 3", metadata["follow_up_depth"])
	}
}

func TestNewQuizModelRequestIncludesAcceptanceReadiness(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "curated-blog-context.md"), `---
title: "Curated Blog Context"
status: "curated"
case: "blog"
---

# Curated Blog Context

A CLI user runs a publish command for an existing project. The expected success
path writes a report file and returns an accepted status.
`)

	operations := newTestService(t, projectDir)
	operations.now = serviceQuizClock()

	state, err := operations.newQuizRuntimeState(core.QuizScope{ProjectID: "service-project", CaseSlug: "blog"}, core.QuizModeDeep)
	if err != nil {
		t.Fatalf("newQuizRuntimeState returned error: %v", err)
	}

	request, err := operations.newQuizModelRequest(state)
	if err != nil {
		t.Fatalf("newQuizModelRequest returned error: %v", err)
	}

	readiness := request.Payload.AcceptanceReadiness
	if readiness == nil {
		t.Fatal("AcceptanceReadiness = nil")
	}

	if readiness.Decision != core.AcceptanceReadinessGenerate {
		t.Fatalf("decision = %q, want generate", readiness.Decision)
	}

	if len(readiness.InferredElements[core.AcceptanceReadinessSlotActor]) == 0 {
		t.Fatalf("inferred elements = %#v, want actor", readiness.InferredElements)
	}

	foundSection := false

	for _, section := range request.ContextSections {
		if section.Name == "acceptance_readiness" {
			foundSection = true

			break
		}
	}

	if !foundSection {
		t.Fatalf("context sections missing acceptance_readiness: %#v", request.ContextSections)
	}
}

func TestNewQuizModelRequestMarksNonAcceptanceContextNotApplicable(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "curated-blog-context.md"), `---
title: "Curated Blog Context"
status: "curated"
case: "blog"
---

# Curated Blog Context

This is a documentation-only architecture cleanup for a package boundary.
`)

	operations := newTestService(t, projectDir)
	operations.now = serviceQuizClock()

	state, err := operations.newQuizRuntimeState(core.QuizScope{ProjectID: "service-project", CaseSlug: "blog"}, core.QuizModeDeep)
	if err != nil {
		t.Fatalf("newQuizRuntimeState returned error: %v", err)
	}

	request, err := operations.newQuizModelRequest(state)
	if err != nil {
		t.Fatalf("newQuizModelRequest returned error: %v", err)
	}

	readiness := request.Payload.AcceptanceReadiness
	if readiness == nil {
		t.Fatal("AcceptanceReadiness = nil")
	}

	if readiness.Decision != core.AcceptanceReadinessNotApplicable {
		t.Fatalf("decision = %q, want not_applicable", readiness.Decision)
	}
}

func TestNewQuizModelRequestMarksIncompleteAcceptanceContextNeedsMoreInfo(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "curated-blog-context.md"), `---
title: "Curated Blog Context"
status: "curated"
case: "blog"
---

# Curated Blog Context

The project has an existing file and should write an output report.
`)

	operations := newTestService(t, projectDir)
	operations.now = serviceQuizClock()

	state, err := operations.newQuizRuntimeState(core.QuizScope{ProjectID: "service-project", CaseSlug: "blog"}, core.QuizModeDeep)
	if err != nil {
		t.Fatalf("newQuizRuntimeState returned error: %v", err)
	}

	request, err := operations.newQuizModelRequest(state)
	if err != nil {
		t.Fatalf("newQuizModelRequest returned error: %v", err)
	}

	readiness := request.Payload.AcceptanceReadiness
	if readiness == nil {
		t.Fatal("AcceptanceReadiness = nil")
	}

	if readiness.Decision != core.AcceptanceReadinessNeedsMoreInfo {
		t.Fatalf("decision = %q, want needs_more_info", readiness.Decision)
	}

	if !slices.Contains(readiness.Missing, core.AcceptanceReadinessSlotActor) {
		t.Fatalf("missing = %#v, want actor", readiness.Missing)
	}

	if len(readiness.InferredElements[core.AcceptanceReadinessSlotObservableOutcome]) == 0 {
		t.Fatalf("inferred elements = %#v, want observable outcome", readiness.InferredElements)
	}
}

func TestNewQuizModelRequestDoesNotResetAcceptanceSlotsAfterFirstQuestion(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "curated-blog-context.md"), `---
title: "Curated Blog Context"
status: "curated"
case: "blog"
---

# Curated Blog Context

A user sends an API request for an existing project before publication. The
success path accepts the request and writes a visible status response.
`)

	operations := newTestService(t, projectDir)
	operations.now = serviceQuizClock()

	state, err := operations.newQuizRuntimeState(core.QuizScope{ProjectID: "service-project", CaseSlug: "blog"}, core.QuizModeDeep)
	if err != nil {
		t.Fatalf("newQuizRuntimeState returned error: %v", err)
	}

	state.QuestionsAsked = 1
	state.AnswersRecorded = 1

	request, err := operations.newQuizModelRequest(state)
	if err != nil {
		t.Fatalf("newQuizModelRequest returned error: %v", err)
	}

	readiness := request.Payload.AcceptanceReadiness
	if readiness == nil {
		t.Fatal("AcceptanceReadiness = nil")
	}

	if readiness.Decision != core.AcceptanceReadinessGenerate {
		t.Fatalf("decision = %q, want generate from accumulated context", readiness.Decision)
	}

	for _, slot := range []core.AcceptanceReadinessSlot{
		core.AcceptanceReadinessSlotActor,
		core.AcceptanceReadinessSlotTrigger,
		core.AcceptanceReadinessSlotObservableOutcome,
	} {
		if slices.Contains(readiness.Missing, slot) {
			t.Fatalf("missing = %#v, should not reset %s after first question", readiness.Missing, slot)
		}
	}
}

func TestServiceFastQuizActivatesBufferedCandidatesWithoutRawCapture(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponseWithCandidates(
				"response-1",
				"question-1",
				clock(),
				[]core.QuizCandidate{
					serviceQuizCandidate("candidate-2", "What constraint matters next?", clock()),
					serviceQuizCandidate("candidate-3", "What preference matters next?", clock()),
				},
			),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	startResult, err := operations.StartQuiz(QuizStartRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
		Mode:  core.QuizModeFast,
	})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	if len(startResult.State.FastQueue) != 2 {
		t.Fatalf("Candidates = %#v, want two buffered candidates", startResult.State.FastQueue)
	}

	firstAnswer, err := operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  core.QuizScope{CaseSlug: "blog"},
		Answer: "The workflow must preserve ownership.",
	})
	if err != nil {
		t.Fatalf("AnswerQuiz returned error: %v", err)
	}

	if len(model.requests) != 1 {
		t.Fatalf("model requests = %d, want one request", len(model.requests))
	}

	if firstAnswer.State.ActiveQuestion == nil || firstAnswer.State.ActiveQuestion.ID != "candidate-2" {
		t.Fatalf("ActiveQuestion = %#v, want candidate-2", firstAnswer.State.ActiveQuestion)
	}

	if len(firstAnswer.State.FastQueue) != 1 {
		t.Fatalf("Candidates = %#v, want one remaining candidate", firstAnswer.State.FastQueue)
	}

	rawPath := filepath.Join(projectDir, "ops", "default", "quiz", "00000-blog-quiz-raw", "artifact.md")
	rawContent := readServiceWorkspaceFile(t, rawPath)

	if strings.Contains(rawContent, "candidate-2") {
		t.Fatalf("raw capture contains unactivated candidate:\n%s", rawContent)
	}
}

func TestServiceQuizUpdatesCuratedNotesIteratively(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	_, result := answerServiceQuiz(t, projectDir)

	curated := readServiceWorkspaceFile(t, filepath.Join(projectDir, "ops", "default", "quiz", "00000-blog-quiz", "artifact.md"))
	assertServiceQuizContentContains(t, curated, []string{
		`status: "curated"`,
		"## Decisions",
		"workflow: The workflow must preserve editorial review before publication.",
		"## Implementation Implications",
	})

	if strings.Contains(curated, "What should the workflow preserve?") {
		t.Fatalf("curated notes mirrored raw question:\n%s", curated)
	}

	if result.State.Status != core.QuizRuntimeStatusActive {
		t.Fatalf("status = %q, want active", result.State.Status)
	}
}

func TestServiceCompleteQuizDoesNotGenerateDraftSpec(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model, result := completeServiceQuizWithAnswer(t, projectDir)
	if result.State.Status != core.QuizRuntimeStatusDone {
		t.Fatalf("status = %q, want done", result.State.Status)
	}

	if len(model.model.requests) != 2 {
		t.Fatalf("model requests = %d, want start and next question only", len(model.model.requests))
	}

	specPath := filepath.Join(projectDir, ".chr", "default", "specs", "drafts", "blog-quiz-spec.md")
	if _, err := os.Stat(specPath); !os.IsNotExist(err) {
		t.Fatalf("draft spec path stat err = %v, want not exist", err)
	}
}

func TestServiceGenerateQuizDraftSpecUsesCuratedNotes(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model, _ := completeServiceQuizWithAnswer(t, projectDir)

	result, err := model.operations.GenerateQuizDraftSpec(QuizDraftSpecRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateQuizDraftSpec returned error: %v", err)
	}

	if result.State.Status != core.QuizRuntimeStatusDone {
		t.Fatalf("status = %q, want done", result.State.Status)
	}

	spec := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "drafts", "blog-quiz-spec.md"))
	assertServiceQuizContentContains(t, spec, []string{
		`status: "draft"`,
		`case_slug: "blog"`,
		"# Quiz Generated Spec",
		"Generated behavior from dump and quiz context.",
		"- curated quiz: ops/default/quiz/00000-blog-quiz/artifact.md",
	})

	draftRequest := model.model.requests[len(model.model.requests)-1]
	if draftRequest.Action != core.QuizModelActionDraftSpec {
		t.Fatalf("draft action = %q", draftRequest.Action)
	}

	assertQuizModelRequestSections(t, draftRequest, []string{"project", "quiz", "recent_entries", "current_question", "curated_quiz"})
}

type serviceQuizFixture struct {
	model      *fakeServiceQuizModel
	operations service
}

func answerServiceQuiz(t *testing.T, projectDir string) (serviceQuizFixture, QuizOperationResult) {
	t.Helper()

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponse("response-2", "question-2", clock()),
			serviceQuizDraftSpecResponse("response-3", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.StartQuiz(QuizStartRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	result, err := operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  core.QuizScope{CaseSlug: "blog"},
		Answer: "The workflow must preserve editorial review before publication.",
	})
	if err != nil {
		t.Fatalf("AnswerQuiz returned error: %v", err)
	}

	return serviceQuizFixture{model: model, operations: operations}, result
}

func completeServiceQuizWithAnswer(
	t *testing.T,
	projectDir string,
) (serviceQuizFixture, QuizOperationResult) {
	t.Helper()

	fixture, _ := answerServiceQuiz(t, projectDir)

	result, err := fixture.operations.CompleteQuiz(QuizCompleteRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatalf("CompleteQuiz returned error: %v", err)
	}

	return fixture, result
}

func assertServiceQuizContentContains(t *testing.T, content string, values []string) {
	t.Helper()

	for _, want := range values {
		if !strings.Contains(content, want) {
			t.Fatalf("content missing %q:\n%s", want, content)
		}
	}
}

func assertServiceQuizRedWrightPayload(t *testing.T, request core.QuizModelRequest) {
	t.Helper()

	if request.Payload.Project.Name != "Quiz Project" {
		t.Fatalf("payload project name = %q, want Quiz Project", request.Payload.Project.Name)
	}

	if request.Payload.Quiz.UseCaseID != "blog" {
		t.Fatalf("payload use case id = %q, want blog", request.Payload.Quiz.UseCaseID)
	}

	if request.Payload.OutputContract.RenderedTextField != "question" {
		t.Fatalf("payload rendered text field = %q, want question", request.Payload.OutputContract.RenderedTextField)
	}
}

func TestServiceQuizTrailRecoversAfterMaxDepthBranch(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponseWithQuestion("response-2", serviceQuizQuestion(core.QuizQuestion{
				ID:        "question-2",
				Prompt:    "What follow-up matters?",
				Topic:     "workflow",
				Depth:     2,
				ParentID:  "question-1",
				TrunkID:   "question-1",
				CreatedAt: clock(),
			})),
			serviceQuizModelResponseWithQuestion("response-3", serviceQuizQuestion(core.QuizQuestion{
				ID:        "question-3",
				Prompt:    "What final detail matters?",
				Topic:     "workflow",
				Depth:     3,
				ParentID:  "question-2",
				TrunkID:   "question-1",
				CreatedAt: clock(),
			})),
			serviceQuizModelResponse("response-4", "question-4", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.StartQuiz(QuizStartRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	if len(result.State.Trail) != 1 || result.State.Trail[0].QuestionID != "question-1" {
		t.Fatalf("trail = %#v, want question-1", result.State.Trail)
	}

	result = answerServiceQuizForTrail(t, operations, "First answer.")
	assertServiceQuizTrailIDs(t, result.State.Trail, []string{"question-1", "question-2"})

	result = answerServiceQuizForTrail(t, operations, "Second answer.")
	assertServiceQuizTrailIDs(t, result.State.Trail, []string{"question-1", "question-2", "question-3"})

	result = answerServiceQuizForTrail(t, operations, "Third answer.")
	if result.State.BranchReturnReason != "" {
		t.Fatalf("BranchReturnReason = %q, want cleared after next trunk activation", result.State.BranchReturnReason)
	}

	assertServiceQuizTrailIDs(t, result.State.Trail, []string{"question-4"})
}

func TestServiceQuizSkipReturnsToTrunkCoverage(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponseWithQuestion("response-2", serviceQuizQuestion(core.QuizQuestion{
				ID:        "question-2",
				Prompt:    "What follow-up matters?",
				Topic:     "workflow",
				Depth:     2,
				ParentID:  "question-1",
				TrunkID:   "question-1",
				CreatedAt: clock(),
			})),
			serviceQuizModelResponseWithQuestion("response-3", serviceQuizQuestion(core.QuizQuestion{
				ID:        "question-3",
				Prompt:    "What broader area should be covered next?",
				Topic:     "coverage",
				Depth:     1,
				TrunkID:   "question-3",
				CreatedAt: clock(),
			})),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.StartQuiz(QuizStartRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	result := answerServiceQuizForTrail(t, operations, "This needs a follow-up.")
	assertServiceQuizTrailIDs(t, result.State.Trail, []string{"question-1", "question-2"})

	result, err = operations.SkipQuiz(QuizSkipRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatalf("SkipQuiz returned error: %v", err)
	}

	if result.State.BranchReturnReason != "" {
		t.Fatalf("BranchReturnReason = %q, want cleared after trunk activation", result.State.BranchReturnReason)
	}

	assertServiceQuizTrailIDs(t, result.State.Trail, []string{"question-3"})

	if result.State.ActiveQuestion == nil || result.State.ActiveQuestion.Depth != 1 {
		t.Fatalf("ActiveQuestion = %#v, want trunk question", result.State.ActiveQuestion)
	}
}

func answerServiceQuizForTrail(
	t *testing.T,
	operations service,
	answer string,
) QuizOperationResult {
	t.Helper()

	result, err := operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  core.QuizScope{CaseSlug: "blog"},
		Answer: answer,
	})
	if err != nil {
		t.Fatalf("AnswerQuiz returned error: %v", err)
	}

	return result
}

func assertServiceQuizTrailIDs(t *testing.T, trail []core.QuizTrailNode, ids []string) {
	t.Helper()

	if len(trail) != len(ids) {
		t.Fatalf("trail = %#v, want ids %#v", trail, ids)
	}

	for index, id := range ids {
		if trail[index].QuestionID != id {
			t.Fatalf("trail = %#v, want ids %#v", trail, ids)
		}
	}
}

func TestServiceQuizLifecycleOperations(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponseWithCandidates("response-2", "question-2", clock(), []core.QuizCandidate{serviceQuizCandidate("question-3", "Which constraint matters next?", clock())}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	assertServiceQuizSwitchAndSkip(t, operations)
	assertServiceQuizResumeListAndComplete(t, operations)
	assertServiceQuizCancel(t, operations)
}

func TestServiceListQuizzesReturnsScopedReadModels(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	operations := newTestService(t, projectDir)
	updated := serviceQuizClock()()
	writeServiceQuizReadModelState(t, projectDir, updated)

	result, err := operations.ListQuizzes()
	if err != nil {
		t.Fatalf("ListQuizzes returned error: %v", err)
	}

	if len(result.Views) != 2 {
		t.Fatalf("Views = %#v, want project and blog quiz scopes", result.Views)
	}

	projectView := findQuizViewByScope(t, result.Views, "")
	if projectView.Status != QuizViewStatusNotStarted {
		t.Fatalf("project status = %q, want not_started", projectView.Status)
	}

	blogView := findQuizViewByScope(t, result.Views, "blog")
	if blogView.Status != string(core.QuizRuntimeStatusActive) {
		t.Fatalf("blog status = %q, want active", blogView.Status)
	}

	if blogView.Mode != core.QuizModeFast {
		t.Fatalf("blog mode = %q, want fast", blogView.Mode)
	}

	if blogView.ActiveQuestionSummary != "What should be clarified before writing?" {
		t.Fatalf("active question summary = %q", blogView.ActiveQuestionSummary)
	}

	if !strings.Contains(blogView.CuratedPreview, "Preserve editorial review.") {
		t.Fatalf("curated preview = %q, want curated decision", blogView.CuratedPreview)
	}
}

func assertServiceQuizSwitchAndSkip(t *testing.T, operations service) {
	t.Helper()

	_, err := operations.StartQuiz(QuizStartRequest{})
	if err != nil {
		t.Fatalf("StartQuiz returned error: %v", err)
	}

	modeResult, err := operations.SwitchQuizMode(QuizModeRequest{
		Mode: core.QuizModeFast,
	})
	if err != nil {
		t.Fatalf("SwitchQuizMode returned error: %v", err)
	}

	if modeResult.State.Mode != core.QuizModeFast {
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

	if completeResult.State.Status != core.QuizRuntimeStatusDone {
		t.Fatalf("complete status = %q, want done", completeResult.State.Status)
	}
}

func assertServiceQuizCancel(t *testing.T, operations service) {
	t.Helper()

	_, err := operations.StartQuiz(QuizStartRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("StartQuiz case returned error: %v", err)
	}

	cancelResult, err := operations.CancelQuiz(QuizCancelRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("CancelQuiz returned error: %v", err)
	}

	if cancelResult.State.Status != core.QuizRuntimeStatusCancelled {
		t.Fatalf("cancel status = %q, want cancelled", cancelResult.State.Status)
	}
}

func findQuizViewByScope(t *testing.T, views []QuizView, caseSlug string) QuizView {
	t.Helper()

	for _, view := range views {
		if view.Scope.CaseSlug == caseSlug {
			return view
		}
	}

	t.Fatalf("quiz view for case %q not found in %#v", caseSlug, views)

	return QuizView{}
}

func writeServiceQuizReadModelState(t *testing.T, projectDir string, updated time.Time) {
	t.Helper()

	repository := core.NewQuizArtifactRepository(projectDir)
	state := serviceQuizReadModelRuntimeState(updated)

	_, err := repository.SaveRuntimeState(t.Context(), state)
	if err != nil {
		t.Fatalf("SaveRuntimeState returned error: %v", err)
	}

	_, err = repository.WriteCuratedNotes(t.Context(), core.QuizCuratedNotes{
		Scope:     core.QuizScope{ProjectID: "service-project", CaseSlug: "blog"},
		Status:    "curated",
		Decisions: []string{"Preserve editorial review."},
		UpdatedAt: updated.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("WriteCuratedNotes returned error: %v", err)
	}
}

func serviceQuizReadModelRuntimeState(updated time.Time) core.QuizRuntimeState {
	return core.QuizRuntimeState{
		ID:              "quiz-blog",
		ProjectID:       "service-project",
		CaseSlug:        "blog",
		Status:          core.QuizRuntimeStatusActive,
		Mode:            core.QuizModeFast,
		QuestionBudget:  50,
		QuestionsAsked:  1,
		AnswersRecorded: 0,
		SkipsRecorded:   0,
		ActiveQuestion: &core.QuizQuestion{
			ID:        "question-1",
			Prompt:    "What should be clarified before writing?",
			Topic:     "clarity",
			Depth:     1,
			TrunkID:   "question-1",
			CreatedAt: updated,
		},
		Trail: []core.QuizTrailNode{
			{QuestionID: "question-1", Prompt: "What should be clarified?", Topic: "clarity", Depth: 1},
		},
		RawPath:     "ops/default/quiz/00000-blog-quiz-raw/artifact.md",
		CuratedPath: "ops/default/quiz/00000-blog-quiz/artifact.md",
		UpdatedAt:   updated,
	}
}

type fakeServiceQuizModel struct {
	requests  []core.QuizModelRequest
	responses []core.QuizModelResponse
}

//nolint:nestif // Test fake keeps action fallback behavior in one place.
func (m *fakeServiceQuizModel) InvokeQuizModel(
	_ context.Context,
	request core.QuizModelRequest,
) (core.QuizModelResponse, error) {
	m.requests = append(m.requests, request)

	if len(m.responses) == 0 {
		if request.Action == core.QuizModelActionDraftSpec || request.Action == core.QuizModelActionGenerateSpec {
			return serviceQuizDraftSpecResponse("response-fallback", request.CreatedAt), nil
		}

		if request.Action == core.QuizModelActionGeneratePlanOutline {
			return serviceQuizPlanOutlineResponse("response-fallback-outline", request.CreatedAt), nil
		}

		if request.Action == core.QuizModelActionGeneratePlanSliceDetail {
			return serviceQuizPlanSliceDetailResponse("response-fallback-slice-detail", request.CreatedAt), nil
		}

		if request.Action == core.QuizModelActionGeneratePlan {
			return serviceQuizPlanResponse("response-fallback-plan", request.CreatedAt), nil
		}

		if request.Action == core.QuizModelActionGenerateAcceptanceFeature {
			return serviceQuizAcceptanceFeatureResponse("response-fallback-feature", request.CreatedAt, core.AcceptanceFeatureGeneration{
				Status:          core.AcceptanceFeatureStatusNotApplicable,
				RejectionReason: "No acceptance feature context is available.",
			}), nil
		}

		if request.Action == core.QuizModelActionMaterializePlanTracking {
			return serviceQuizPlanTrackingResponse("response-fallback", request.CreatedAt), nil
		}

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
	writeServiceWorkspaceFile(t, core.ProjectMetadataPath(projectDir), core.RenderProjectMetadata(core.ProjectMetadata{
		Name:        "Quiz Project",
		Description: "Project metadata for quiz context.",
	}))
	writeServiceWorkspaceFile(t, core.CaseMetadataPath(projectDir, "blog"), core.RenderCaseMetadata(core.CaseMetadata{
		Name:        "Blog Publishing",
		Description: "Publish articles.",
	}))
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "project-context.md"), `---
title: "Project Context"
status: "raw"
---

# Project Context

Project-wide dump body.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "blog-context.md"), `---
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
) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		Question: core.QuizQuestion{
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

func serviceQuizDraftSpecResponse(responseID string, createdAt time.Time) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		DraftSpec: core.QuizDraftSpec{
			Scope:          core.QuizScope{ProjectID: "service-project"},
			Title:          "Quiz Generated Spec",
			Body:           "Generated behavior from dump and quiz context.",
			SourceQuizPath: "ops/default/quiz/00000-project-quiz/artifact.md",
			SourceRawPath:  "ops/default/quiz/00000-project-quiz-raw/artifact.md",
			CreatedAt:      createdAt,
		},
		RawText:   "raw draft spec",
		CreatedAt: createdAt,
	}
}

func serviceQuizDraftSpecResponseWithFeature(
	responseID string,
	createdAt time.Time,
	feature core.AcceptanceFeatureGeneration,
) core.QuizModelResponse {
	response := serviceQuizDraftSpecResponse(responseID, createdAt)
	response.DraftSpec.Feature = feature

	return response
}

func serviceQuizAcceptanceFeatureResponse(responseID string, createdAt time.Time, feature core.AcceptanceFeatureGeneration) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		Feature:   feature,
		RawText:   "raw acceptance feature",
		CreatedAt: createdAt,
	}
}

func serviceQuizStagedPlanResponses(createdAt time.Time) []core.QuizModelResponse {
	return []core.QuizModelResponse{
		serviceQuizPlanOutlineResponse("response-plan-outline", createdAt),
		serviceQuizPlanSliceDetailResponse("response-plan-slice-detail", createdAt),
	}
}

func serviceQuizPlanResponse(responseID string, createdAt time.Time) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		Plan: core.GeneratedPlan{
			Scope:     core.QuizScope{ProjectID: "service-project"},
			Title:     "Generated Implementation Plan",
			Body:      "## High-level implementation strategy\n\nImplement the generated plan in one pass.\n",
			Order:     []core.PlanOrderSlice{{ID: "slice-1", Order: 1, Title: "Foundation", Tasks: []core.PlanOrderTask{{ID: "task-t1.1", Order: 1, Title: "Implement foundation"}}}},
			CreatedAt: createdAt,
		},
		RawText:   "raw generated plan",
		CreatedAt: createdAt,
	}
}

func serviceQuizPlanOutlineResponse(responseID string, createdAt time.Time) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		PlanOutline: core.GeneratedPlanOutline{
			Scope:              core.QuizScope{ProjectID: "service-project"},
			Title:              "Generated Implementation Plan",
			Status:             "planned",
			Summary:            "Implement the planning foundation.",
			Assumptions:        []string{"Existing planning context is authoritative."},
			OpenQuestions:      []string{},
			Risks:              []string{"Generated output needs review."},
			ValidationStrategy: "Run focused and repository tests.",
			Slices: []core.GeneratedPlanOutlineSlice{
				{
					ID:                     "Slice 1",
					ShortName:              "Foundation",
					Objective:              "Implement the planning foundation.",
					Rationale:              "The generated plan needs a first executable slice.",
					Dependencies:           []string{},
					BranchName:             "feat/slice-1-foundation",
					PRTitle:                "feat(slice-1): foundation",
					ReportPath:             ".prj/reports/slices/generated-plan/slice-1-foundation.md",
					Complexity:             "medium",
					ImplementationBoundary: "Foundation behavior only.",
				},
			},
			Dependencies: []core.GeneratedPlanOutlineDependency{},
			CreatedAt:    createdAt,
		},
		RawText:   "raw generated plan outline",
		CreatedAt: createdAt,
	}
}

func serviceQuizPlanSliceDetailResponse(responseID string, createdAt time.Time) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		SliceDetail: core.GeneratedPlanSliceDetail{
			Scope:               core.QuizScope{ProjectID: "service-project"},
			SliceID:             "Slice 1",
			ShortName:           "Foundation",
			Objective:           "Implement the planning foundation.",
			ImplementationNotes: []string{"Keep the first slice focused on the planning foundation."},
			Tasks: []core.GeneratedPlanSliceTask{
				{
					ID:         "T1.1",
					Summary:    "Implement the planning foundation.",
					CommitHint: "feat(plan): implement planning foundation",
				},
			},
			AcceptanceCriteria: []string{"Foundation behavior is covered by tests."},
			TicketMetadata: core.GeneratedPlanSliceTicketMetadata{
				Title:    "Implement planning foundation",
				Synopsis: "Build the planning foundation slice.",
				Priority: "normal",
				Severity: "normal",
				Type:     "task",
				Tags:     []string{"plan"},
			},
			BranchMetadata: core.GeneratedPlanSliceBranchMetadata{
				BranchName: "feat/slice-1-foundation",
			},
			PRExpectations:         []string{"Open one PR against dev."},
			ValidationExpectations: []string{"Run focused and repository tests."},
			Dependencies:           []string{},
			NonGoals:               []string{"Do not add unrelated work."},
			CreatedAt:              createdAt,
		},
		RawText:   "raw generated plan slice detail",
		CreatedAt: createdAt,
	}
}

func serviceQuizPlanTrackingResponse(responseID string, createdAt time.Time) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		PlanTracking: core.GeneratedPlanTracking{
			Scope:     core.QuizScope{ProjectID: "service-project"},
			Title:     "Generated Plan Tracking",
			Body:      "Track the generated plan slices through tickets.",
			CreatedAt: createdAt,
			Tickets: []core.GeneratedPlanTrackingTicket{
				{
					SliceID:  "Slice 1",
					Title:    "Implement planning foundation",
					Synopsis: "Build the planning foundation slice.",
					Summary:  "Implement the foundational planning work.",
					Body:     "## Tasks\n\n- Implement the planning foundation.\n\n## Acceptance Criteria\n\n- Foundation behavior is covered by tests.",
					Priority: "normal",
					Severity: "normal",
					Type:     "task",
					Tags:     []string{"plan-tracking"},
				},
			},
		},
		RawText:   "raw plan tracking",
		CreatedAt: createdAt,
	}
}

func serviceQuizModelResponseWithQuestion(responseID string, question core.QuizQuestion) core.QuizModelResponse {
	return core.QuizModelResponse{
		ID:        responseID,
		RequestID: "filled-by-fake",
		Provider:  "fake",
		Status:    core.QuizModelResponseReady,
		Question:  question,
		RawText:   "raw model text",
		CreatedAt: question.CreatedAt,
	}
}

func serviceQuizQuestion(question core.QuizQuestion) core.QuizQuestion {
	question.Required = true

	return question
}

func serviceQuizModelResponseWithCandidates(
	responseID string,
	questionID string,
	createdAt time.Time,
	candidates []core.QuizCandidate,
) core.QuizModelResponse {
	response := serviceQuizModelResponse(responseID, questionID, createdAt)
	response.Candidates = candidates

	return response
}

func serviceQuizCandidate(
	id string,
	prompt string,
	createdAt time.Time,
) core.QuizCandidate {
	return core.QuizCandidate{
		ID:        id,
		Prompt:    prompt,
		Topic:     "workflow",
		Depth:     1,
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
	request core.QuizModelRequest,
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
