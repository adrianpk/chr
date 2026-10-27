package app

import quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/webui"
)

// Real forms create draft and directly approved documents without generation,
// acknowledge retries once and edit the exact selected manual/generated identity.
func TestWebSpecAuthoring(t *testing.T) {
	provider, model := newWebQuizFixture(t)

	handler := webQuizHandler(t, provider)
	for _, draft := range []bool{true, false} {
		checkWebSpecCreate(t, provider, handler, draft)
	}

	if len(model.requests) != 0 {
		t.Fatal("manual editor generated content")
	}

	operations := provider.service(t.Context())
	documents, err := operations.ListSpecDocuments()
	checkSpecDocumentCount(t, documents, err, 2)
	checkWebSpecCreatedStates(t, documents)

	model.responses = []quizmodel.QuizModelResponse{serviceQuizDraftSpecResponse("generated", serviceQuizClock()())}

	generated, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}})
	if err != nil {
		t.Fatal(err)
	}

	for _, reference := range []string{documents[0].Reference, generated.Spec.RelativePath} {
		checkWebSpecEdit(t, provider, handler, reference)
	}
}

func checkWebSpecEdit(t *testing.T, provider webViewProvider, handler http.Handler, reference string) {
	t.Helper()

	detail := webSpecDetail(t, provider, reference)
	if detail.Spec.New || detail.Content == "" {
		t.Fatal(detail)
	}

	form := webSpecForm(detail, "save-"+detail.Spec.RequestID, "Edited exact identity", "Corrected opening.\n\nMore retained requirements.", detail.Spec.Draft)

	response := webDumpEditorPost(handler, form)
	if response.Header().Get("HX-Retarget") != "#modal-root" {
		t.Fatal(response.Body.String())
	}

	after := webSpecDetail(t, provider, reference)
	if after.ID != reference || after.Content != form.Get("content") || after.Spec.CaseSlug != detail.Spec.CaseSlug || after.Spec.Draft != detail.Spec.Draft {
		t.Fatal(after)
	}

	form.Set("request_id", "stale-"+detail.Spec.RequestID)

	response = webDumpEditorPost(handler, form)
	if !strings.Contains(response.Body.String(), `data-spec-write-status="stale"`) || !strings.Contains(response.Body.String(), "Reload spec") {
		t.Fatal(response.Body.String())
	}
}

// Catalog and new/edit GETs are read-only, expose named choices and one body
// editor, and never present operational metadata as card content.
func TestWebSpecReadOnly(t *testing.T) {
	provider, model := newWebQuizFixture(t)
	handler := webQuizHandler(t, provider)

	for _, path := range []string{"/specs", "/specs/detail/fragment"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("HX-Request", "true")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatal(response.Code, response.Body.String())
		}

		checkWebSpecReadMarkup(t, path, response.Body.String())
	}

	documents, err := provider.service(t.Context()).ListSpecDocuments()
	if err != nil || len(documents) != 0 || len(model.requests) != 0 {
		t.Fatal(documents, model.requests, err)
	}
}

// A rejected native form returns its submitted fields in the same editor so
// validation does not discard authored text when HTMX is unavailable.
func TestWebSpecNativeFailure(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	form := webSpecForm(webSpecDetail(t, provider, ""), "native-invalid", "Keep this title", "Keep this body", true)
	form.Set("case_slug", "missing-case")

	request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()
	webQuizHandler(t, provider).ServeHTTP(response, request)

	for _, want := range []string{`id="spec-editor-form"`, "Keep this title", "Keep this body"} {
		if !strings.Contains(response.Body.String(), want) {
			t.Fatal(want, response.Body.String())
		}
	}
}

func webSpecDetail(t *testing.T, provider webViewProvider, reference string) webui.ArtifactDetailView {
	t.Helper()

	detail, err := provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "specs", ID: reference})
	if err != nil {
		t.Fatal(err)
	}

	return detail
}

func webSpecForm(detail webui.ArtifactDetailView, id, title, body string, draft bool) url.Values {
	action := "spec.save"
	if detail.Spec.New {
		action = "spec.create"
	}

	form := url.Values{"action": {action}, "family": {"specs"}, "surface": {"spec"}, "modal": {"true"}, "confirm": {"true"}, "id": {detail.ID}, "revision": {detail.Revision}, "request_id": {id}, "case_slug": {detail.Spec.CaseSlug}, "title": {title}, "content": {body}}
	if draft {
		form.Set("draft", "true")
	}

	return form
}

func checkWebSpecCreate(t *testing.T, provider webViewProvider, handler http.Handler, draft bool) {
	t.Helper()

	detail := webSpecDetail(t, provider, "")

	form := webSpecForm(detail, "create-"+strconv.FormatBool(draft), "Manual spec", "Opening manual text.", draft)
	if draft {
		form.Set("case_slug", "named-work")
	}

	for range 2 {
		response := webDumpEditorPost(handler, form)
		if response.Code != http.StatusOK || response.Header().Get("HX-Retarget") != "#modal-root" {
			t.Fatal(response.Code, response.Body.String())
		}
	}
}

func checkWebSpecCreatedStates(t *testing.T, documents []core.SpecDocument) {
	t.Helper()

	for _, document := range documents {
		want := "approved"
		if document.CaseSlug != "" {
			want = "draft"
		}

		if document.Status != want {
			t.Fatal(document)
		}
	}
}

func checkWebSpecReadMarkup(t *testing.T, path, body string) {
	t.Helper()

	if path == "/specs/detail/fragment" {
		for _, want := range []string{"Project-wide", "Named work", `name="draft"`, `name="content"`, `role="dialog"`} {
			if !strings.Contains(body, want) {
				t.Fatal(want, body)
			}
		}

		if strings.Count(body, "<textarea") != 1 {
			t.Fatal("duplicate body editor")
		}
	}
}
