package app

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// TestGuidedUseCaseControlWeb submits rendered control guards independently of
// a pending stage's older revision and resumes its exact original answer.
func TestGuidedUseCaseControlWeb(t *testing.T) {
	provider, model := guidedWebFixture(t)
	handler := guidedWebHandler(t, provider)
	operations := provider.service(t.Context())
	paused, pause := pauseGuidedWebStage(t, provider, handler)
	page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(paused.ID, paused.Namespace), nil)

	if strings.Contains(page.Body.String(), "Retry this step") || strings.Contains(page.Body.String(), "Answer and continue") || len(model.requests) != 1 {
		t.Fatal("paused view exposed dispatch or inspection invoked the provider")
	}

	assertGuidedWebControlStale(t, handler, pause)

	resumed := resumeGuidedWebStage(t, provider, handler, paused, page.Body.String())
	if len(model.requests) != 3 {
		t.Fatal("resume repeated the retained answer or specification generation")
	}

	page = guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(resumed.ID, resumed.Namespace), nil)
	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", guidedWebControlForm(t, page.Body.String(), "cancel"))

	cancelled := readWebGuidedSession(t, operations, paused.ID)
	if response.Code != http.StatusSeeOther || cancelled.Guided.Status != "cancelled" || len(model.requests) != 3 {
		t.Fatal("rendered cancel advanced authoring", response.Code)
	}
}

func guidedWebControlForm(t *testing.T, body, action string) url.Values {
	t.Helper()

	for _, form := range regexp.MustCompile(`(?s)<form\b.*?</form>`).FindAllString(body, -1) {
		if !strings.Contains(form, `name="step" value="`+action+`"`) {
			continue
		}

		fields := url.Values{}
		for _, name := range []string{"surface", "action", "confirm", "id", "namespace", "revision", "native_revision", "request_id", "step", "content", "question_id", "input_digest"} {
			fields.Set(name, webWorkflowHidden(t, form, name))
		}

		return fields
	}

	t.Fatal("control form missing", action)

	return nil
}

func assertGuidedWebControlStale(t *testing.T, handler http.Handler, original url.Values) {
	t.Helper()

	for _, variant := range []string{"revision", "namespace"} {
		fields, err := url.ParseQuery(original.Encode())
		if err != nil {
			t.Fatal(err)
		}

		fields.Set("request_id", "stale-"+variant)

		if variant == "namespace" {
			fields.Set("namespace", "foreign")
		}

		response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", fields)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `data-write-status="stale"`) {
			t.Fatal("stale or foreign control accepted", variant, response.Code, response.Body.String())
		}
	}
}

func pauseGuidedWebStage(t *testing.T, provider webViewProvider, handler http.Handler) (authoringcore.Session, url.Values) {
	t.Helper()

	operations := provider.service(t.Context())
	question := captureGuidedWebInput(t, provider, handler)
	answer := guidedQuizCommand(question, "retained-answer", "answer", "Preserve the bounded scope.")

	pending, err := authoringcore.ReserveGuidedUseCaseStep(t.Context(), operations.projectContext.ProjectDir, answer)
	if err != nil {
		t.Fatal(err)
	}

	page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(pending.ID, pending.Namespace), nil)

	if strings.Contains(page.Body.String(), `name="step" value="pause"`) {
		t.Fatal("authoring recovery offered a redundant pause")
	}

	// A persisted pause from another adapter must remain resumable in the Web UI.
	pause := guidedWebCommandForm(authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: answer.GuidedUseCaseScope, OperationID: "external-pause", ExpectedRevision: pending.Revision, Action: "pause"})
	pause.Set("native_revision", "0")

	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", pause)

	paused := readWebGuidedSession(t, operations, question.ID)
	if response.Code != http.StatusSeeOther || paused.Guided.Status != "paused" || paused.Guided.PendingOperation.ID != answer.OperationID {
		t.Fatal("rendered pause lost pending input", response.Code, response.Body.String())
	}

	return paused, pause
}

func resumeGuidedWebStage(t *testing.T, provider webViewProvider, handler http.Handler, paused authoringcore.Session, body string) authoringcore.Session {
	t.Helper()

	operations := provider.service(t.Context())
	resume := guidedWebControlForm(t, body, "resume")
	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", resume)

	resumed := readWebGuidedSession(t, operations, paused.ID)
	if response.Code != http.StatusSeeOther || resumed.Guided.Stage != "preparation" || resumed.Guided.Control.StopIntent != "" {
		t.Fatal("resume lost the original answer or repeated generation", response.Code, response.Body.String())
	}

	return resumed
}
