package app

import (
	"context"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) replayPlanPublication(ctx context.Context, operation core.PlanPublicationOperation) (PlanOperationResult, bool, error) {
	if operation.ID == "" {
		return PlanOperationResult{}, false, nil
	}

	committed, found, err := core.LookupPlanPublication(ctx, s.projectContext.ProjectDir, operation)
	if err != nil || !found {
		return PlanOperationResult{}, found, err
	}

	reference := committed.Reference
	if committed.Binding.BodyRef != "" {
		layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
		if err != nil {
			return PlanOperationResult{}, false, err
		}

		reference = filepath.ToSlash(filepath.Join(layout.RelativeRoot(), committed.Binding.BodyRef))
	}

	return PlanOperationResult{Message: "Plan generation already committed.", Plan: s.planViewFromPath(operation.Scope, reference, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(reference)))}, true, nil
}

func (s service) retainedPlanTracking(plan PlanView, binding core.PlanRevisionBinding) (PlanTrackingResult, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return PlanTrackingResult{}, err
	}

	reference := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), binding.CurrentTrackerRef))

	return PlanTrackingResult{Message: "Plan order is already retained. Approve native task materialization to allocate tickets.", Plan: plan, TrackingPath: reference}, nil
}
