package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/webui"
	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// The real HTTP Save replaces only the observed body's canonical identity.
// Frontmatter stays outside the editor, and a second editor cannot overwrite
// that saved body with its older revision.
func TestWebDumpSaveDocument(t *testing.T) {
	provider := newWebDumpCaptureProvider(t)
	root := provider.projectContext.ProjectDir
	reference := serviceDumpArtifact(t, root, "team", "Existing", "---\ntitle: Existing\ncustom: retained\n---\nOriginal material.\n")

	if _, err := provider.operations.CreateDump(DumpCreateRequest{OperationID: "capture", Title: "Existing", Content: "Captured material."}); err != nil {
		t.Fatal(err)
	}

	detail := webDumpEditorDetail(t, provider, reference)
	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{"action": {webActionDumpSave}, "family": {"dumps"}, "surface": {"modal"}, "confirm": {"true"}, "id": {detail.ID}, "revision": {detail.Revision}, "request_id": {detail.SaveRequestID}, "content": {"Corrected material.\n\nAdded in the same editor.\n"}}

	recorder := webDumpEditorPost(handler, form)
	if recorder.Code != http.StatusOK || recorder.Header().Get(htmx.HeaderRetarget) != "#modal-root" {
		t.Fatal("Save did not commit through HTTP", recorder.Code, recorder.Body.String())
	}

	checkWebDumpSavedDocument(t, root, reference)
	form.Set("request_id", "stale-editor")
	form.Set("content", "Stale replacement.")

	recorder = webDumpEditorPost(handler, form)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `data-write-status="stale"`) {
		t.Fatal("HTTP accepted stale replacement", recorder.Code, recorder.Body.String())
	}

	checkWebDumpSavedDocument(t, root, reference)
}

func webDumpEditorDetail(t *testing.T, provider webDumpCaptureProvider, reference string) webui.ArtifactDetailView {
	t.Helper()

	detail, err := provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "dumps", ID: reference})
	if err != nil || detail.ID != reference || strings.Contains(detail.Content, "custom: retained") || !strings.Contains(detail.Content, "Original material.") {
		t.Fatal("editor did not isolate the observed body", detail, err)
	}

	return detail
}

func webDumpEditorPost(handler http.Handler, form url.Values) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set(htmx.HeaderRequest, "true")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return recorder
}

func checkWebDumpSavedDocument(t *testing.T, root, reference string) {
	t.Helper()

	document, err := authoringcore.ReadDumpDocument(t.Context(), root, reference)
	if err != nil || !strings.Contains(document.Content, "custom: retained") || !strings.HasSuffix(document.Content, "Corrected material.\n\nAdded in the same editor.\n") || strings.Contains(document.Content, "Original material.") {
		t.Fatal("Save lost frontmatter or failed to replace body", document, err)
	}

	observed, err := authoringcore.ObserveProjectDump(t.Context(), root)
	if err != nil || observed.Destination != reference || len(observed.Candidates) != 1 || len(observed.Components) != 1 {
		t.Fatal("Save changed identity or raw provenance", observed, err)
	}
}
