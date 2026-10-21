package app

import (
	"context"
	"errors"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func createTelegramNotification(ctx context.Context, repository core.NotificationEventRepository, record core.NotificationAuditRecord, project commandProjectContext, config core.TelegramNotificationConfig) (core.TelegramNotificationTarget, error) {
	input, err := telegramNotificationInput(record, project, config)
	if err == nil {
		err = repository.CreateTelegramNotification(ctx, input)
	}

	if err == nil {
		return input.Target, nil
	}
	// A failed notification presentation cannot undo completed domain work. Keep a
	// truthful failed audit record, without a partial view or a different send.
	code, _ := core.TelegramNotificationFailure(err)
	record.Event.Status = core.NotificationStatusFailed

	record.Event.FailureReason = code
	if saveErr := repository.CreateNotificationEvent(ctx, record); saveErr != nil {
		return core.TelegramNotificationTarget{}, errors.Join(err, saveErr)
	}

	return core.TelegramNotificationTarget{}, err
}

func dispatchTelegramNotification(ctx context.Context, repository core.NotificationEventRepository, record core.NotificationAuditRecord, project commandProjectContext, target core.TelegramNotificationTarget) (core.TelegramSendResult, error) {
	delivery, frozen, err := repository.ReadTelegramNotification(ctx, record.Event.ID, target)
	if err != nil {
		code, uncertain := core.TelegramNotificationFailure(err)
		saveErr := repository.FinishTelegramNotification(ctx, record.Event.ID, target, core.TelegramDocumentAcknowledgement{}, code, uncertain, time.Now().UTC())

		return core.TelegramSendResult{}, errors.Join(err, saveErr)
	}

	if delivery.Status != core.NotificationStatusPending {
		return core.TelegramSendResult{}, nil
	}

	if !frozen {
		return dispatchLegacyTelegramNotification(ctx, repository, record, project)
	}

	reserved, err := repository.ReserveTelegramNotification(ctx, record.Event.ID, target, false)
	if err != nil || !reserved {
		return core.TelegramSendResult{}, err
	}

	ack, sendErr := sendTelegramNotificationWithFallback(ctx, repository, record.Event.ID, project, target, delivery)

	code, uncertain := "", false
	if sendErr != nil {
		code, uncertain = core.TelegramNotificationFailure(sendErr)
	}

	if err := repository.FinishTelegramNotification(ctx, record.Event.ID, target, ack, code, uncertain, time.Now().UTC()); err != nil {
		return core.TelegramSendResult{}, err
	}

	return core.TelegramSendResult{StatusCode: 200}, sendErr
}

func sendTelegramNotificationWithFallback(ctx context.Context, repository core.NotificationEventRepository, id string, project commandProjectContext, target core.TelegramNotificationTarget, delivery core.TelegramNotificationDelivery) (core.TelegramDocumentAcknowledgement, error) {
	ack, sendErr := sendTelegramNotification(ctx, project, target, delivery)

	var apiFailure *core.TelegramAPIError
	if errors.As(sendErr, &apiFailure) && apiFailure.Code == "format_rejected" {
		fallback, fallbackErr := repository.ReserveTelegramNotification(ctx, id, target, true)
		if fallbackErr != nil {
			return core.TelegramDocumentAcknowledgement{}, fallbackErr
		}

		if fallback {
			delivery.PlainFallback = true
			ack, sendErr = sendTelegramNotification(ctx, project, target, delivery)
		}
	}

	return ack, sendErr
}

func sendTelegramNotification(ctx context.Context, project commandProjectContext, target core.TelegramNotificationTarget, delivery core.TelegramNotificationDelivery) (core.TelegramDocumentAcknowledgement, error) {
	config, err := core.LoadWorkspaceConfig(project.ProjectDir)
	if err != nil {
		return core.TelegramDocumentAcknowledgement{}, core.ErrTelegramNotificationTargetUnavailable
	}

	current, err := telegramNotificationTarget(project, config.Notifications.Telegram)
	if err != nil || current != target {
		return core.TelegramDocumentAcknowledgement{}, core.ErrTelegramNotificationTargetUnavailable
	}

	return notifyTelegramSender.SendPresentation(ctx, os.Getenv(config.Notifications.Telegram.BotTokenEnv), target.Destination, delivery)
}

func dispatchLegacyTelegramNotification(ctx context.Context, repository core.NotificationEventRepository, record core.NotificationAuditRecord, project commandProjectContext) (core.TelegramSendResult, error) {
	record, err := repository.GetNotificationEvent(ctx, record.Event.ID)
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

	record.Event.Status = core.NotificationStatusFailed
	if sendErr != nil {
		record.Event.FailureReason, _ = core.TelegramNotificationFailure(sendErr)
	} else {
		record.Event.Status = core.NotificationStatusDelivered
		record.Event.DeliveredAt = &now
	}

	return result, errors.Join(sendErr, repository.UpdateNotificationEvent(ctx, record))
}
