package app

import (
	"context"
	"crypto/rand"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
)

type tuiHarnessSelectionExecutor struct{ operations service }

func newTUIHarnessSelectionExecutor(operations service) tuiHarnessSelectionExecutor {
	return tuiHarnessSelectionExecutor{operations: operations}
}

func (e tuiHarnessSelectionExecutor) InspectHarnessSelections() (tui.HarnessSelectionView, error) {
	ctx, cancel := context.WithTimeout(e.operations.ctx, 10*time.Second)
	defer cancel()

	return e.inspectHarnessSelections(ctx)
}

func (e tuiHarnessSelectionExecutor) inspectHarnessSelections(ctx context.Context) (tui.HarnessSelectionView, error) {
	status, err := e.operations.InspectHarnessSelections(ctx, "spec_generation")

	view := tui.HarnessSelectionView{RequestID: rand.Text(), Snapshot: status.Snapshot, Worker: status.Worker, Generation: status.Generation}
	if err != nil {
		return view, err
	}

	for _, p := range status.Profiles {
		view.Profiles = append(view.Profiles, tui.HarnessProfileView{Name: p.Name, Harness: p.Harness, Provider: p.ModelProvider, Model: p.Model, Worker: p.Worker, Generation: p.Generation})
	}

	view.Retained, view.More, view.RetainedProblem = e.retainedHarnessBindings(ctx)

	return view, nil
}

func (e tuiHarnessSelectionExecutor) retainedHarnessBindings(ctx context.Context) ([]tui.HarnessRetainedView, bool, string) {
	page, err := e.operations.WorkflowStatusPage(ctx, 0, 4)
	if err != nil {
		return nil, false, "Retained work could not be loaded; use workflow status in the command palette."
	}

	var bindings []tui.HarnessRetainedView
	for _, run := range page.Items {
		bindings = append(bindings, tui.HarnessRetainedView{ID: run.ID, Status: run.Status, Binding: run.HarnessBinding})
	}

	return bindings, page.NextOffset >= 0, ""
}

func (e tuiHarnessSelectionExecutor) SaveHarnessSelections(request tui.HarnessSelectionRequest) (tui.HarnessSelectionView, error) {
	actor, err := core.ResolveProjectUserspace(core.ResolveProjectUserspaceOptions{ProjectDir: e.operations.projectContext.ProjectDir})
	if err != nil {
		return tui.HarnessSelectionView{}, err
	}

	ctx, cancel := context.WithTimeout(e.operations.ctx, 10*time.Second)
	defer cancel()

	_, err = e.operations.ChangeHarnessSelections(ctx, core.HarnessSelectionChangeRequest{
		Binding:    core.WorkflowMutationBinding{RequestID: request.RequestID, Operator: actor.OperatorID},
		ProjectRef: request.ProjectRef, ExpectedRevision: request.Revision, Worker: request.Worker, Generation: request.Generation,
	}, localHarnessSelectionAuthority(e.operations.projectContext.ProjectDir, actor.OperatorID))
	if err != nil {
		return tui.HarnessSelectionView{}, err
	}

	return e.inspectHarnessSelections(ctx)
}
