package app

import (
	"context"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (r telegramReceiver) dispatchIntentControl(ctx context.Context, user telegramcore.TelegramAllowedUser, input telegramcore.TelegramConversationInput, intent ConversationIntent) (bool, error) {
	command, recognized, err := telegramIntentControl(input.Update, intent)
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

func (r telegramReceiver) intentControlTarget(ctx context.Context, input telegramcore.TelegramConversationInput, intent ConversationIntent) (telegramcore.TelegramOperationTarget, bool, error) {
	command, recognized, err := telegramIntentControl(input.Update, intent)
	if !recognized || err != nil {
		return telegramcore.TelegramOperationTarget{}, recognized, err
	}

	command.Input.Config, command.Input.Projects = input.Config, input.Projects
	command.Input.ContextDigest = input.ContextDigest
	target, err := r.resolveControlTarget(ctx, command)

	return target, true, err
}

func (r telegramReceiver) dispatchIntentNativeWork(ctx context.Context, user telegramcore.TelegramAllowedUser, input telegramcore.TelegramConversationInput, intent ConversationIntent) (bool, error) {
	if handled, err := r.dispatchIntentControl(ctx, user, input, intent); handled || err != nil {
		return handled, err
	}

	if work, recognized, err := telegramIntentWork(input.Update, intent); recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		work.Input.Config, work.Input.Projects, work.Input.Namespace = input.Config, input.Projects, input.Namespace
		work.Input.ContextDigest = input.ContextDigest

		return true, r.dispatchWork(ctx, user, work)
	}

	return false, nil
}

func (r telegramReceiver) intentNativeWorkTarget(ctx context.Context, input telegramcore.TelegramConversationInput, intent ConversationIntent) (telegramcore.TelegramOperationTarget, bool, error) {
	if target, needed, err := r.intentControlTarget(ctx, input, intent); needed || err != nil {
		return target, needed, err
	}

	if work, recognized, err := telegramIntentWork(input.Update, intent); recognized {
		if err != nil {
			return telegramcore.TelegramOperationTarget{}, true, err
		}

		work.Input.Config, work.Input.Projects = input.Config, input.Projects
		work.Input.ContextDigest = input.ContextDigest
		target, err := r.store.ResolveTelegramOperationTarget(ctx, work.Input, time.Now())

		return target, true, err
	}

	return telegramcore.TelegramOperationTarget{}, false, nil
}
