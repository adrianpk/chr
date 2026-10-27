package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramExplanationService interface {
	PrepareTelegramExplanation(context.Context, telegramcore.OperationTarget, AuthoringExplanationRequest) (PreparedAuthoringExplanation, error)
	ExplainPreparedTelegram(context.Context, telegramcore.OperationTarget, PreparedAuthoringExplanation) (AuthoringExplanation, error)
	ReadPreparedTelegramDocument(context.Context, telegramcore.OperationTarget, PreparedAuthoringExplanation) (core.AuthoringArtifactObservation, error)
}

// PrepareTelegramExplanation maps registry routing to the neutral source reader.
func (s service) PrepareTelegramExplanation(ctx context.Context, target telegramcore.OperationTarget, request AuthoringExplanationRequest) (PreparedAuthoringExplanation, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return PreparedAuthoringExplanation{}, err
	}

	return scoped.PrepareAuthoringExplanation(ctx, request)
}

// ExplainPreparedTelegram uses the same configured provider as other adapters.
func (s service) ExplainPreparedTelegram(ctx context.Context, target telegramcore.OperationTarget, prepared PreparedAuthoringExplanation) (AuthoringExplanation, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return AuthoringExplanation{}, err
	}

	return scoped.ExplainPreparedAuthoring(ctx, prepared)
}

// ReadPreparedTelegramDocument resolves the current registry before reading the original.
func (s service) ReadPreparedTelegramDocument(ctx context.Context, target telegramcore.OperationTarget, prepared PreparedAuthoringExplanation) (core.AuthoringArtifactObservation, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return core.AuthoringArtifactObservation{}, err
	}

	return scoped.ReadPreparedAuthoringDocument(ctx, prepared)
}

func (s service) telegramBoundProjectService(ctx context.Context, target telegramcore.OperationTarget) (service, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return scoped, err
	}

	project := projectcore.Project{ID: scoped.projectContext.ProjectID, Path: scoped.projectContext.ProjectDir, Enabled: true}
	if err := telegramcore.ValidateProjectTarget(project, target); err != nil {
		return scoped, err
	}

	return scoped, nil
}
