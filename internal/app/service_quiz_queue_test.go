package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// A successful Fast batch survives switches and process reconstruction. Deep
// may adapt after the preserved active question; Fast consumes its queue once.
func TestServiceQuizRetainedFastBatch(t *testing.T) {
	operations, model := newRetainedQuizFixture(t)

	started, err := operations.StartQuiz(QuizStartRequest{Mode: core.QuizModeFast})
	if err != nil {
		t.Fatal(err)
	}

	initial := started.State.ActiveQuestion.ID

	deep, err := operations.SwitchQuizMode(QuizModeRequest{Mode: core.QuizModeDeep})
	if err != nil || deep.State.ActiveQuestion.ID != initial || len(deep.State.FastQueue) != 2 {
		t.Fatal(deep, err)
	}

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Retain ownership."})
	if err != nil || answered.State.ActiveQuestion.ID != "deep-follow-up" {
		t.Fatal(answered, err)
	}

	fast, err := operations.SwitchQuizMode(QuizModeRequest{Mode: core.QuizModeFast})
	if err != nil || fast.State.ActiveQuestion.ID != "deep-follow-up" {
		t.Fatal(fast, err)
	}

	checkResumedFastQueue(t, operations, model)
}

func checkResumedFastQueue(t *testing.T, operations service, model *fakeServiceQuizModel) {
	t.Helper()
	// A fresh service must use the persisted queue, not provider memory.
	restarted := newTestService(t, operations.projectContext.ProjectDir)

	restarted.quizModel = model
	for index, want := range []string{"fast-2", "fast-3", ""} {
		result, err := restarted.SkipQuiz(QuizSkipRequest{})
		if err != nil {
			t.Fatal(err)
		}

		if want == "" {
			if result.State.ActiveQuestion != nil || result.State.Status != core.QuizRuntimeStatusIdle {
				t.Fatal(result)
			}
		} else if result.State.ActiveQuestion == nil || result.State.ActiveQuestion.ID != want {
			t.Fatalf("advance %d: %#v", index, result.State)
		}
	}

	checkFastQuizExhaustion(t, restarted, model)
}

func checkFastQuizExhaustion(t *testing.T, restarted service, model *fakeServiceQuizModel) {
	t.Helper()

	resumed, err := restarted.ResumeQuiz(QuizResumeRequest{})
	if err != nil || resumed.State.ActiveQuestion != nil || len(model.requests) != 2 {
		t.Fatal("exhaustion generated another batch", resumed, model.requests, err)
	}

	observed, err := restarted.quizRepository().Observe(t.Context(), core.QuizScope{ProjectID: "service-project"})
	if err != nil || len(observed.Entries) != 4 || observed.Entries[0].Answer != "Retain ownership." {
		t.Fatal(observed, err)
	}
}

// Answer publication is independent of provider continuation. Retrying the same
// operation after a restart returns its committed receipt without another entry.
func TestServiceQuizContinuationFailure(t *testing.T) {
	operations, _ := newRetainedQuizFixture(t)
	provider := &quizContinuationFailureModel{}
	operations.quizModel = provider

	started, err := operations.StartQuiz(QuizStartRequest{})
	if err != nil {
		t.Fatal(err)
	}

	request := QuizAnswerRequest{Answer: "First paragraph.\n\nSecond paragraph.", Binding: quizStateBinding(started.State, "answer-once")}

	answered, err := operations.AnswerQuiz(request)
	if err != nil || !answered.State.ContinuationPending || answered.State.AnswersRecorded != 1 || answered.State.ActiveQuestion != nil {
		t.Fatal(answered, err)
	}

	restarted := newTestService(t, operations.projectContext.ProjectDir)
	restarted.quizModel = provider

	checkQuizAnswerReplay(t, restarted, provider, request, answered.State.Revision)

	checkQuizRecoveredContinuation(t, restarted, provider, answered.State)
}

