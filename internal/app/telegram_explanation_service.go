package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramExplanationService interface {
	PrepareTelegramExplanation(context.Context, core.TelegramOperationTarget, AuthoringExplanationRequest) (PreparedAuthoringExplanation, error)
	ExplainPreparedTelegram(context.Context, core.TelegramOperationTarget, PreparedAuthoringExplanation) (AuthoringExplanation, error)
	ReadPreparedTelegramDocument(context.Context, core.TelegramOperationTarget, PreparedAuthoringExplanation) (core.AuthoringArtifactObservation, error)
}

// PrepareTelegramExplanation maps registry routing to the neutral source reader.
func (s service) PrepareTelegramExplanation(ctx context.Context, target core.TelegramOperationTarget, request AuthoringExplanationRequest) (PreparedAuthoringExplanation, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return PreparedAuthoringExplanation{}, err
	}

	return scoped.PrepareAuthoringExplanation(ctx, request)
}

// ExplainPreparedTelegram uses the same configured provider as other adapters.
func (s service) ExplainPreparedTelegram(ctx context.Context, target core.TelegramOperationTarget, prepared PreparedAuthoringExplanation) (AuthoringExplanation, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return AuthoringExplanation{}, err
	}

	return scoped.ExplainPreparedAuthoring(ctx, prepared)
}

// ReadPreparedTelegramDocument resolves the current registry before reading the original.
func (s service) ReadPreparedTelegramDocument(ctx context.Context, target core.TelegramOperationTarget, prepared PreparedAuthoringExplanation) (core.AuthoringArtifactObservation, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return core.AuthoringArtifactObservation{}, err
	}

	return scoped.ReadPreparedAuthoringDocument(ctx, prepared)
}

func (s service) telegramBoundProjectService(ctx context.Context, target core.TelegramOperationTarget) (service, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return scoped, err
	}

	project := core.Project{ID: scoped.projectContext.ProjectID, Path: scoped.projectContext.ProjectDir, Enabled: true}
	if err := core.ValidateTelegramProjectTarget(project, target); err != nil {
		return scoped, err
	}

	return scoped, nil
}
