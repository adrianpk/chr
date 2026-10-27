package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/tui"
	"forge.adrianpk.com/chr/internal/webui"
)

func (f *harnessJourney) change(t *testing.T, choice harnessJourneyChoice) {
	t.Helper()

	switch choice.transport {
	case "web":
		handler, cookies := webHarnessAuthenticated(t, f.operations)
		form := webHarnessForm(t, handler, cookies)
		form.Set("worker", choice.worker)
		form.Set("generation", choice.generation)
		webHarnessAssertWrite(t, handler, cookies, form, webui.WriteStatusOK)
	case "tui":
		view, err := newTUIHarnessSelectionExecutor(f.operations).InspectHarnessSelections()
		if err != nil {
			t.Fatal(err)
		}

		_, err = newTUIHarnessSelectionExecutor(f.operations).SaveHarnessSelections(tui.HarnessSelectionRequest{ProjectRef: view.Snapshot.ProjectRef, Revision: view.Snapshot.Revision, RequestID: view.RequestID, Worker: &choice.worker, Generation: &choice.generation})
		if err != nil {
			t.Fatal(err)
		}
	case "telegram":
		_, action, _ := telegramHarnessOffer(t, f.telegram, "/harness worker="+choice.worker+" generation="+choice.generation)
		if notice := f.telegram.send(t, "/answer "+action.Token+" confirm", 0); !strings.Contains(notice.Text, "Harness defaults saved") {
			t.Fatal("Telegram confirmation failed", notice)
		}
	case "cli":
		snapshot := webHarnessSnapshot(t, f.operations)

		args := []string{"harness", "select", "--dev-root", f.devRoot(), "--config-revision", snapshot.Revision, "--request-id", "journey-cli", "--worker", choice.worker, "--generation", choice.generation}
		if err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("unknown journey transport")
	}
}

func (f *harnessJourney) devRoot() string {
	return filepath.Dir(filepath.Dir(f.operations.env.ConfigDir()))
}

func (f *harnessJourney) checkViews(t *testing.T, choice harnessJourneyChoice) {
	t.Helper()

	snapshot := webHarnessSnapshot(t, f.operations)
	if snapshot.Worker != choice.worker || snapshot.Generation != choice.generation {
		t.Fatal("adapter did not publish both role choices", snapshot)
	}

	f.checkSharedViews(t, snapshot)
	f.checkTelegramView(t, choice)
}

func (f *harnessJourney) checkSharedViews(t *testing.T, snapshot core.HarnessSelectionSnapshot) {
	t.Helper()

	web, webErr := webHarnessProvider(f.operations).HarnessSelectionView(t.Context())
	terminal, tuiErr := newTUIHarnessSelectionExecutor(f.operations).InspectHarnessSelections()

	cli := f.cliStatus(t)
	if webErr != nil || tuiErr != nil || web.Revision != snapshot.Revision || terminal.Snapshot != snapshot || cli.Snapshot != snapshot {
		t.Fatal("interface readers did not observe one complete configuration", webErr, tuiErr)
	}
}

func (f *harnessJourney) checkTelegramView(t *testing.T, choice harnessJourneyChoice) {
	t.Helper()

	if notice := f.telegram.send(t, "/harness", 0); !strings.Contains(notice.Text, "Worker: "+choice.worker) || !strings.Contains(notice.Text, "Generation: "+choice.generation) {
		t.Fatal("Telegram read did not observe the published defaults", notice)
	}

	data, err := os.ReadFile(workspacecore.UserPath(f.operations.projectContext.ProjectDir, "config.yml"))
	if err != nil || !strings.Contains(string(data), "unknown_extension:\n  retained: yes\n") {
		t.Fatal("selection replaced unrelated configuration", err)
	}
}

func (f *harnessJourney) resumeGeneration(t *testing.T) {
	t.Helper()

	generation, _ := harnesshost.GenerationSelection(f.active)
	resumed := telegramHarnessGenerationBinding(t.Context(), t, f.operations, "original-generation")

	nested := telegramHarnessGenerationBinding(f.active, t, f.operations, "nested-generation")
	if generation.Binding.BindingHash != resumed.BindingHash || nested.BindingHash != generation.Binding.BindingHash || generation.Binding.Selection.Model != "fixture-model" {
		t.Fatal("nested or resumed generation replaced the active binding")
	}

	f.generate(f.active, t, "nested-original", "code-gen")
	f.generate(t.Context(), t, "original-generation", "code-gen")
	checkGenerationFixtureModels(t, f.operations.projectContext.ProjectDir, map[string]string{"original-generation": "fixture-model", "fresh-web": "replacement-model", "fresh-tui": "replacement-model", "fresh-telegram": "fixture-model", "fresh-cli": "fixture-model"})
}

func (f *harnessJourney) checkLegacyAPI(t *testing.T) {
	t.Helper()
	t.Setenv("CHR_SPEC_GENERATION_BACKEND", "openai")

	snapshot := webHarnessSnapshot(t, f.operations)

	args := []string{"harness", "clear", "--dev-root", f.devRoot(), "--config-revision", snapshot.Revision, "--request-id", "journey-api-legacy", "--generation"}
	if err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	f.generate(t.Context(), t, "legacy-api", "")

	if f.apiCalls.Load() != 1 {
		t.Fatal("legacy direct API selection changed its transport", f.apiCalls.Load())
	}

	status := f.cliStatus(t)
	if status.Generation.Backend != "openai" || status.Generation.Source != "legacy" || status.Snapshot.Worker != "code-worker" {
		t.Fatal("clearing generation changed worker or direct API compatibility", status.Generation)
	}
}

func (f *harnessJourney) cliStatus(t *testing.T) HarnessSelectionStatus {
	t.Helper()

	var output bytes.Buffer
	if err := run(t.Context(), []string{"harness", "status", "--dev-root", f.devRoot(), "--json"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var status HarnessSelectionStatus
	if err := json.Unmarshal(output.Bytes(), &status); err != nil {
		t.Fatal(err)
	}

	return status
}