func checkQuizRecoveredContinuation(t *testing.T, restarted service, provider *quizContinuationFailureModel, state core.QuizRuntimeState) {
	t.Helper()

	provider.recovered = true

	resumed, err := restarted.ResumeQuiz(QuizResumeRequest{Binding: quizStateBinding(state, "continue")})
	if err != nil || resumed.State.ActiveQuestion == nil || resumed.State.AnswersRecorded != 1 || resumed.State.ContinuationPending {
		t.Fatal(resumed, err)
	}

	entries, err := restarted.quizRepository().LoadRawCaptureEntries(t.Context(), core.QuizScope{ProjectID: "service-project"})
	if err != nil || len(entries) != 1 || entries[0].Answer != "First paragraph.\n\nSecond paragraph." {
		t.Fatal(entries, err)
	}
}

// Skip retries cannot consume newly active questions. Completion and cancellation
// retain prior answers, stop generation and keep distinct lifecycle statuses.
func TestServiceQuizGuardedLifecycle(t *testing.T) {
	for _, action := range []string{"done", "cancel"} {
		t.Run(action, func(t *testing.T) { checkQuizGuardedLifecycle(t, action) })
	}
}

func checkQuizGuardedLifecycle(t *testing.T, action string) {
	t.Helper()

	operations, model := newRetainedQuizFixture(t)
	model.responses[0].Candidates = append(model.responses[0].Candidates, serviceQuizCandidate("fast-4", "What final concern remains?", serviceQuizClock()()))

	started, err := operations.StartQuiz(QuizStartRequest{Mode: core.QuizModeFast})
	if err != nil {
		t.Fatal(err)
	}

	first := checkQuizSkipReplay(t, operations, started.State)

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Keep confirmed answers.", Binding: quizStateBinding(first.State, "answer")})
	if err != nil {
		t.Fatal(err)
	}

	var stopped QuizOperationResult
	if action == "done" {
		stopped, err = operations.CompleteQuiz(QuizCompleteRequest{Binding: quizStateBinding(answered.State, "stop")})
	} else {
		stopped, err = operations.CancelQuiz(QuizCancelRequest{Binding: quizStateBinding(answered.State, "stop")})
	}

	want := map[string]core.QuizRuntimeStatus{"done": core.QuizRuntimeStatusDone, "cancel": core.QuizRuntimeStatusCancelled}[action]

	if err != nil {
		t.Fatal(err)
	}

	checkStoppedQuizState(t, stopped.State, want)

	if len(model.requests) != 1 {
		t.Fatal(stopped, err)
	}

	if action == "done" {
		resumed, err := operations.ResumeQuiz(QuizResumeRequest{})
		if err != nil || resumed.State.Status != core.QuizRuntimeStatusDone || len(model.requests) != 1 {
			t.Fatal(resumed, err)
		}
	}
}

func newRetainedQuizFixture(t *testing.T) (service, *fakeServiceQuizModel) {
	t.Helper()

	project := t.TempDir()
	writeServiceQuizFixture(t, project)

	operations := newTestService(t, project)
	now := serviceQuizClock()()
	model := &fakeServiceQuizModel{responses: []core.QuizModelResponse{
		serviceQuizModelResponseWithCandidates("batch", "first", now, []core.QuizCandidate{serviceQuizCandidate("fast-2", "Which constraint applies?", now), serviceQuizCandidate("fast-3", "What remains open?", now)}),
		serviceQuizModelResponseWithQuestion("adaptive", serviceQuizQuestion(core.QuizQuestion{ID: "deep-follow-up", Prompt: "Why does ownership matter?", Topic: "workflow", Depth: 2, ParentID: "first", TrunkID: "first", CreatedAt: now})),
	}}
	operations.quizModel, operations.now = model, serviceQuizClock()

	return operations, model
}

func quizStateBinding(state core.QuizRuntimeState, operationID string) core.QuizActionBinding {
	binding := core.QuizActionBinding{SessionID: state.SessionID, ExpectedRevision: state.Revision, OperationID: operationID}
	if state.ActiveQuestion != nil {
		binding.QuestionID = state.ActiveQuestion.ID
	}

	return binding
}

type quizContinuationFailureModel struct {
	calls     int
	recovered bool
}

