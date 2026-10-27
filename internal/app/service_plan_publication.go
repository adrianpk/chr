package app

import (
	"context"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) prepareAndPublishPlan(ctx context.Context, operation generationcore.PlanPublicationOperation) (PlanOperationResult, error) {
	bound, ctx, err := s.bindGeneration(ctx, "plan-"+dumpRequestID(operation.ID), string(quizmodel.ActionGeneratePlan))
	if err != nil {
		return PlanOperationResult{}, err
	}

	s = bound
	scope, expected, replace := operation.Scope, operation.ExpectedRevision, operation.Replace
	s.appendPlanGenerationLog("generate_plan_started", map[string]string{"case_slug": scope.CaseSlug, "timeout": longGenerationTimeout.String()})

	sources, state, target, err := s.preparePlanGeneration(ctx, scope, expected)
	if err != nil {
		return PlanOperationResult{}, err
	}

	plan, requests, responses, err := s.generatePlanWithConfiguredBackend(ctx, scope, state, sources)
	if err != nil {
		return PlanOperationResult{}, err
	}

	plan.SourceRequestID, plan.SourceResponseID = strings.Join(requests, ","), strings.Join(responses, ",")
	plan.ExpectedRevision, plan.SourceObservationHash = target, sources.Hash

	plan.RequireOrderMap = true
	plan.Order = core.ScopePlanDeliveryOrder(plan.Order, operation)

	plan.TrackerReplaceHash = operation.TrackerReplaceHash
	if operation.ID != "" {
		plan.Operation = &operation
	}

	plan.CreateOnly = !replace
	for _, input := range sources.Inputs {
		plan.SourceInputs = append(plan.SourceInputs, generationcore.PlanSourceBinding{Kind: input.Kind, Reference: input.Reference, Revision: input.Revision})
	}

	if replace {
		plan.ArchiveAt = s.utcNow()
	}

	if err := generationcore.ValidateGeneratedPlan(plan); err != nil {
		return PlanOperationResult{}, err
	}

	paths, err := s.publishObservedPlan(ctx, plan, sources)
	if err != nil {
		return PlanOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_written", map[string]string{"case_slug": scope.CaseSlug, "source_observation_hash": sources.Hash, "relative_path": paths.RelativePath})

	message := "Plan generated."
	if replace {
		message = "Plan regenerated."
	}

	return PlanOperationResult{Message: message, Plan: s.planViewFromPath(scope, paths.RelativePath, paths.Path)}, nil
}

func (s service) preparePlanGeneration(ctx context.Context, scope quizcore.Scope, expected string) (core.PlanSourceObservation, quizcore.RuntimeState, string, error) {
	var (
		sources core.PlanSourceObservation
		state   quizcore.RuntimeState
	)

	target, err := core.PlanTargetRevision(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return sources, state, target, err
	}

	if expected != "" && expected != target {
		return sources, state, target, workflowstate.ErrConflict
	}

	sources, err = s.observePlanContext(ctx, scope)
	if err != nil {
		return sources, state, target, err
	}

	if _, err := s.assessPlanReadiness(ctx, sources); err != nil {
		return sources, state, target, err
	}

	state, err = s.planObservedRuntimeState(sources)

	return sources, state, target, err
}

func (s service) planObservedRuntimeState(sources core.PlanSourceObservation) (quizcore.RuntimeState, error) {
	if sources.Quiz.Exists {
		return sources.Quiz.State, nil
	}

	state, err := s.newQuizRuntimeState(sources.Scope, quizcore.ModeDeep)
	state.Status = quizcore.RuntimeStatusDone

	return state, err
}

func (s service) publishObservedPlan(ctx context.Context, plan generationcore.GeneratedPlan, sources core.PlanSourceObservation) (paths core.GeneratedPlanPaths, err error) {
	err = core.WithAuthoringMutation(ctx, s.projectContext.ProjectDir, func(ctx context.Context) error {
		current, readErr := s.observePlanContext(ctx, sources.Scope)
		if readErr != nil {
			return readErr
		}

		if current.Hash != sources.Hash {
			return fmt.Errorf("plan sources changed; reload context and generate again: %w", workflowstate.ErrConflict)
		}

		paths, readErr = core.WriteGeneratedPlan(ctx, s.projectContext.ProjectDir, plan)

		return readErr
	})

	return paths, err
}
