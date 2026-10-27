package app

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramAuthoringService interface {
	ObserveTelegramAuthoring(context.Context, telegramcore.OperationTarget, string) (authoringcore.Snapshot, error)
	ContinueTelegramAuthoring(context.Context, telegramcore.OperationTarget, authoringcore.ContinuationRequest, authoringcore.CommitGuard) (authoringcore.ContinuationResult, error)
}

// ObserveTelegramAuthoring resolves a frozen target before the shared operation.
func (s service) ObserveTelegramAuthoring(ctx context.Context, target telegramcore.OperationTarget, id string) (authoringcore.Snapshot, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return authoringcore.Snapshot{}, err
	}

	return scoped.ObserveAuthoring(ctx, id)
}

// ContinueTelegramAuthoring adapts registry routing to the neutral guarded owner.
func (s service) ContinueTelegramAuthoring(ctx context.Context, target telegramcore.OperationTarget, request authoringcore.ContinuationRequest, guard authoringcore.CommitGuard) (authoringcore.ContinuationResult, error) {
	if request.ProjectID != target.ProjectID || request.Namespace != target.Namespace {
		return authoringcore.ContinuationResult{}, workflowstate.ErrConflict
	}

	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return authoringcore.ContinuationResult{}, err
	}

	return scoped.ContinueAuthoringGuarded(ctx, request, guard)
}
