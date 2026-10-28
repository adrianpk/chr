package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/webui"
)

// Real application projections offer role-specific choices without launching a
// driver or serializing private resource references into the project control.
func TestWebHarnessSelectionRead(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	marker := prepareWebHarnessProfiles(t, operations)
	provider := webHarnessProvider(operations)
	handler := webHarnessHandler(t, provider, nil)

	response := webWorkflowHTTP(t, handler, http.MethodGet, "/project/harnesses", nil)
	if response.Code != http.StatusOK {
		t.Fatal("chooser did not render", response.Code)
	}

	for _, text := range []string{"Worker default", "Generation default", "alpha · Codex · model-a", "example/model-pi · ready", "Restore legacy dispatch", "Changes apply to new operations", "Remote authentication: not tested"} {
		if !strings.Contains(response.Body.String(), text) {
			t.Fatalf("chooser omitted %q", text)
		}
	}

	for _, private := range []string{operations.env.ConfigDir(), "fixture-provider-secret", "credential_ref", "model_config_ref"} {
		if strings.Contains(response.Body.String(), private) {
			t.Fatal("chooser exposed private configuration")
		}
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("inspection launched a driver")
	}
}

// A claimed invocation projects its original binding separately from defaults.
// Reading this projection cannot resume work or mutate the retained invocation.
func TestWebHarnessSelectionBinding(t *testing.T) {
	operations, _, request := serviceWorkflowTicketFixture(t)
	prepareWebHarnessProfiles(t, operations)

	before := claimWebHarnessInvocation(t, operations, request)

	view, err := webHarnessProvider(operations).HarnessSelectionView(t.Context())
	if err != nil || len(view.Retained) != 1 || view.Retained[0].ID != before.ID || view.Retained[0].Label != "alpha · Codex · model-a" {
		t.Fatal("retained binding was not projected", view, err)
	}

	after, err := operations.ReadWorkflow(t.Context(), before.ID)
	if err != nil || after.Revision != before.Revision || *after.HarnessBinding != *before.HarnessBinding {
		t.Fatal("inspection changed the retained invocation", err)
	}
}

func claimWebHarnessInvocation(t *testing.T, operations service, request workflowcore.StartRequest) workflowstate.Invocation {
	t.Helper()

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), started.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})

	before, err := operations.ReadWorkflow(t.Context(), started.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	return before
}

func prepareWebHarnessProfiles(t *testing.T, operations service) string {
	t.Helper()
	t.Setenv(workspacecore.ActiveOperatorIDEnv, "test-operator")

	directory := operations.env.ConfigDir()
	program, marker := filepath.Join(directory, "driver"), filepath.Join(directory, "spawned")
	harnessSelectionWrite(t, program, "#!/bin/sh\ntouch '"+marker+"'\n")

	if err := os.Chmod(program, 0o700); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(directory, "auth.json"), `{"example":{"type":"api_key","key":"fixture-provider-secret"}}`)

	profiles := fmt.Sprintf("profiles:\n  alpha:\n    harness: codex\n    command: %s\n    model: model-a\n  beta:\n    harness: codex\n    command: %s\n    model: model-b\n  pi-writing:\n    harness: pi\n    command: %s\n    model_provider: example\n    model: model-pi\n    credential_ref: auth.json\n  unavailable:\n    harness: codex\n    command: /unavailable-driver\n", program, program, program)
	harnessSelectionWrite(t, filepath.Join(directory, "harnesses.yml"), profiles)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nharnesses:\n  worker: alpha\n")

	return marker
}

func webHarnessProvider(operations service) webViewProvider {
	return webViewProvider{env: operations.env, assetsFS: operations.assetsFS, projectContext: operations.projectContext, serviceFactory: func(ctx context.Context) service {
		copy := operations
		copy.ctx = ctx

		return copy
	}}
}

func webHarnessHandler(t *testing.T, provider webViewProvider, login *webui.LoginCredentials) http.Handler {
	t.Helper()

	handler, err := webui.NewHandler(webui.HandlerOptions{Provider: provider, Login: login, Data: webui.ShellData{Title: "chr", ProjectID: provider.projectContext.ProjectID, ProjectName: "Fixture project", RefreshSeconds: 30}})
	if err != nil {
		t.Fatal(err)
	}

	return handler
}
