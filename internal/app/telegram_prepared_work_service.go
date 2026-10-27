package app

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramPreparedWorkService interface {
	PrepareTelegramAuthoringWork(context.Context, telegramcore.OperationTarget, AuthoringWorkRequest) (PreparedAuthoringWork, error)
	ApplyTelegramAuthoringWork(context.Context, telegramcore.OperationTarget, PreparedAuthoringWork, authoringcore.CommitGuard) (AuthoringWorkResult, error)
}

// PrepareTelegramAuthoringWork resolves the current bound root before entering
// the shared authoring and native-record preparation owners.
func (s service) PrepareTelegramAuthoringWork(ctx context.Context, target telegramcore.OperationTarget, request AuthoringWorkRequest) (PreparedAuthoringWork, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return PreparedAuthoringWork{}, err
	}

	prepared, err := scoped.PrepareAuthoringWork(ctx, request)
	if err == nil && (prepared.Scope.ProjectID != target.ProjectID || prepared.Scope.Namespace != target.Namespace) {
		return PreparedAuthoringWork{}, workflowstate.ErrConflict
	}

	return prepared, err
}

// ApplyTelegramAuthoringWork uses the shared guarded owner and cannot admit work.
func (s service) ApplyTelegramAuthoringWork(ctx context.Context, target telegramcore.OperationTarget, prepared PreparedAuthoringWork, guard authoringcore.CommitGuard) (AuthoringWorkResult, error) {
	if prepared.Scope.ProjectID != target.ProjectID || prepared.Scope.Namespace != target.Namespace {
		return AuthoringWorkResult{}, workflowstate.ErrConflict
	}

	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return AuthoringWorkResult{}, err
	}

	return scoped.ApplyPreparedAuthoringWorkGuarded(ctx, prepared, guard)
}
