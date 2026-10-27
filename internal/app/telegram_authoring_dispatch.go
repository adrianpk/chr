package app

import (
	"context"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (r telegramReceiver) respondAuthoring(ctx context.Context, user telegramcore.TelegramAllowedUser, config telegramcore.TelegramReceiverConfig, update telegramcore.TelegramReceivedUpdate) (bool, error) {
	retained, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
	}

	if found && retained.OperationID == "authoring.continue" {
		return true, r.continueTelegramAuthoring(ctx, retained, config)
	}

	command, recognized, err := parseTelegramAuthoring(update)
	if !recognized {
		return r.respondAuthoringReply(ctx, user, config, update)
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, telegramAuthoringUsage(update))
	}

	command.Input.Config = config

	command.Input.Projects, err = r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	return true, r.dispatchAuthoring(ctx, user, command)
}

func (r telegramReceiver) respondAuthoringReply(ctx context.Context, user telegramcore.TelegramAllowedUser, config telegramcore.TelegramReceiverConfig, update telegramcore.TelegramReceivedUpdate) (bool, error) {
	if update.Kind != "message" || update.ReplyToMessageID == 0 || strings.HasPrefix(strings.TrimSpace(update.Text), "/") {
		return false, nil
	}

	if work, err := r.store.IsTelegramWorkReply(ctx, update); work || err != nil {
		return false, err
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	return r.dispatchAuthoringConversation(ctx, user, telegramcore.TelegramConversationInput{Update: update, Config: config, Projects: projects, Action: telegramcore.TelegramConversationCapture, Text: update.Text})
}

func (r telegramReceiver) dispatchAuthoringConversation(ctx context.Context, user telegramcore.TelegramAllowedUser, input telegramcore.TelegramConversationInput) (bool, error) {
	command := telegramAuthoringCommand{Input: input}
	switch input.Action {
	case telegramcore.TelegramConversationDone:
		command.Action, command.Target = core.AuthoringContinuationGenerate, core.AuthoringArtifactKindDump
	case telegramcore.TelegramConversationCapture:
		c, _, err := r.store.ResolveTelegramAuthoringTarget(ctx, input)
		if err != nil {
			return true, r.recordDiagnostic(ctx, input.Update, input.Config, telegramAuthoringFailure(err))
		}

		if c.AuthoringSessionRef == "" || c.QuestionBinding == "" {
			return false, nil
		}

		command.Action = core.AuthoringContinuationAnswer
	default:
		return false, nil
	}

	return true, r.dispatchAuthoring(ctx, user, command)
}

func (r telegramReceiver) dispatchAuthoring(ctx context.Context, user telegramcore.TelegramAllowedUser, command telegramAuthoringCommand) error {
	operations, ok := r.operations.(telegramAuthoringService)
	if !ok {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Shared authoring is unavailable. Retained context is unchanged.")
	}

	request, err := r.reserveAuthoring(ctx, operations, user, command)
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, telegramAuthoringFailure(err))
	}

	return r.continueTelegramAuthoring(ctx, request, command.Input.Config)
}
