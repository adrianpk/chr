package app

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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

	view.FormRevision = view.Revision
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

func webHarnessRole(selection harnesscore.Selection, profile string, generation bool) webui.HarnessRoleView {
	view := webui.HarnessRoleView{Current: webHarnessLabel(profile, selection.Harness, selection.ModelProvider, selection.Model), Problem: selection.Readiness.Problem}

	view.Selected = webui.HarnessKeepProfile
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

func (p webViewProvider) writeHarnessSelection(ctx context.Context, operations service, request webui.WriteActionRequest) webui.WriteActionResponse {
	if !webHarnessEnvelopeValid(request) || !webHarnessFieldsValid(request.Fields) {
		return webValidationWrite(errors.New("use the exact project harness form and named profile choices"))
	}

	if request.Fields["project_ref"] != p.projectContext.ProjectID {
		return webBlockedWrite("Selection changes belong to the current project.")
	}

	worker, generation := webHarnessChoice(request.Fields["worker"]), webHarnessChoice(request.Fields["generation"])
	if worker == nil && generation == nil {
		return webValidationWrite(errors.New("choose a change for at least one role"))
	}

	actor, err := workspacecore.ResolveUserspace(workspacecore.ResolveUserspaceOptions{ProjectDir: operations.projectContext.ProjectDir})
	if err != nil {
		return webBlockedWrite("The active local operator is unavailable.")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = operations.ChangeHarnessSelections(ctx, core.HarnessSelectionChangeRequest{
		Binding:    workflowstate.MutationBinding{RequestID: request.Fields["request_id"], Operator: actor.OperatorID},
		ProjectRef: request.Fields["project_ref"], ExpectedRevision: request.Revision, Worker: worker, Generation: generation,
	}, localHarnessSelectionAuthority(operations.projectContext.ProjectDir, actor.OperatorID))
	if errors.Is(err, workflowstate.ErrConflict) {
		return webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "The defaults or request identity changed. Your choices are retained; reload current defaults before saving."}
	}

	if err != nil {
		return webValidationWrite(errors.New("the selected profiles are unavailable or not locally ready; check private profile setup and choose a prepared profile"))
	}

	return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: "Harness defaults saved. Changes apply to new operations."}
}

func webHarnessEnvelopeValid(request webui.WriteActionRequest) bool {
	return request.Action == webui.HarnessSelectionAction && request.Surface == "harnesses" && request.ID == "" && request.Family == "" && len(request.Revision) == 64 && len(request.Fields) == 4
}

func webHarnessFieldsValid(fields map[string]string) bool {
	limits := map[string]int{"project_ref": 256, "request_id": 256, "worker": 128, "generation": 128}
	for name, value := range fields {
		limit, found := limits[name]
		if !found || len(value) > limit {
			return false
		}
	}

	return true
}

func webHarnessChoice(value string) *string {
	if value == webui.HarnessKeepProfile {
		return nil
	}

	return &value
}

func webHarnessChoiceSetup(readiness harnesscore.Readiness) string {
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