func (m *quizContinuationFailureModel) InvokeQuizModel(_ context.Context, request core.QuizModelRequest) (core.QuizModelResponse, error) {
	m.calls++
	if m.calls > 1 && !m.recovered {
		return core.QuizModelResponse{}, fmt.Errorf("fixture provider unavailable")
	}

	response := serviceQuizModelResponse(fmt.Sprintf("response-%d", m.calls), fmt.Sprintf("question-%d", m.calls), request.CreatedAt)
	response.RequestID = request.ID

	return response, nil
}

func checkQuizAnswerReplay(t *testing.T, restarted service, provider *quizContinuationFailureModel, request QuizAnswerRequest, revision string) {
	t.Helper()

	replayed, err := restarted.AnswerQuiz(request)
	if err != nil || replayed.State.Revision != revision || provider.calls != 2 {
		t.Fatal("retry regenerated or lost acknowledgement", replayed, err)
	}

	request.Answer = "Changed input"
	if _, err := restarted.AnswerQuiz(request); !errors.Is(err, core.ErrWorkflowConflict) {
		t.Fatal("changed retry accepted", err)
	}
}

func checkQuizSkipReplay(t *testing.T, operations service, state core.QuizRuntimeState) QuizOperationResult {
	t.Helper()

	request := QuizSkipRequest{Binding: quizStateBinding(state, "skip-once")}

	first, err := operations.SkipQuiz(request)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.SkipQuiz(request); err != nil {
		t.Fatal(err)
	}

	request.Binding.OperationID = "stale-skip"
	if _, err := operations.SkipQuiz(request); !errors.Is(err, core.ErrWorkflowConflict) {
		t.Fatal("stale skip consumed a new question", err)
	}

	return first
}

func checkStoppedQuizState(t *testing.T, state core.QuizRuntimeState, want core.QuizRuntimeStatus) {
	t.Helper()

	if state.Status != want || state.ActiveQuestion != nil || state.AnswersRecorded != 1 || state.SkipsRecorded != 1 || len(state.FastQueue) != 1 {
		t.Fatal("stopped session lost retained facts", state)
	}
}

// A provider cannot reuse a recorded question identity for a different prompt;
// old browser drafts and stale callers must never bind to a newly asked question.
func TestServiceQuizQuestionIdentity(t *testing.T) {
	operations, model := newRetainedQuizFixture(t)
	model.responses[1].Question.ID = "first"

	started, err := operations.StartQuiz(QuizStartRequest{})
	if err != nil {
		t.Fatal(err)
	}

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Keep the question identity stable.", Binding: quizStateBinding(started.State, "identity")})
	if err != nil || answered.State.ActiveQuestion != nil || !answered.State.ContinuationPending || answered.State.AnswersRecorded != 1 {
		t.Fatal("reused question was activated", answered, err)
	}
}

// Deep may surface a question already retained in Fast work. Returning to Fast
// keeps that active question once and removes its duplicate from the pending queue.
func TestServiceQuizCoveredFastQuestion(t *testing.T) {
	operations, model := newRetainedQuizFixture(t)
	model.responses[1].Question.ID = "fast-2"
	model.responses[1].Question.Prompt = "Which constraint applies?"

	if _, err := operations.StartQuiz(QuizStartRequest{Mode: core.QuizModeFast}); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.SwitchQuizMode(QuizModeRequest{Mode: core.QuizModeDeep}); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Preserve ownership."}); err != nil {
		t.Fatal(err)
	}

	resumed, err := operations.SwitchQuizMode(QuizModeRequest{Mode: core.QuizModeFast})
	if err != nil || resumed.State.ActiveQuestion.ID != "fast-2" || len(resumed.State.FastQueue) != 1 {
		t.Fatal("covered Fast question remained queued", resumed, err)
	}

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Keep the boundary explicit."})
	if err != nil || answered.State.ActiveQuestion.ID != "fast-3" || len(model.requests) != 2 {
		t.Fatal("Fast repeated a covered question", answered, err)
	}
}
