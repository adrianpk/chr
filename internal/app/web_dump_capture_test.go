package app

import (
	"context"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/webui"
)

// New capture lists only existing case names, reuses the project destination and
// keeps multiple named captures independent. Missing identity and unknown scope
// fail without manufacturing a case or another capture.
func TestWebDumpCapture(t *testing.T) {
	provider := newWebDumpCaptureProvider(t)
	checkWebDumpCaptureChoices(t, provider.webViewProvider)

	first := webDumpCaptureRequest("general-first", "", "First material.")
	for _, request := range []webui.WriteActionRequest{first, first, webDumpCaptureRequest("general-second", "", "Second material."), webDumpCaptureRequest("case-first", "named-work", "Named material."), webDumpCaptureRequest("case-second", "named-work", "More named material.")} {
		result, err := provider.WriteAction(t.Context(), request)
		if err != nil || result.Status != webui.WriteStatusOK || result.Operation.Status == "" {
			t.Fatal("capture failed", result, err)
		}
	}

	for _, request := range []webui.WriteActionRequest{webDumpCaptureRequest("", "", "Unbound material."), webDumpCaptureRequest("unknown", "unknown", "Unknown case.")} {
		result, err := provider.WriteAction(t.Context(), request)
		if err != nil || result.Status != webui.WriteStatusValidation {
			t.Fatal("invalid capture accepted", result, err)
		}
	}

	checkWebDumpCaptureInventory(t, provider.webViewProvider)
}

func webDumpCaptureRequest(id, slug, content string) webui.WriteActionRequest {
	return webui.WriteActionRequest{Action: webActionDumpCreate, Family: "dumps", Confirm: true, Fields: map[string]string{"request_id": id, "case_slug": slug, "title": "Captured context", "content": content}}
}

func checkWebDumpCaptureInventory(t *testing.T, provider webViewProvider) {
	t.Helper()

	root := provider.projectContext.ProjectDir

	observed, err := authoringcore.ObserveProjectDump(t.Context(), root)
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 2 || len(observed.Scoped) != 2 {
		t.Fatal("capture scope fragmented or collapsed", observed, err)
	}

	document, err := authoringcore.ReadDumpDocument(t.Context(), root, observed.Destination)
	if err != nil || strings.Count(document.Content, "First material.") != 1 || strings.Count(document.Content, "Second material.") != 1 || strings.Contains(document.Content, "Named material.") {
		t.Fatal("capture duplicated or mixed scope", document, err)
	}
}

func checkWebDumpCaptureChoices(t *testing.T, provider webViewProvider) {
	t.Helper()

	catalog, err := provider.CatalogView(t.Context(), "dumps")
	if err != nil || len(catalog.Cases) != 1 || catalog.Cases[0].Name != "Named work" || catalog.Cases[0].Slug != "named-work" {
		t.Fatal("capture choices lost canonical case identity", catalog, err)
	}
}

// Capture fixtures substitute only the provider boundary; the real application
// operation, journal, catalog and HTTP forms remain exercised.
type webDumpCaptureProvider struct {
	webViewProvider
	operations service
}

func newWebDumpCaptureProvider(t *testing.T) webDumpCaptureProvider {
	t.Helper()

	provider := webTicketEditorProvider(t, false)
	operations := provider.service(t.Context())
	operations.dumpCurator = &countingDumpCurator{}

	return webDumpCaptureProvider{webViewProvider: provider, operations: operations}
}

func (p webDumpCaptureProvider) WriteAction(ctx context.Context, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if request.Action != webActionDumpCreate || !request.Confirm {
		return p.webViewProvider.WriteAction(ctx, request)
	}

	operations := p.operations
	operations.ctx = ctx

	return p.writeDumpCreate(operations, request)
}
