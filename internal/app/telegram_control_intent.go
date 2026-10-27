package app

import (
	"context"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (r telegramReceiver) dispatchIntentControl(ctx context.Context, user telegramcore.AllowedUser, input telegramcore.ConversationInput, intent ConversationIntent) (bool, error) {
	parsedControl, err := telegramIntentControl(input.Update, intent)

	command, recognized := parsedControl.Command, parsedControl.Recognized
	if !recognized {
		return false, nil
	}

	if err != nil {
		return true, r.intentCorrection(ctx, input.Update, input.Config)
	}

	command.Input.Config, command.Input.Projects = input.Config, input.Projects
	command.Input.ContextDigest = input.ContextDigest

	return true, r.dispatchControl(ctx, user, command)
}

func (r telegramReceiver) intentControlTarget(ctx context.Context, input telegramcore.ConversationInput, intent ConversationIntent) (telegramTargetResolution, error) {
	parsedControl2, err := telegramIntentControl(input.Update, intent)

	command, recognized := parsedControl2.Command, parsedControl2.Recognized
	if !recognized || err != nil {
		return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: recognized}, err
	}

	command.Input.Config, command.Input.Projects = input.Config, input.Projects
	command.Input.ContextDigest = input.ContextDigest
	target, err := r.resolveControlTarget(ctx, command)

	return telegramTargetResolution{Target: target, Needed: true}, err
}

func (r telegramReceiver) dispatchIntentNativeWork(ctx context.Context, user telegramcore.AllowedUser, input telegramcore.ConversationInput, intent ConversationIntent) (bool, error) {
	if handled, err := r.dispatchIntentControl(ctx, user, input, intent); handled || err != nil {
		return handled, err
	}

	if parsedWork, err := telegramIntentWork(input.Update, intent); parsedWork.Recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		parsedWork.Command.Input.Config, parsedWork.Command.Input.Projects, parsedWork.Command.Input.Namespace = input.Config, input.Projects, input.Namespace
		parsedWork.Command.Input.ContextDigest = input.ContextDigest

		return true, r.dispatchWork(ctx, user, parsedWork.Command)
	}

	return false, nil
}

func (r telegramReceiver) intentNativeWorkTarget(ctx context.Context, input telegramcore.ConversationInput, intent ConversationIntent) (telegramTargetResolution, error) {
	if targetResolution, err := r.intentControlTarget(ctx, input, intent); targetResolution.Needed || err != nil {
		return telegramTargetResolution{Target: targetResolution.Target, Needed: targetResolution.Needed}, err
	}

	if parsedWork2, err := telegramIntentWork(input.Update, intent); parsedWork2.Recognized {
		if err != nil {
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: true}, err
		}

		parsedWork2.Command.Input.Config, parsedWork2.Command.Input.Projects = input.Config, input.Projects
		parsedWork2.Command.Input.ContextDigest = input.ContextDigest
		target, err := r.store.ResolveOperationTarget(ctx, parsedWork2.Command.Input, time.Now())

		return telegramTargetResolution{Target: target, Needed: true}, err
	}

	return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, nil
}
