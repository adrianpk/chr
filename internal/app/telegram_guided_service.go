package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramGuidedService interface {
	InspectTelegramGuided(context.Context, core.TelegramOperationTarget, string) (GuidedUseCaseInspection, error)
	ReadTelegramGuidedOriginal(context.Context, core.TelegramOperationTarget, string, string) (GuidedUseCaseArtifact, error)
	ApplyTelegramGuided(context.Context, core.TelegramOperationTarget, telegramGuidedInput) (core.AuthoringSession, error)
}

type telegramGuidedInput struct {
	Document string                    `json:"document,omitempty"`
	Name     string                    `json:"name,omitempty"`
	Choices  []string                  `json:"choices,omitempty"`
	Actor    string                    `json:"actor"`
	Inspect  bool                      `json:"inspect,omitempty"`
	Retry    bool                      `json:"retry,omitempty"`
	Offer    bool                      `json:"offer,omitempty"`
	Start    bool                      `json:"start,omitempty"`
	Command  core.GuidedUseCaseCommand `json:"command"`
}

func (s service) telegramGuidedProject(ctx context.Context, target core.TelegramOperationTarget) (service, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return scoped, err
	}

	project := core.Project{ID: scoped.projectContext.ProjectID, Path: scoped.projectContext.ProjectDir, Enabled: true}

	return scoped, core.ValidateTelegramProjectTarget(project, target)
}

// InspectTelegramGuided reads the canonical run after checking registry identity.
func (s service) InspectTelegramGuided(ctx context.Context, target core.TelegramOperationTarget, id string) (GuidedUseCaseInspection, error) {
	scoped, err := s.telegramGuidedProject(ctx, target)
	if err != nil {
		return GuidedUseCaseInspection{}, err
	}

	return scoped.InspectGuidedUseCase(ctx, id, target.Namespace)
}

// ApplyTelegramGuided maps a frozen transport request to shared guided commands.
func (s service) ApplyTelegramGuided(ctx context.Context, target core.TelegramOperationTarget, input telegramGuidedInput) (core.AuthoringSession, error) {
	command := input.Command
	if command.ProjectID != target.ProjectID || command.Namespace != target.Namespace || command.RunID == "" {
		return core.AuthoringSession{}, core.ErrWorkflowConflict
	}

	scoped, err := s.telegramGuidedProject(ctx, target)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	if input.Inspect {
		observed, err := scoped.InspectGuidedUseCase(ctx, command.RunID, command.Namespace)
		if err != nil {
			return core.AuthoringSession{}, err
		}

		return *observed.Session, nil
	}

	if input.Start {
		return scoped.StartGuidedUseCase(ctx, command.RunID, command.Namespace, input.Actor)
	}

	return scoped.ChangeGuidedUseCase(ctx, command)
}

// ReadTelegramGuidedOriginal checks the registered scope before reading sealed bytes.
func (s service) ReadTelegramGuidedOriginal(ctx context.Context, target core.TelegramOperationTarget, id, kind string) (GuidedUseCaseArtifact, error) {
	scoped, err := s.telegramGuidedProject(ctx, target)
	if err != nil {
		return GuidedUseCaseArtifact{}, err
	}

	return scoped.ReadGuidedUseCaseOriginal(ctx, id, target.Namespace, kind)
}
