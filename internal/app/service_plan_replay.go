package app

import (
	"context"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

type planPublicationReplay struct {
	Result PlanOperationResult
	Found  bool
}

func (s service) replayPlanPublication(ctx context.Context, operation generationcore.PlanPublicationOperation) (planPublicationReplay, error) {
	if operation.ID == "" {
		return planPublicationReplay{}, nil
	}

	observed, err := core.LookupPlanPublication(ctx, s.projectContext.ProjectDir, operation)
	if err != nil || !observed.Found {
		return planPublicationReplay{Found: observed.Found}, err
	}

	committed := observed.Result

	reference := committed.Reference
	if committed.Binding.BodyRef != "" {
		layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: operation.Scope.Namespace})
		if err != nil {
			return planPublicationReplay{}, err
		}

		reference = filepath.ToSlash(filepath.Join(layout.RelativeRoot(), committed.Binding.BodyRef))
	}

	return planPublicationReplay{Result: PlanOperationResult{Message: "Plan generation already committed.", Plan: s.planViewFromPath(operation.Scope, reference, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(reference)))}, Found: true}, nil
}

func (s service) retainedPlanTracking(plan PlanView, binding generationcore.PlanRevisionBinding) (PlanTrackingResult, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: plan.Scope.Namespace})
	if err != nil {
		return PlanTrackingResult{}, err
	}

	reference := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), binding.CurrentTrackerRef))

	return PlanTrackingResult{Message: "Plan order is already retained. Approve native task materialization to allocate tickets.", Plan: plan, TrackingPath: reference}, nil
}
