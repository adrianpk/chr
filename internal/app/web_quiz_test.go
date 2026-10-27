package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/webui"
)

// Real HTTP forms bind the observed question; retries acknowledge exactly once
// and stale or incomplete requests cannot advance the next question.
func TestWebQuizActions(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeFast})
	if err != nil {
		t.Fatal(err)
	}

	handler := webQuizHandler(t, provider)
	request := quizHTTPForm(started.State, "quiz.answer", "answer-once")
	request.Set("answer", "A confirmed answer.\n\nWith its complete second paragraph.")

	for range 2 {
		recorder := webDumpEditorPost(handler, request)
		if recorder.Code != http.StatusOK || recorder.Header().Get("HX-Retarget") != "#quiz-surface" || !strings.Contains(recorder.Body.String(), `data-ack-question="question-1"`) {
			t.Fatal(recorder.Code, recorder.Body.String())
		}
	}

	request.Set("request_id", "stale")

	recorder := webDumpEditorPost(handler, request)
	if !strings.Contains(recorder.Body.String(), `data-quiz-write-status="stale"`) || !strings.Contains(recorder.Body.String(), "Reload current question") {
		t.Fatal("stale answer accepted", recorder.Body.String())
	}

	checkWebQuizConfirmedEntry(t, operations, request.Get("answer"))

	current, err := operations.quizRepository().LoadRuntimeState(t.Context(), operations.quizScope(quizcore.Scope{}))
	if err != nil {
		t.Fatal(err)
	}

	checkWebQuizBindings(t, handler, current)
}

func checkWebQuizBindings(t *testing.T, handler http.Handler, state quizcore.RuntimeState) {
	t.Helper()

	for _, missing := range []string{"revision", "request_id", "session_id", "question_id"} {
		form := quizHTTPForm(state, "quiz.skip", "missing-"+missing)
		form.Del(missing)

		recorder := webDumpEditorPost(handler, form)
		if strings.Contains(recorder.Body.String(), `data-ack-action="skip"`) {
			t.Fatal("unbound question accepted", missing)
		}
	}
}

// The answering view reads canonical entries, excludes skips and orders equal
// timestamps by stable entry identity rather than source file ordering.
func TestWebQuizHistory(t *testing.T) {
	now := serviceQuizClock()()
	entries := []quizcore.Entry{{ID: "a", Status: quizcore.EntryStatusAnswered, AnsweredAt: now}, {ID: "c", Status: quizcore.EntryStatusSkipped, AnsweredAt: now}, {ID: "b", Status: quizcore.EntryStatusAnswered, AnsweredAt: now}}

	history := webQuizHistory(entries)
	if len(history) != 2 || history[0].ID != "b" || history[1].ID != "a" {
		t.Fatal(history)
	}
}

// Model failure after a confirmed answer renders that retained answer and a
// resumable blocked state. The error screen never implies the answer was lost.
func TestWebQuizContinuationFailure(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())
	operations.quizModel = &quizContinuationFailureModel{}
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	started, err := operations.StartQuiz(QuizStartRequest{})
	if err != nil {
		t.Fatal(err)
	}

	form := quizHTTPForm(started.State, "quiz.answer", "answer-before-failure")
	form.Set("answer", "Preserve confirmed context.")

	recorder := webDumpEditorPost(webQuizHandler(t, provider), form)
	for _, want := range []string{"Preserve confirmed context.", "Resume to retry.", `data-ack-action="answer"`, `value="quiz.resume"`} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatal("continuation failure lost confirmed state", want, recorder.Body.String())
		}
	}
}

func newWebQuizFixture(t *testing.T) (webViewProvider, *fakeServiceQuizModel) {
	t.Helper()

	provider := webTicketEditorProvider(t, false)

	operations := provider.service(t.Context())
	if _, err := core.NewProjectMetadataService(provider.projectContext.ProjectDir).Save(core.ProjectMetadata{Name: "Workflow discovery", Description: "Clarify workflow boundaries and retained context."}); err != nil {
		t.Fatal(err)
	}

	now := serviceQuizClock()()

	candidates := []quizcore.Candidate{}
	for index, prompt := range []string{"Which boundary must remain explicit?", "How should a failed request be retried?", "What confirms that the change succeeded?", "Which context must survive a restart?", "What is still unresolved?"} {
		candidates = append(candidates, serviceQuizCandidate("question-"+strconv.Itoa(index+2), prompt, now))
	}

	model := &fakeServiceQuizModel{responses: []core.QuizModelResponse{serviceQuizModelResponseWithCandidates("batch", "question-1", now, candidates)}}
	operations.quizModel, operations.now = model, serviceQuizClock()
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	return provider, model
}

