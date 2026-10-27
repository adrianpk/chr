package app

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	"forge.adrianpk.com/chr/internal/webui"
)

type webJourney struct {
	provider webViewProvider
	model    *fakeServiceQuizModel
	handler  http.Handler
	dump     string
	specs    []string
	plan     string
	ticket   string
}

// One persisted workspace crosses discovery, authoring, approved ordering,
// Board edits and read-only monitoring; separate feature fixtures cannot prove
// that their identities and scope bindings remain usable together.
func TestWebConsolidatedJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("real Git cross-feature journey is selected explicitly")
	}

	journey := newWebJourney(t)
	journey.discovery(t)
	journey.quiz(t)
	journey.authorSpecs(t)
	journey.planAndBind(t)
	journey.boardAndObserve(t)
	journey.reopen(t)
}

func newWebJourney(t *testing.T) *webJourney {
	t.Helper()
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	provider, model := webPlanFixture(t)
	now := serviceQuizClock()()
	model.responses = []quizmodel.Response{serviceQuizModelResponseWithCandidates("journey-batch", "question-1", now, []quizcore.Candidate{serviceQuizCandidate("question-2", "What verifies the output?", now)})}

	return &webJourney{provider: provider, model: model, handler: webQuizHandler(t, provider)}
}

func webJourneyPost(t *testing.T, handler http.Handler, form url.Values) {
	t.Helper()

	response := webDumpEditorPost(handler, form)
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		t.Fatal(form.Get("action"), response.Code, response.Body.String())
	}

	for _, failure := range []string{`data-write-status="validation_failed"`, `data-write-status="blocked"`, `data-write-status="stale"`} {
		if strings.Contains(response.Body.String(), failure) {
			t.Fatal(form.Get("action"), response.Body.String())
		}
	}
}

func (j *webJourney) discovery(t *testing.T) {
	t.Helper()

	for _, item := range []struct{ id, scope, body string }{{"general-one", "", "Project boundaries remain explicit."}, {"general-one", "", "Project boundaries remain explicit."}, {"general-two", "", "Retain context across retries."}, {"case-one", "named-work", "Print hello once. Verify stdout equals hello. Preserve existing files."}} {
		webJourneyPost(t, j.handler, url.Values{"action": {"dump.create"}, "family": {"dumps"}, "confirm": {"true"}, "request_id": {item.id}, "case_slug": {item.scope}, "title": {"Journey context"}, "content": {item.body}})
	}

	root := j.provider.projectContext.ProjectDir

	observed, err := core.ObserveProjectDump(t.Context(), root)
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 2 || len(observed.Scoped) != 1 {
		t.Fatal(observed, err)
	}

	j.dump = observed.Destination
	j.saveDiscovery(t)
}

func (j *webJourney) saveDiscovery(t *testing.T) {
	t.Helper()

	detail, err := j.provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "dumps", ID: j.dump})
	if err != nil || strings.Count(detail.Content, "Project boundaries remain explicit.") != 1 || strings.Contains(detail.Content, "Print hello") {
		t.Fatal(detail, err)
	}

	body := detail.Content + "\n\nKeep confirmed results visible."
	webJourneyPost(t, j.handler, url.Values{"action": {"dump.save"}, "family": {"dumps"}, "surface": {"modal"}, "confirm": {"true"}, "id": {j.dump}, "revision": {detail.Revision}, "request_id": {detail.SaveRequestID}, "content": {body}})

	after, err := j.provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "dumps", ID: j.dump})
	if err != nil || after.Content != body {
		t.Fatal(after, err)
	}
}

func (j *webJourney) quiz(t *testing.T) {
	t.Helper()

	operations := j.provider.service(t.Context())

	started, err := operations.StartQuiz(QuizStartRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}, Mode: quizcore.ModeFast})
	if err != nil {
		t.Fatal(err)
	}

	form := quizHTTPForm(started.State, "quiz.answer", "journey-answer")
	form.Set("answer", "A successful check prints hello once without changing existing files.")

	for range 2 {
		webJourneyPost(t, j.handler, form)
	}

	state, err := operations.quizRepository().LoadRuntimeState(t.Context(), operations.quizScope(quizcore.Scope{CaseSlug: "named-work"}))
	if err != nil || state.ActiveQuestion == nil || state.ActiveQuestion.ID == started.State.ActiveQuestion.ID || len(j.model.requests) != 1 {
		t.Fatal(state, err, len(j.model.requests))
	}

	webJourneyPost(t, j.handler, quizHTTPForm(state, "quiz.done", "journey-done"))

	view, err := j.provider.CatalogView(t.Context(), "quizzes")
	if err != nil || len(view.Items) != 2 || !strings.Contains(webPlanGet(j.handler, "/quizzes").Body.String(), form.Get("answer")) {
		t.Fatal(view, err)
	}
}

func (j *webJourney) authorSpecs(t *testing.T) {
	t.Helper()

	detail := webSpecDetail(t, j.provider, "")
	form := webSpecForm(detail, "journey-manual", "Manual journey contract", "Preserve context and confirm the selected output.", false)
	form.Set("case_slug", "named-work")
	webJourneyPost(t, j.handler, form)

	operations := j.provider.service(t.Context())

	documents, err := operations.ListSpecDocuments()
	if err != nil || len(documents) != 1 || documents[0].Status != "approved" || len(j.model.requests) != 1 {
		t.Fatal(documents, err)
	}

	generated, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}})
	if err != nil {
		t.Fatal(err)
	}

	j.specs = []string{documents[0].Reference, generated.Spec.RelativePath}
	for _, reference := range j.specs {
		checkWebSpecEdit(t, j.provider, j.handler, reference)
	}
}

func (j *webJourney) reopen(t *testing.T) {
	t.Helper()
	// Discard the transport handler and reopen from the same persisted workspace.
	handler := webQuizHandler(t, j.provider)
	before := webPlanFiles(t, j.provider)
	calls := len(j.model.requests)

	for _, path := range []string{"/", "/cases", "/dumps", "/quizzes", "/specs", "/plans", "/board", "/dash", "/dash/fragment"} {
		response := webPlanGet(handler, path)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "workspace") && path != "/dash/fragment" {
			t.Fatal(path, response.Code, response.Body.String())
		}
	}

	if len(j.model.requests) != calls || !reflect.DeepEqual(before, webPlanFiles(t, j.provider)) {
		t.Fatal("reopening generated content or changed retained files")
	}
}
