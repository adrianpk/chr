package app

import (
	"context"
	"errors"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) conversationGenerationScope(ctx context.Context, input ConversationIntentContext) (service, error) {
	if s.env == nil {
		return s, errors.New("conversation generation requires its host environment")
	}

	s.generationArtifactsDir = s.env.StateDir()
	if input.SelectedProject != "" && s.projectContext.ProjectID != input.SelectedProject {
		if !conversationGenerationProjectAllowed(input.SelectedProject, input) {
			return s, core.ErrWorkflowConflict
		}

		return s.telegramProjectService(ctx, core.TelegramOperationTarget{ProjectID: input.SelectedProject})
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

func (s service) bindIntentGeneration(ctx context.Context, requestID string, input ConversationIntentContext) (service, context.Context, error) {
	scoped, err := s.conversationGenerationScope(ctx, input)
	if err != nil {
		return s, ctx, err
	}

	s, ctx, err = scoped.bindGeneration(ctx, "intent-"+requestID, "conversation")
	if err != nil {
		return s, ctx, err
	}

	if location, bound := core.HarnessGenerationLocationFromContext(ctx); bound {
		if err := s.validateIntentGenerationProject(ctx, location, input); err != nil {
			return s, ctx, err
		}

		s.projectContext.ProjectDir = location.ExecutionDir
	}

	return s, ctx, nil
}

func (s service) validateIntentGenerationProject(ctx context.Context, location core.HarnessGenerationLocation, input ConversationIntentContext) error {
	if !conversationGenerationProjectAllowed(location.ProjectRef, input) {
		return core.ErrWorkflowConflict
	}

	if location.ProjectRef == "" {
		return nil
	}

	current, err := s.telegramProjectService(ctx, core.TelegramOperationTarget{ProjectID: location.ProjectRef})
	if err != nil {
		return err
	}

	root, err := filepath.Abs(current.projectContext.ProjectDir)
	if err != nil || root != location.ConfigDir {
		return core.ErrWorkflowConflict
	}

	return nil
}