func webQuizHandler(t *testing.T, provider webViewProvider) http.Handler {
	t.Helper()

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

func quizHTTPForm(state quizcore.RuntimeState, action, id string) url.Values {
	question := ""
	if state.ActiveQuestion != nil {
		question = state.ActiveQuestion.ID
	}

	return url.Values{"action": {action}, "family": {"quizzes"}, "surface": {"quiz"}, "confirm": {"true"}, "id": {state.CuratedPath}, "case_slug": {state.CaseSlug}, "session_id": {state.SessionID}, "question_id": {question}, "revision": {state.Revision}, "request_id": {id}}
}

func webQuizDetail(t *testing.T, provider webViewProvider, reference string) webui.ArtifactDetailView {
	t.Helper()

	view, err := provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "quizzes", ID: reference})
	if err != nil {
		t.Fatal(err)
	}

	return view
}

func webQuizGET(t *testing.T, handler http.Handler, reference string) string {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, quizFocusURL(reference), nil))

	if recorder.Code != http.StatusOK {
		t.Fatal(recorder.Code, recorder.Body.String())
	}

	return recorder.Body.String()
}

func checkWebQuizConfirmedEntry(t *testing.T, operations service, want string) {
	t.Helper()

	entries, err := operations.quizRepository().LoadRawCaptureEntries(t.Context(), operations.quizScope(quizcore.Scope{}))
	if err != nil || len(entries) != 1 || entries[0].Answer != want {
		t.Fatal(entries, err)
	}
}

// The focused HTTP read stays side-effect-free; the browser receives a guarded
// initial POST binding without creating a session or an answer during projection.
func TestWebQuizReadOnly(t *testing.T) {
	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	paths, err := operations.quizRepository().Paths(operations.quizScope(quizcore.Scope{}))
	if err != nil {
		t.Fatal(err)
	}

	detail := webQuizDetail(t, provider, paths.CuratedRelativePath)

	content := webQuizGET(t, webQuizHandler(t, provider), paths.CuratedRelativePath)
	if detail.Quiz.Started || !strings.Contains(content, `data-quiz-auto-start="true"`) || len(model.requests) != 0 {
		t.Fatal("GET started a quiz", detail, content)
	}

	if _, err := os.Stat(paths.RuntimePath); !os.IsNotExist(err) {
		t.Fatal("GET allocated runtime state", err)
	}
}

// Focused Start and Resume submit the same observed scope, and retries do not
// generate another batch or replace the active question.
func TestWebQuizStartResume(t *testing.T) {
	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	paths, err := operations.quizRepository().Paths(operations.quizScope(quizcore.Scope{}))
	if err != nil {
		t.Fatal(err)
	}

	detail := webQuizDetail(t, provider, paths.CuratedRelativePath)
	state := quizcore.RuntimeState{CuratedPath: detail.ID, Revision: detail.Revision}
	form := quizHTTPForm(state, "quiz.start", "focused-start")
	form.Set("mode", "fast")

	handler := webQuizHandler(t, provider)
	for range 2 {
		recorder := webDumpEditorPost(handler, form)
		if !strings.Contains(recorder.Body.String(), `data-quiz-question="question-1"`) {
			t.Fatal(recorder.Body.String())
		}
	}

	current, err := operations.quizRepository().LoadRuntimeState(t.Context(), operations.quizScope(quizcore.Scope{}))
	if err != nil {
		t.Fatal(err)
	}

	recorder := webDumpEditorPost(handler, quizHTTPForm(current, "quiz.resume", "focused-resume"))
	if !strings.Contains(recorder.Body.String(), "Quiz resumed.") || len(model.requests) != 1 {
		t.Fatal(recorder.Body.String(), model.requests)
	}
}

// The final observation supplies mode/status, and draft context distinguishes
// byte-identical scope/session records in different project roots.
func TestWebQuizObservationProjection(t *testing.T) {
	view := webQuizAnswerView(QuizView{Mode: quizcore.ModeDeep, Status: "idle"}, core.QuizObservation{Exists: true, State: quizcore.RuntimeState{Mode: quizcore.ModeFast, Status: quizcore.RuntimeStatusDone}})
	if view.Mode != "fast" || view.Status != "done" || view.Open {
		t.Fatal("projection used an older catalog state", view)
	}

	first, _ := newWebQuizFixture(t)
	second, _ := newWebQuizFixture(t)
	reference := "ops/default/quiz/00000-project-quiz/artifact.md"

	a, b := webQuizDetail(t, first, reference), webQuizDetail(t, second, reference)
	if a.Quiz.DraftContext == "" || a.Quiz.DraftContext == b.Quiz.DraftContext {
		t.Fatal("draft context is not bound to the project")
	}
}
