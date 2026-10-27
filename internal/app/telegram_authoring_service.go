package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramAuthoringService interface {
	ObserveTelegramAuthoring(context.Context, telegramcore.OperationTarget, string) (core.AuthoringSnapshot, error)
	ContinueTelegramAuthoring(context.Context, telegramcore.OperationTarget, core.AuthoringContinuationRequest, core.AuthoringCommitGuard) (core.AuthoringContinuationResult, error)
}

// ObserveTelegramAuthoring resolves a frozen target before the shared operation.
func (s service) ObserveTelegramAuthoring(ctx context.Context, target telegramcore.OperationTarget, id string) (core.AuthoringSnapshot, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return core.AuthoringSnapshot{}, err
	}

	return scoped.ObserveAuthoring(ctx, id)
}

// ContinueTelegramAuthoring adapts registry routing to the neutral guarded owner.
func (s service) ContinueTelegramAuthoring(ctx context.Context, target telegramcore.OperationTarget, request core.AuthoringContinuationRequest, guard core.AuthoringCommitGuard) (core.AuthoringContinuationResult, error) {
	if request.ProjectID != target.ProjectID || request.Namespace != target.Namespace {
		return core.AuthoringContinuationResult{}, workflowstate.ErrConflict
	}

	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return core.AuthoringContinuationResult{}, err
	}

	return scoped.ContinueAuthoringGuarded(ctx, request, guard)
}
