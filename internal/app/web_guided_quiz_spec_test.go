package app

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestGuidedUseCaseQuizWeb advances from capture to only the current question without read-side effects.
func TestGuidedUseCaseQuizWeb(t *testing.T) {
	provider, model := guidedWebFixture(t)
	handler := guidedWebHandler(t, provider)

	current := captureGuidedWebInput(t, provider, handler)
	if current.Guided.Stage != "quiz" || current.Guided.Quiz.QuestionID != "question-1" || len(model.requests) != 1 {
		t.Fatal("capture did not start clarification")
	}

	for range 2 {
		page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(current.ID, current.Namespace), nil)
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "Answer and continue") {
			t.Fatal("current question missing")
		}
	}

	foreign := guidedWebCommandForm(guidedQuizCommand(current, "foreign", "answer", "A foreign answer."))
	foreign.Set("question_id", "other-question")

	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", foreign)

	after := readWebGuidedSession(t, provider.service(t.Context()), current.ID)
	if response.Code != http.StatusBadRequest || after.Revision != current.Revision || len(model.requests) != 1 {
		t.Fatal("query or foreign reply advanced clarification")
	}
}

// TestGuidedUseCaseSpecWeb automatically approves the generated revision and inspects frozen full documents.
func TestGuidedUseCaseSpecWeb(t *testing.T) {
	provider, model := guidedWebFixture(t)
	handler := guidedWebHandler(t, provider)
	current := captureGuidedWebInput(t, provider, handler)
	replaceGuidedWorkingDump(t, provider.projectContext.ProjectDir, current.Guided.Checkpoints[0])

	answer := guidedWebCommandForm(guidedQuizCommand(current, "web-answer", "answer", "Retain the exact agreed behavior."))
	for range 2 {
		response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", answer)
		if response.Code != http.StatusSeeOther {
			t.Fatal("automatic approval failed", response.Code, response.Body.String())
		}
	}

	completed := readWebGuidedSession(t, provider.service(t.Context()), current.ID)
	if completed.Guided.Stage != "preparation" || completed.Guided.SpecApproval == nil || len(model.requests) != 3 {
		t.Fatal("duplicate answer repeated or failed progression")
	}

	assertGuidedWebOriginals(t, provider, handler, completed)

	if len(model.requests) != 3 {
		t.Fatal("inspection dispatched generation")
	}
}

func assertGuidedWebOriginals(t *testing.T, provider webViewProvider, handler http.Handler, completed core.AuthoringSession) {
	t.Helper()
	replaceGuidedWebDocuments(t, provider.projectContext.ProjectDir, completed)

	page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(completed.ID, completed.Namespace), nil)
	for _, text := range []string{"Deliver a bounded workflow.", "Retain the exact agreed behavior.", "Inspect committed quiz", "Inspect committed specification", "Configure native workflow version 1 before retrying work preparation.", "Execution has not started."} {
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), text) {
			t.Fatalf("original document missing %q", text)
		}
	}

	if strings.Contains(page.Body.String(), "Later replacement") || strings.Contains(page.Body.String(), "Approve specification") {
		t.Fatal("inspection replaced the original or dispatched a routine confirmation")
	}
}

// TestGuidedUseCaseSpecWebRetry sends the original pending operation after an approval acknowledgement is lost.
func TestGuidedUseCaseSpecWebRetry(t *testing.T) {
	provider, model := guidedWebFixture(t)
	operations := provider.service(t.Context())
	lost := true
	operations.afterGuidedSpec = func(phase string) error {
		if lost && phase == "approval" {
			return errors.New("lost approval response")
		}

		return nil
	}
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }
	handler := guidedWebHandler(t, provider)
	current := captureGuidedWebInput(t, provider, handler)
	webGuidedInput(t, handler, current, "answer", "Retain the original approved input.", http.StatusSeeOther)

	pending := readWebGuidedSession(t, operations, current.ID)

	assertGuidedWebRetryBinding(t, provider, pending)

	lost = false

	command, ready := guidedProgressCommand(pending)
	if !ready {
		t.Fatal("pending operation not recoverable")
	}

	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", guidedWebCommandForm(command))

	completed := readWebGuidedSession(t, operations, current.ID)
	if response.Code != http.StatusSeeOther || completed.Guided.Stage != "preparation" || len(model.requests) != 3 {
		t.Fatal("retry repeated generation or approval")
	}
}

func guidedWebHandler(t *testing.T, provider webViewProvider) http.Handler {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git guided HTTP fixture")
	}

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

func captureGuidedWebInput(t *testing.T, provider webViewProvider, handler http.Handler) core.AuthoringSession {
	t.Helper()

	operations := provider.service(t.Context())

	start := url.Values{"action": {"guided.start"}, "surface": {"guided"}, "confirm": {"true"}, "namespace": {"default"}, "request_id": {"guided-web-quiz"}}
	if response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", start); response.Code != http.StatusSeeOther {
		t.Fatal(response.Body.String())
	}

	current := readWebGuidedSession(t, operations, "guided-web-quiz")
	webGuidedInput(t, handler, current, "name", "Guided Web requirements", http.StatusSeeOther)

	current = readWebGuidedSession(t, operations, current.ID)
	webGuidedInput(t, handler, current, "append", "Deliver a bounded workflow.", http.StatusSeeOther)

	current = readWebGuidedSession(t, operations, current.ID)
	webGuidedInput(t, handler, current, "finish", "", http.StatusSeeOther)

	return readWebGuidedSession(t, operations, current.ID)
}

func guidedWebCommandForm(command core.GuidedUseCaseCommand) url.Values {
	return url.Values{"action": {"guided.input"}, "surface": {"guided"}, "confirm": {"true"}, "id": {command.RunID}, "namespace": {command.Namespace}, "request_id": {command.OperationID}, "revision": {strconv.FormatInt(command.ExpectedRevision, 10)}, "step": {command.Action}, "content": {command.Content}, "question_id": {command.QuestionID}, "input_digest": {command.InputDigest}}
}

func replaceGuidedWebDocuments(t *testing.T, root string, session core.AuthoringSession) {
	t.Helper()

	for _, checkpoint := range session.Guided.Checkpoints {
		for _, artifact := range checkpoint.Artifacts {
			if !strings.HasSuffix(artifact.Path, "/artifact.md") {
				continue
			}

			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(artifact.Path)), []byte("Later replacement"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func assertGuidedWebRetryBinding(t *testing.T, provider webViewProvider, pending core.AuthoringSession) {
	t.Helper()

	view, err := provider.GuidedUseCaseView(t.Context(), pending.ID, pending.Namespace)
	if err != nil || view.PendingAction != "spec" || pending.Guided.PendingOperation == nil {
		t.Fatal("retained retry missing", err)
	}

	if view.Revision != strconv.FormatInt(pending.Guided.PendingOperation.ExpectedRevision, 10) || view.RequestID != pending.Guided.PendingOperation.ID {
		t.Fatal("retry exchanged original identity or revision")
	}
}
