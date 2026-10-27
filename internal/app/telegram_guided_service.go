package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramGuidedService interface {
	InspectTelegramGuided(context.Context, telegramcore.TelegramOperationTarget, string) (GuidedUseCaseInspection, error)
	ReadTelegramGuidedOriginal(context.Context, telegramcore.TelegramOperationTarget, string, string) (GuidedUseCaseArtifact, error)
	ApplyTelegramGuided(context.Context, telegramcore.TelegramOperationTarget, telegramGuidedInput) (core.AuthoringSession, error)
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

func (s service) telegramGuidedProject(ctx context.Context, target telegramcore.TelegramOperationTarget) (service, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return scoped, err
	}

	project := projectcore.Project{ID: scoped.projectContext.ProjectID, Path: scoped.projectContext.ProjectDir, Enabled: true}

	return scoped, telegramcore.ValidateTelegramProjectTarget(project, target)
}

// InspectTelegramGuided reads the canonical run after checking registry identity.
func (s service) InspectTelegramGuided(ctx context.Context, target telegramcore.TelegramOperationTarget, id string) (GuidedUseCaseInspection, error) {
	scoped, err := s.telegramGuidedProject(ctx, target)
	if err != nil {
		return GuidedUseCaseInspection{}, err
	}

	return scoped.InspectGuidedUseCase(ctx, id, target.Namespace)
}

// ApplyTelegramGuided maps a frozen transport request to shared guided commands.
func (s service) ApplyTelegramGuided(ctx context.Context, target telegramcore.TelegramOperationTarget, input telegramGuidedInput) (core.AuthoringSession, error) {
	command := input.Command
	if command.ProjectID != target.ProjectID || command.Namespace != target.Namespace || command.RunID == "" {
		return core.AuthoringSession{}, workflowstate.ErrConflict
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
func (s service) ReadTelegramGuidedOriginal(ctx context.Context, target telegramcore.TelegramOperationTarget, id, kind string) (GuidedUseCaseArtifact, error) {
	scoped, err := s.telegramGuidedProject(ctx, target)
	if err != nil {
		return GuidedUseCaseArtifact{}, err
	}

	return scoped.ReadGuidedUseCaseOriginal(ctx, id, target.Namespace, kind)
}
