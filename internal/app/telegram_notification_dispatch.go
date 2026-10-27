package app

import (
	"context"
	"errors"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func createTelegramNotification(ctx context.Context, repository notifycore.Repository, record notifycore.AuditRecord, project commandProjectContext, config notifycore.TelegramConfig) (notifycore.TelegramTarget, error) {
	input, err := telegramNotificationInput(record, project, config)
	if err == nil {
		err = repository.CreateTelegram(ctx, input)
	}

	if err == nil {
		return input.Target, nil
	}
	// A failed notification presentation cannot undo completed domain work. Keep a
	// truthful failed audit record, without a partial view or a different send.
	code, _ := core.TelegramNotificationFailure(err)
	record.Event.Status = notifycore.StatusFailed

	record.Event.FailureReason = code
	if saveErr := repository.Create(ctx, record); saveErr != nil {
		return notifycore.TelegramTarget{}, errors.Join(err, saveErr)
	}

	return notifycore.TelegramTarget{}, err
}

func dispatchTelegramNotification(ctx context.Context, repository notifycore.Repository, record notifycore.AuditRecord, project commandProjectContext, target notifycore.TelegramTarget) (core.TelegramSendResult, error) {
	lookup, err := repository.ReadTelegram(ctx, record.Event.ID, target)
	delivery, frozen := lookup.Delivery, lookup.Frozen

	if err != nil {
		code, uncertain := core.TelegramNotificationFailure(err)
		saveErr := repository.FinishTelegram(ctx, notifycore.TelegramCompletion{ID: record.Event.ID, Target: target, Acknowledgement: tgmessage.Acknowledgement{}, Failure: notifycore.DeliveryFailure{Code: code, Uncertain: uncertain}, At: time.Now().UTC()})

		return core.TelegramSendResult{}, errors.Join(err, saveErr)
	}

	if delivery.Status != notifycore.StatusPending {
		return core.TelegramSendResult{}, nil
	}

	if !frozen {
		return dispatchLegacyTelegramNotification(ctx, repository, record, project)
	}

	reserved, err := repository.ReserveTelegram(ctx, record.Event.ID, target, false)
	if err != nil || !reserved {
		return core.TelegramSendResult{}, err
	}

	ack, sendErr := sendTelegramNotificationWithFallback(ctx, repository, record.Event.ID, project, target, delivery)

	code, uncertain := "", false
	if sendErr != nil {
		code, uncertain = core.TelegramNotificationFailure(sendErr)
	}

	if err := repository.FinishTelegram(ctx, notifycore.TelegramCompletion{ID: record.Event.ID, Target: target, Acknowledgement: ack, Failure: notifycore.DeliveryFailure{Code: code, Uncertain: uncertain}, At: time.Now().UTC()}); err != nil {
		return core.TelegramSendResult{}, err
	}

	return core.TelegramSendResult{StatusCode: 200}, sendErr
}

func sendTelegramNotificationWithFallback(ctx context.Context, repository notifycore.Repository, id string, project commandProjectContext, target notifycore.TelegramTarget, delivery notifycore.TelegramDelivery) (tgmessage.Acknowledgement, error) {
	ack, sendErr := sendTelegramNotification(ctx, project, target, delivery)

	var apiFailure *core.TelegramAPIError
	if errors.As(sendErr, &apiFailure) && apiFailure.Code == "format_rejected" {
		fallback, fallbackErr := repository.ReserveTelegram(ctx, id, target, true)
		if fallbackErr != nil {
			return tgmessage.Acknowledgement{}, fallbackErr
		}

		if fallback {
			delivery.PlainFallback = true
			ack, sendErr = sendTelegramNotification(ctx, project, target, delivery)
		}
	}

	return ack, sendErr
}

func sendTelegramNotification(ctx context.Context, project commandProjectContext, target notifycore.TelegramTarget, delivery notifycore.TelegramDelivery) (tgmessage.Acknowledgement, error) {
	config, err := core.LoadWorkspaceConfig(project.ProjectDir)
	if err != nil {
		return tgmessage.Acknowledgement{}, notifycore.ErrTelegramTargetUnavailable
	}

	current, err := telegramNotificationTarget(project, config.Notifications.Telegram)
	if err != nil || current != target {
		return tgmessage.Acknowledgement{}, notifycore.ErrTelegramTargetUnavailable
	}

	return notifyTelegramSender.SendPresentation(ctx, os.Getenv(config.Notifications.Telegram.BotTokenEnv), target.Destination, delivery)
}

func dispatchLegacyTelegramNotification(ctx context.Context, repository notifycore.Repository, record notifycore.AuditRecord, project commandProjectContext) (core.TelegramSendResult, error) {
	record, err := repository.Get(ctx, record.Event.ID)
	if err != nil {
		return core.TelegramSendResult{}, err
	}

	config, err := core.LoadWorkspaceConfig(project.ProjectDir)
	if err != nil {
		return core.TelegramSendResult{}, err
	}

	message := core.RenderTelegramNotificationMessage(record.Event)
	message.ChatID = config.Notifications.Telegram.ChatID
	result, sendErr := notifyTelegramSender.SendMessage(ctx, os.Getenv(config.Notifications.Telegram.BotTokenEnv), message)
	now := time.Now().UTC()
	record.Event.UpdatedAt = now

	record.Event.Status = notifycore.StatusFailed
	if sendErr != nil {
		record.Event.FailureReason, _ = core.TelegramNotificationFailure(sendErr)
	} else {
		record.Event.Status = notifycore.StatusDelivered
		record.Event.DeliveredAt = &now
	}

	return result, errors.Join(sendErr, repository.Update(ctx, record))
}
