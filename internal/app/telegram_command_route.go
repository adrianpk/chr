package app

import (
	"context"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// Each route retains its own operation identity, parser, diagnostic and resume
// policy. Only retained lookup, recognition and current input population are shared.
type telegramCommandRoute struct {
	OperationID string
	Continue    func(context.Context, telegramcore.TelegramOperationRequest, telegramcore.TelegramReceiverConfig) error
	Help        string
	Parse       func(telegramcore.TelegramReceivedUpdate) (telegramParsedCommand, error)
}

type telegramParsedCommand struct {
	Recognized bool
	Input      *telegramcore.TelegramConversationInput
	Dispatch   func(context.Context, telegramcore.TelegramAllowedUser) error
}

func (r telegramReceiver) respondCommand(ctx context.Context, user telegramcore.TelegramAllowedUser, config telegramcore.TelegramReceiverConfig, update telegramcore.TelegramReceivedUpdate, route telegramCommandRoute) (bool, error) {
	retained, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
	}

	if found && retained.OperationID == route.OperationID {
		return true, route.Continue(ctx, retained, config)
	}

	command, err := route.Parse(update)
	if !command.Recognized {
		return false, nil
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, route.Help)
	}

	command.Input.Config = config

	command.Input.Projects, err = r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	return true, command.Dispatch(ctx, user)
}
