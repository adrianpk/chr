package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func prepareTelegramGuidedInput(ctx context.Context, operations telegramGuidedService, target telegramcore.OperationTarget, input telegramGuidedInput) (telegramGuidedInput, error) {
	observed, err := operations.InspectTelegramGuided(ctx, target, input.Command.RunID)
	if err != nil {
		return input, err
	}

	if input.Retry {
		command, ready := guidedControlCommand(*observed.Session)
		if !ready {
			command, ready = guidedProgressCommand(*observed.Session)
		}

		if !ready {
			return input, workflowstate.ErrConflict
		}

		input.Command = command
	}

	if input.Offer {
		input.Name = observed.Session.Topic

		input.Command.ExpectedRevision = observed.Session.Revision
		if observed.Work != nil {
			input.Command.NativeRevision = observed.Work.Revision
		}

		input.Choices = guidedControlChoices(observed)
		if len(input.Choices) == 0 {
			return input, workflowstate.ErrConflict
		}
	}

	return input, nil
}

func telegramGuidedOwner(c telegramcore.Conversation, input telegramcore.ConversationInput, start bool, actor string) (telegramGuidedInput, error) {
	if start {
		return telegramGuidedInput{Actor: actor, Start: true, Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: c.ID, ProjectID: c.ProjectID, Namespace: c.Namespace}}}, nil
	}

	if input.Conversation != "" {
		return telegramGuidedInput{Actor: actor, Inspect: true, Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: c.GuidedRunID, ProjectID: c.ProjectID, Namespace: c.Namespace}, OperationID: telegramcore.OperationID(input.Update, 0)}}, nil
	}

	return telegramGuidedCommand(c, input.Update, actor)
}
