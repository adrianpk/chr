package app

import (
	"path/filepath"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	"forge.adrianpk.com/chr/internal/webui"
)

// Every Web ticket operation uses the same canonical document and guarded service.
func TestWebTicketEditor(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "unadopted", true: "native"}[native], func(t *testing.T) {
			provider := webTicketEditorProvider(t, native)

			response, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{
				Action: webActionTicketCreate, Confirm: true,
				Fields: map[string]string{"title": "Created ticket", "description": "Full new description.\n\nRetained second paragraph.", "use_case": "named-work", "priority": "high", "type": "bug", "owner": "Maintainer", "assignee": "Developer", "status": ticketcore.TicketStatusReadyForDev},
			})
			if err != nil || response.Status != webui.WriteStatusOK {
				t.Fatal("create failed", response, err)
			}

			view, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: response.Operation.ID})
			if err != nil || view.Description != "Full new description.\n\nRetained second paragraph." || view.Fields["owner"] != "Maintainer" || view.Fields["assignee"] != "Developer" || view.NativeWorkflow != native {
				t.Fatal("reload lost created fields", view.Fields, err)
			}

			checkWebTicketMoves(t, provider, view)
		})
	}
}

func webTicketEditorProvider(t *testing.T, native bool) webViewProvider {
	t.Helper()

	project := newTestWorkspace(t)
	if native {
		serviceWorkflowDeliveryGit(t, project, "init", "-q")
		writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: CHR\nwork:\n  state_authority: tickets\n")
	}

	if _, err := authoringcore.NewCaseMetadataService(project).Save(authoringcore.CaseMetadata{Name: "Named work", Slug: "named-work", Description: "Selected work."}); err != nil {
		t.Fatal(err)
	}

	return newWebServeBoardTestProvider(project, newWebServeBoardTestEnv(t, t.TempDir()), newWebServeBoardProjectContext(t, project))
}

func checkWebTicketMoves(t *testing.T, provider webViewProvider, original webui.TicketDetailView) {
	t.Helper()

	response, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{
		Action: webActionTicketMove, ID: original.Key, Revision: original.Revision, Confirm: true,
		Fields: map[string]string{"status": ticketcore.TicketStatusBacklog},
	})
	if err != nil || response.Status != webui.WriteStatusOK {
		t.Fatal("allowed move rejected", response, err)
	}

	current, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: original.Key})
	if err != nil || current.Status != ticketcore.TicketStatusBacklog || current.Description != original.Description || current.UseCase != "named-work" {
		t.Fatal("move changed fields or case membership", current.Fields, err)
	}

	checkWebTicketRejections(t, provider, original, current)

	if filepath.IsAbs(current.RelativePath) {
		t.Fatal("response exposed an absolute storage path")
	}
}

func checkWebTicketRejections(t *testing.T, provider webViewProvider, original, current webui.TicketDetailView) {
	t.Helper()

	for _, test := range []struct {
		action, revision, status string
		fields                   map[string]string
	}{
		{webActionTicketMove, current.Revision, webui.WriteStatusValidation, map[string]string{"status": ticketcore.TicketStatusDone}},
		{webActionTicketUpdate, original.Revision, webui.WriteStatusStale, map[string]string{"title": "Stale title"}},
	} {
		response, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{Action: test.action, ID: original.Key, Revision: test.revision, Confirm: true, Fields: test.fields})
		if err != nil || response.Status != test.status {
			t.Fatal("unsafe mutation was accepted", response, err)
		}
	}
}
