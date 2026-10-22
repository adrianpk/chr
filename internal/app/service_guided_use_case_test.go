package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

// TestGuidedUseCaseCaptureWeb traverses the real HTTP, shared service and Git owners.
func TestGuidedUseCaseCaptureWeb(t *testing.T) {
	if testing.Short() {
		t.Skip("real Git checkpoint fixture")
	}

	provider, _ := guidedWebFixture(t)

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	page := guidedCaptureHTTP(t, handler, http.MethodGet, "/guided", nil)
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "Develop a new use case") || !strings.Contains(page.Body.String(), "Project policy") {
		t.Fatal("start or readiness missing", page.Code, page.Body.String())
	}

	operations := provider.service(t.Context())

	start := url.Values{"action": {"guided.start"}, "surface": {"guided"}, "confirm": {"true"}, "namespace": {"default"}, "request_id": {"web-guided"}}
	for range 2 {
		response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", start)
		if response.Code != http.StatusSeeOther {
			t.Fatal("start failed", response.Code, response.Body.String())
		}
	}

	session := readWebGuidedSession(t, operations, "web-guided")
	webGuidedInput(t, handler, session, "name", "Useful delivery", http.StatusSeeOther)

	session = readWebGuidedSession(t, operations, session.ID)
	webGuidedInput(t, handler, session, "finish", "", http.StatusBadRequest)

	stale := session
	webGuidedInput(t, handler, session, "append", "First requirement.", http.StatusSeeOther)

	session = readWebGuidedSession(t, operations, session.ID)
	webGuidedInput(t, handler, stale, "append", "Stale requirement.", http.StatusBadRequest)
	webGuidedInput(t, handler, session, "append", "Second requirement.", http.StatusSeeOther)

	session = readWebGuidedSession(t, operations, session.ID)
	webGuidedInput(t, handler, session, "finish", "", http.StatusSeeOther)
	assertWebGuidedCheckpoint(t, provider, handler, session.ID)
}

func webGuidedInput(t *testing.T, handler http.Handler, session core.AuthoringSession, action, content string, status int) {
	t.Helper()

	form := url.Values{"question_id": {""}, "input_digest": {""}, "action": {"guided.input"}, "surface": {"guided"}, "confirm": {"true"}, "id": {session.ID}, "namespace": {session.Namespace}, "revision": {strconv.FormatInt(session.Revision, 10)}, "request_id": {action + "-" + strconv.FormatInt(session.Revision, 10)}, "step": {action}, "content": {content}}

	if session.Guided.Quiz != nil {
		form.Set("question_id", session.Guided.Quiz.QuestionID)
		form.Set("input_digest", session.Guided.Quiz.InputDigest)
	}

	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", form)
	if response.Code != status {
		t.Fatalf("%s: status %d, expected %d: %s", action, response.Code, status, response.Body.String())
	}
}

func readWebGuidedSession(t *testing.T, operations service, id string) core.AuthoringSession {
	t.Helper()

	result, err := operations.InspectGuidedUseCase(t.Context(), id, "default")
	if err != nil {
		t.Fatal(err)
	}

	return *result.Session
}

func assertWebGuidedCheckpoint(t *testing.T, provider webViewProvider, handler http.Handler, id string) {
	t.Helper()

	operations := provider.service(t.Context())

	result, err := operations.InspectGuidedUseCase(t.Context(), id, "default")
	if err != nil {
		t.Fatal(err)
	}

	checkpoint := result.Session.Guided.Checkpoints[0]

	path := filepath.Join(provider.projectContext.ProjectDir, filepath.FromSlash(checkpoint.Artifacts[0].Path))
	if err := os.WriteFile(path, []byte("Later replacement"), 0o600); err != nil {
		t.Fatal(err)
	}

	page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(id, "default"), nil)
	for _, text := range []string{"First requirement.", "Second requirement.", checkpoint.Commit, "Answer and continue", "Execution has not started"} {
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), text) {
			t.Fatalf("reloaded page missing %q: %d %s", text, page.Code, page.Body.String())
		}
	}

	if strings.Contains(page.Body.String(), "Later replacement") || strings.Contains(page.Body.String(), "Stale requirement.") {
		t.Fatal("page replaced the admitted capture")
	}
}

// TestGuidedUseCaseCaptureService records start replay and reports missing prerequisites.
func TestGuidedUseCaseCaptureService(t *testing.T) {
	provider := webCaseCreateProvider(t)
	operations := provider.service(t.Context())

	first, err := operations.StartGuidedUseCase(t.Context(), "service-guided", "default", "local")
	if err != nil {
		t.Fatal(err)
	}

	again, err := operations.StartGuidedUseCase(t.Context(), "service-guided", "default", "local")
	if err != nil || first.ID != again.ID || first.Revision != again.Revision {
		t.Fatalf("duplicate start: %v", err)
	}

	assertGuidedServiceReadiness(t, operations, first.ID)

	response, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{Action: "guided.start", Surface: "guided", Confirm: false, Fields: map[string]string{"request_id": "unconfirmed"}})
	if err != nil || response.Status != webui.WriteStatusBlocked {
		t.Fatalf("unconfirmed start: %v", err)
	}
}

func initializeGuidedWebGit(t *testing.T, root string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, "initial.txt"), []byte("initial fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{{"init", "-b", "dev"}, {"config", "user.name", "Web Test"}, {"config", "user.email", "web@example.invalid"}, {"add", "initial.txt"}, {"commit", "-m", "Initial fixture"}} {
		command := exec.CommandContext(t.Context(), "git", append([]string{"-C", root}, args...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, output)
		}
	}
}

func assertGuidedServiceReadiness(t *testing.T, operations service, id string) {
	t.Helper()

	view, err := operations.InspectGuidedUseCase(t.Context(), id, "default")
	if err != nil || len(view.Problems) == 0 || len(view.Policy) == 0 {
		t.Fatalf("missing readiness explanations: %v %#v", err, view)
	}

	if _, err := operations.InspectGuidedUseCase(t.Context(), id, "other"); err == nil {
		t.Fatal("foreign namespace inspected")
	}
}

func guidedCaptureHTTP(t *testing.T, handler http.Handler, method, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

func guidedWebFixture(t *testing.T) (webViewProvider, *fakeServiceQuizModel) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git guided Web fixture")
	}

	provider := webCaseCreateProvider(t)
	initializeGuidedWebGit(t, provider.projectContext.ProjectDir)

	operations := newTestService(t, provider.projectContext.ProjectDir)
	now := operations.utcNow()
	completion := serviceQuizModelResponse("complete", "complete", now)
	completion.Question.Required = false
	model := &fakeServiceQuizModel{responses: []core.QuizModelResponse{serviceQuizModelResponse("first", "question-1", now), completion}}
	operations.quizModel = model
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	return provider, model
}
