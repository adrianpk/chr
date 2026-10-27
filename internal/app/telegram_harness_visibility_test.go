package app

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramHarnessWorkerFixture(t *testing.T) (*telegramConversationAppFixture, service, core.WorkflowRecord, core.WorkflowStartRequest, string) {
	t.Helper()

	operations, record, request := serviceWorkflowTicketFixture(t)
	f := newTelegramConversationAppFixture(t)

	registry, err := projectcore.NewRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	updateErr := registry.AddProject(t.Context(), projectcore.Project{ID: operations.projectContext.ProjectID, Name: "Worker fixture", Path: operations.projectContext.ProjectDir, Enabled: true, RegisteredAt: time.Now()})

	closeErr := registry.Close()
	if updateErr != nil || closeErr != nil {
		t.Fatal(updateErr, closeErr)
	}

	operations.env = f.receiver.env
	marker := prepareWebHarnessProfiles(t, operations)

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = append(config.AllowedUsers[0].ProjectIDs, operations.projectContext.ProjectID)
	if err := telegramcore.SaveReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	f.send(t, "/project "+operations.projectContext.ProjectID, 0)

	return f, operations, record, request, marker
}

// Telegram publication reaches existing Web/TUI readers and a fresh process.
// Claimed workers and resumed/nested generation retain their original selection;
// independently admitted work observes the complete new defaults.
func TestTelegramHarnessSelectionVisibility(t *testing.T) {
	f, operations, record, start, marker := telegramHarnessWorkerFixture(t)
	switchGenerationFixture(t, operations, "alpha")

	worker := claimWebHarnessInvocation(t, operations, start)

	bound, err := operations.bindHarnessGeneration(t.Context(), "active-generation", "spec_generation")
	if err != nil {
		t.Fatal(err)
	}

	web := webHarnessProvider(operations)
	tui := newTUIHarnessSelectionExecutor(operations)

	request, action, _ := telegramHarnessOffer(t, f, "/harness worker=beta generation=pi-writing")
	if notice := f.send(t, "/answer "+action.Token+" confirm", 0); !strings.Contains(notice.Text, "Harness defaults saved") {
		t.Fatal(notice)
	}

	saved := telegramHarnessCheckSaved(t, operations, request.ID)
	webView, webErr := web.HarnessSelectionView(t.Context())

	tuiView, tuiErr := tui.InspectHarnessSelections()
	if webErr != nil || tuiErr != nil || webView.Revision != saved.Revision || tuiView.Snapshot != saved {
		t.Fatal("existing interface readers did not see Telegram publication", webErr, tuiErr)
	}

	webHarnessCheckProcess(t, operations, saved)
	webHarnessCheckRetained(t, operations, worker)
	webHarnessCheckNewInvocation(t, operations, record, start)
	telegramHarnessCheckGeneration(t.Context(), t, operations, bound)

	if notice := f.send(t, "/harness", 0); !strings.Contains(notice.Text, "Worker: beta") || !strings.Contains(notice.Text, "alpha · Codex · model-a") {
		t.Fatal("Telegram confused defaults with retained worker binding", notice)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("selection invoked a driver")
	}
}

func telegramHarnessCheckGeneration(ctx context.Context, t *testing.T, operations service, active context.Context) {
	t.Helper()

	id, original, bound := core.HarnessGenerationSelection(active)
	if !bound || original.Selection.Profile != "alpha" {
		t.Fatal("original generation fixture was not bound")
	}

	nested := telegramHarnessGenerationBinding(active, t, operations, "nested-call")

	resumed := telegramHarnessGenerationBinding(ctx, t, operations, id)
	if nested.BindingHash != original.BindingHash || resumed.BindingHash != original.BindingHash {
		t.Fatal("existing generation rebound after Telegram change")
	}

	selected := telegramHarnessGenerationBinding(ctx, t, operations, "new-generation")
	if selected.Selection.Profile != "pi-writing" || selected.Selection.Backend != "pi" {
		t.Fatal("independent generation missed the changed default")
	}
}

func telegramHarnessGenerationBinding(ctx context.Context, t *testing.T, operations service, id string) core.HarnessGenerationBinding {
	t.Helper()

	ctx, err := operations.bindHarnessGeneration(ctx, id, "spec_generation")
	if err != nil {
		t.Fatal(err)
	}

	_, binding, bound := core.HarnessGenerationSelection(ctx)
	if !bound {
		t.Fatal("generation selection was not bound")
	}

	return binding
}
