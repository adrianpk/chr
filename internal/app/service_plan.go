package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"forge.adrianpk.com/chr/internal/core"
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

//nolint:wsl_v5 // Guard clauses read clearer in this service flow.
func (s service) GeneratePlan(ctx context.Context, request PlanGenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	operation := generationcore.PlanPublicationOperation{ID: request.OperationID, Scope: scope, ExpectedRevision: request.ExpectedRevision, TrackerReplaceHash: request.TrackerReplaceHash, TicketDelivery: request.TicketDelivery}
	if result, found, err := s.replayPlanPublication(ctx, operation); found || err != nil {
		return result, err
	}
	existing, exists, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanOperationResult{}, err
	}

	if exists {
		return PlanOperationResult{
			Existing: true,
			Message:  "Plan already exists. Use regenerate plan to replace it.",
			Plan:     existing,
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

func (s service) existingGeneratedPlan(scope quizcore.Scope) (PlanView, bool, error) {
	paths, err := core.GeneratedPlanLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return PlanView{}, false, err
	}

	_, err = os.Stat(paths.Path)
	if errors.Is(err, os.ErrNotExist) {
		return PlanView{}, false, nil
	}

	if err != nil {
		return PlanView{}, false, fmt.Errorf("stat generated plan: %w", err)
	}

	return s.planViewFromPath(scope, paths.RelativePath, paths.Path), true, nil
}

//nolint:wsl_v5 // Keeps result capture adjacent to the guarded runtime call.
func (s service) RegeneratePlan(ctx context.Context, request PlanRegenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	operation := generationcore.PlanPublicationOperation{ID: request.OperationID, Scope: scope, ExpectedRevision: request.ExpectedRevision, Replace: true, TrackerReplaceHash: request.TrackerReplaceHash}
	if result, found, err := s.replayPlanPublication(ctx, operation); found || err != nil {
		return result, err
	}
	var result PlanOperationResult

	err := s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var err error

		result, err = s.prepareAndPublishPlan(ctx, operation)

		return err
	})

	return result, err
}
