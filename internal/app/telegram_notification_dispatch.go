package app

import (
	"context"
	"errors"
	"os"
	"time"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
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
	code, _ := telegramcore.NotificationFailure(err)
	record.Event.Status = notifycore.StatusFailed

	record.Event.FailureReason = code
	if saveErr := repository.Create(ctx, record); saveErr != nil {
		return notifycore.TelegramTarget{}, errors.Join(err, saveErr)
	}

	return notifycore.TelegramTarget{}, err
}

func dispatchTelegramNotification(ctx context.Context, repository notifycore.Repository, record notifycore.AuditRecord, project commandProjectContext, target notifycore.TelegramTarget) (telegramcore.SendResult, error) {
	lookup, err := repository.ReadTelegram(ctx, record.Event.ID, target)
	delivery, frozen := lookup.Delivery, lookup.Frozen

	if err != nil {
		code, uncertain := telegramcore.NotificationFailure(err)
		saveErr := repository.FinishTelegram(ctx, notifycore.TelegramCompletion{ID: record.Event.ID, Target: target, Acknowledgement: tgmessage.Acknowledgement{}, Failure: notifycore.DeliveryFailure{Code: code, Uncertain: uncertain}, At: time.Now().UTC()})

		return telegramcore.SendResult{}, errors.Join(err, saveErr)
	}

	if delivery.Status != notifycore.StatusPending {
		return telegramcore.SendResult{}, nil
	}

	if !frozen {
		return dispatchLegacyTelegramNotification(ctx, repository, record, project)
	}

	reserved, err := repository.ReserveTelegram(ctx, record.Event.ID, target, false)
	if err != nil || !reserved {
		return telegramcore.SendResult{}, err
	}

	ack, sendErr := sendTelegramNotificationWithFallback(ctx, repository, telegramNotificationAttempt{ID: record.Event.ID, Project: project, Target: target, Delivery: delivery})

	code, uncertain := "", false
	if sendErr != nil {
		code, uncertain = telegramcore.NotificationFailure(sendErr)
	}

	if err := repository.FinishTelegram(ctx, notifycore.TelegramCompletion{ID: record.Event.ID, Target: target, Acknowledgement: ack, Failure: notifycore.DeliveryFailure{Code: code, Uncertain: uncertain}, At: time.Now().UTC()}); err != nil {
		return telegramcore.SendResult{}, err
	}

	return telegramcore.SendResult{StatusCode: 200}, sendErr
}

func sendTelegramNotificationWithFallback(ctx context.Context, repository notifycore.Repository, attempt telegramNotificationAttempt) (tgmessage.Acknowledgement, error) {
	ack, sendErr := sendTelegramNotification(ctx, attempt.Project, attempt.Target, attempt.Delivery)

	var apiFailure *telegramcore.APIError
	if errors.As(sendErr, &apiFailure) && apiFailure.Code == "format_rejected" {
		fallback, fallbackErr := repository.ReserveTelegram(ctx, attempt.ID, attempt.Target, true)
		if fallbackErr != nil {
			return tgmessage.Acknowledgement{}, fallbackErr
		}

		if fallback {
			attempt.Delivery.PlainFallback = true
			ack, sendErr = sendTelegramNotification(ctx, attempt.Project, attempt.Target, attempt.Delivery)
		}
	}

	return ack, sendErr
}

func sendTelegramNotification(ctx context.Context, project commandProjectContext, target notifycore.TelegramTarget, delivery notifycore.TelegramDelivery) (tgmessage.Acknowledgement, error) {
	config, err := workspaceconfig.Load(project.ProjectDir)
	if err != nil {
		return tgmessage.Acknowledgement{}, notifycore.ErrTelegramTargetUnavailable
	}

	current, err := telegramNotificationTarget(project, config.Notifications.Telegram)
	if err != nil || current != target {
		return tgmessage.Acknowledgement{}, notifycore.ErrTelegramTargetUnavailable
	}

	return notifyTelegramSender.SendPresentation(ctx, os.Getenv(config.Notifications.Telegram.BotTokenEnv), target.Destination, delivery)
}

func dispatchLegacyTelegramNotification(ctx context.Context, repository notifycore.Repository, record notifycore.AuditRecord, project commandProjectContext) (telegramcore.SendResult, error) {
	record, err := repository.Get(ctx, record.Event.ID)
	if err != nil {
		return telegramcore.SendResult{}, err
	}

	config, err := workspaceconfig.Load(project.ProjectDir)
	if err != nil {
		return telegramcore.SendResult{}, err
	}

	message := telegramcore.RenderNotificationMessage(record.Event)
	message.ChatID = config.Notifications.Telegram.ChatID
	result, sendErr := notifyTelegramSender.SendMessage(ctx, os.Getenv(config.Notifications.Telegram.BotTokenEnv), message)
	now := time.Now().UTC()
	record.Event.UpdatedAt = now

	record.Event.Status = notifycore.StatusFailed
	if sendErr != nil {
		record.Event.FailureReason, _ = telegramcore.NotificationFailure(sendErr)
	} else {
		record.Event.Status = notifycore.StatusDelivered
		record.Event.DeliveredAt = &now
	}

	return result, errors.Join(sendErr, repository.Update(ctx, record))
}

type telegramNotificationAttempt struct {
	ID       string
	Project  commandProjectContext
	Target   notifycore.TelegramTarget
	Delivery notifycore.TelegramDelivery
}
