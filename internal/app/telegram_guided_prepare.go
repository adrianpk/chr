package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

func prepareTelegramGuidedInput(ctx context.Context, operations telegramGuidedService, target core.TelegramOperationTarget, input telegramGuidedInput) (telegramGuidedInput, error) {
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
			return input, core.ErrWorkflowConflict
		}

		input.Command = command
	}

	if input.Offer {
		input.Name = observed.Session.Topic

		input.Command.ExpectedRevision = observed.Session.Revision
		if observed.Work != nil {
			input.Command.NativeRevision = observed.Work.Revision
		}

		input.Choices = telegramGuidedControlChoices(observed)
		if len(input.Choices) == 0 {
			return input, core.ErrWorkflowConflict
		}
	}

	return input, nil
}

func telegramGuidedControlChoices(observed GuidedUseCaseInspection) []string {
	session := *observed.Session

	state := session.Guided
	if guidedWebControlTerminal(state.Status, observed.Work) {
		return nil
	}

	if _, ready := guidedControlCommand(session); ready {
		return nil
	}

	if state.Control != nil && state.Control.StopIntent != "" {
		if state.Status == "paused" {
			return []string{"resume", "cancel"}
		}

		return []string{"cancel"}
	}

	choices := []string{"pause", "cancel"}
	if state.Stage == "dump" && state.PendingOperation == nil {
		choices = append([]string{"finish"}, choices...)
	}

	return choices
}

func telegramGuidedOwner(c core.TelegramConversation, input core.TelegramConversationInput, start bool, actor string) (telegramGuidedInput, error) {
	if start {
		return telegramGuidedInput{Actor: actor, Start: true, Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: c.ID, ProjectID: c.ProjectID, Namespace: c.Namespace}}}, nil
	}

	if input.Conversation != "" {
		return telegramGuidedInput{Actor: actor, Inspect: true, Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: c.GuidedRunID, ProjectID: c.ProjectID, Namespace: c.Namespace}, OperationID: core.TelegramOperationID(input.Update, 0)}}, nil
	}

	return telegramGuidedCommand(c, input.Update, actor)
}
