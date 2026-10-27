package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// PlanGenerateRequest selects the quiz scope used to generate a plan.
type PlanGenerateRequest struct {
	TicketDelivery     bool
	OperationID        string
	TrackerReplaceHash string
	ExpectedRevision   string
	Scope              quizcore.Scope
}

// PlanRegenerateRequest selects the quiz scope used to replace a plan.
type PlanRegenerateRequest struct {
	OperationID        string
	TrackerReplaceHash string
	ExpectedRevision   string
	Scope              quizcore.Scope
}

// PlanOperationResult is the TUI-facing result for plan generation operations.
type PlanOperationResult struct {
	Existing bool
	Message  string
	Plan     PlanView
}

type existingPlanResult struct {
	Plan   PlanView
	Exists bool
}

//nolint:wsl_v5 // Guard clauses read clearer in this service flow.
func (s service) GeneratePlan(ctx context.Context, request PlanGenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	operation := generationcore.PlanPublicationOperation{ID: request.OperationID, Scope: scope, ExpectedRevision: request.ExpectedRevision, TrackerReplaceHash: request.TrackerReplaceHash, TicketDelivery: request.TicketDelivery}
	if replay, err := s.replayPlanPublication(ctx, operation); replay.Found || err != nil {
		return replay.Result, err
	}
	existing, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanOperationResult{}, err
	}

	if existing.Exists {
		return PlanOperationResult{
			Existing: true,
			Message:  "Plan already exists. Use regenerate plan to replace it.",
			Plan:     existing.Plan,
		}, nil
	}

	var result PlanOperationResult

	err = s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var generateErr error

		result, generateErr = s.prepareAndPublishPlan(ctx, operation)

		return generateErr
	})

	return result, err
}

func (s service) existingGeneratedPlan(scope quizcore.Scope) (existingPlanResult, error) {
	paths, err := authoringcore.GeneratedPlanLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return existingPlanResult{}, err
	}

	_, err = os.Stat(paths.Path)
	if errors.Is(err, os.ErrNotExist) {
		return existingPlanResult{}, nil
	}

	if err != nil {
		return existingPlanResult{}, fmt.Errorf("stat generated plan: %w", err)
	}

	return existingPlanResult{Plan: s.planViewFromPath(scope, paths.RelativePath, paths.Path), Exists: true}, nil
}

//nolint:wsl_v5 // Keeps result capture adjacent to the guarded runtime call.
func (s service) RegeneratePlan(ctx context.Context, request PlanRegenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	operation := generationcore.PlanPublicationOperation{ID: request.OperationID, Scope: scope, ExpectedRevision: request.ExpectedRevision, Replace: true, TrackerReplaceHash: request.TrackerReplaceHash}
	if replay, err := s.replayPlanPublication(ctx, operation); replay.Found || err != nil {
		return replay.Result, err
	}
	var result PlanOperationResult

	err := s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var err error

		result, err = s.prepareAndPublishPlan(ctx, operation)

		return err
	})

	return result, err
}
