package app

import (
	"context"
	"crypto/rand"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

// HarnessSelectionView observes the shared selection service and bounded
// retained workflow status. It never exposes private profile references.
func (p webViewProvider) HarnessSelectionView(ctx context.Context) (webui.HarnessSelectionView, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	operations := p.service(ctx)
	status, err := operations.InspectHarnessSelections(ctx, "spec_generation")

	view := webui.HarnessSelectionView{ProjectID: p.projectContext.ProjectID, Revision: status.Snapshot.Revision, RequestID: rand.Text()}
	if err != nil {
		return view, err
	}

	view.Worker = webHarnessRole(status.Worker, status.Snapshot.Worker, false)

	view.Generation = webHarnessRole(status.Generation, status.Snapshot.Generation, true)
	for _, profile := range status.Profiles {
		label := webHarnessLabel(profile.Name, profile.Harness, profile.ModelProvider, profile.Model)
		view.Worker.Choices = append(view.Worker.Choices, webui.TicketOptionView{Value: profile.Name, Label: label + webHarnessChoiceSetup(profile.Worker)})
		view.Generation.Choices = append(view.Generation.Choices, webui.TicketOptionView{Value: profile.Name, Label: label + webHarnessChoiceSetup(profile.Generation)})
	}

	page, err := operations.WorkflowStatusPage(ctx, 0, 8)
	if err != nil {
		view.RetainedProblem = "Retained work could not be loaded. Inspect Workflows for its current state."

		return view, nil
	}

	view.More = page.NextOffset >= 0
	for _, run := range page.Items {
		label := "Legacy worker (original invocation settings)"
		if binding := run.HarnessBinding; binding != nil {
			label = webHarnessLabel(binding.Profile, binding.Harness, binding.ModelProvider, binding.Model)
		}

		view.Retained = append(view.Retained, webui.HarnessBindingView{ID: run.ID, Status: run.Status, Label: label})
	}

	return view, nil
}

func webHarnessRole(selection core.HarnessSelection, profile string, generation bool) webui.HarnessRoleView {
	view := webui.HarnessRoleView{Current: webHarnessLabel(profile, selection.Harness, selection.ModelProvider, selection.Model), Problem: selection.Readiness.Problem}
	if profile == "" {
		view.Current = "Legacy dispatch · Codex CLI default"
		if generation {
			view.Current = "Legacy dispatch · varies by operation"
		}
	}

	readiness := selection.Readiness
	view.Setup = "Local setup" + webHarnessChoiceSetup(readiness) + ". Credentials: " + strings.ReplaceAll(readiness.Credentials, "_", " ") + "; runtime: " + strings.ReplaceAll(readiness.Runtime, "_", " ") + ". Remote authentication: " + strings.ReplaceAll(readiness.RemoteAuthentication, "_", " ") + "."

	return view
}

func webHarnessChoiceSetup(readiness core.HarnessReadiness) string {
	if readiness.Ready() {
		return " · ready"
	}

	return " · not ready"
}

func webHarnessLabel(profile, identity, provider, model string) string {
	parts := []string{}
	if profile != "" {
		parts = append(parts, profile)
	}

	if identity != "" {
		parts = append(parts, map[string]string{"codex": "Codex", "pi": "Pi"}[identity])
	}

	if model == "cli_default" {
		model = "CLI default"
	}

	if provider != "" && provider != "cli_default" {
		model = provider + "/" + model
	}

	if model != "" {
		parts = append(parts, model)
	}

	return strings.Join(parts, " · ")
}
