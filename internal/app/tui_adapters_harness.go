package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/tui"
)

type tuiHarnessSelectionExecutor struct{ operations service }

func newTUIHarnessSelectionExecutor(operations service) tuiHarnessSelectionExecutor {
	return tuiHarnessSelectionExecutor{operations: operations}
}

func (e tuiHarnessSelectionExecutor) InspectHarnessSelections() (tui.HarnessSelectionView, error) {
	ctx, cancel := context.WithTimeout(e.operations.ctx, 10*time.Second)
	defer cancel()

	status, err := e.operations.InspectHarnessSelections(ctx, "spec_generation")

	view := tui.HarnessSelectionView{Snapshot: status.Snapshot, Worker: status.Worker, Generation: status.Generation}
	if err != nil {
		return view, err
	}

	for _, p := range status.Profiles {
		view.Profiles = append(view.Profiles, tui.HarnessProfileView{Name: p.Name, Harness: p.Harness, Provider: p.ModelProvider, Model: p.Model, Worker: p.Worker, Generation: p.Generation})
	}

	page, err := e.operations.WorkflowStatusPage(ctx, 0, 4)
	if err != nil {
		view.RetainedProblem = "Retained work could not be loaded. Inspect Workflows."

		return view, nil
	}

	view.More = page.NextOffset >= 0
	for _, run := range page.Items {
		view.Retained = append(view.Retained, tui.HarnessRetainedView{ID: run.ID, Status: run.Status, Binding: run.HarnessBinding})
	}

	return view, nil
}
