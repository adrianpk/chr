package app

import (
	"context"
	"errors"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) conversationGenerationScope(ctx context.Context, input ConversationIntentContext) (service, error) {
	if s.env == nil {
		return s, errors.New("conversation generation requires its host environment")
	}

	s.generationArtifactsDir = s.env.StateDir()
	if input.SelectedProject != "" && s.projectContext.ProjectID != input.SelectedProject {
		if !conversationGenerationProjectAllowed(input.SelectedProject, input) {
			return s, workflowstate.ErrConflict
		}

		return s.telegramProjectService(ctx, telegramcore.OperationTarget{ProjectID: input.SelectedProject})
	}

	if s.projectContext.ProjectDir == "" {
		// Before any project is selected the existing host conversation uses its
		// legacy provider. The same artifact owner survives later selection changes.
		s.projectContext.ProjectDir = s.env.StateDir()
	}

	return s, nil
}

func conversationGenerationProjectAllowed(id string, input ConversationIntentContext) bool {
	if id == "" {
		return true
	}

	for _, project := range input.Projects {
		if project.ID == id {
			return true
		}
	}

	return false
}

func (s service) validateIntentGenerationProject(ctx context.Context, location core.HarnessGenerationLocation, input ConversationIntentContext) error {
	if !conversationGenerationProjectAllowed(location.ProjectRef, input) {
		return workflowstate.ErrConflict
	}

	if location.ProjectRef == "" {
		return nil
	}

	current, err := s.telegramProjectService(ctx, telegramcore.OperationTarget{ProjectID: location.ProjectRef})
	if err != nil {
		return err
	}

	root, err := filepath.Abs(current.projectContext.ProjectDir)
	if err != nil || root != location.ConfigDir {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s service) intentGenerationProject(ctx context.Context, input ConversationIntentContext) (service, error) {
	location, bound := core.HarnessGenerationLocationFromContext(ctx)
	if !bound {
		return s, nil
	}

	if err := s.validateIntentGenerationProject(ctx, location, input); err != nil {
		return s, err
	}

	s.projectContext.ProjectDir = location.ExecutionDir

	return s, nil
}
