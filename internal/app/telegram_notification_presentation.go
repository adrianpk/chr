package app

import notifycore "forge.adrianpk.com/chr/internal/core/notify"

import (
	"encoding/json"
	"fmt"
	"os"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramNotificationTarget(project commandProjectContext, config notifycore.TelegramNotificationConfig) (notifycore.TelegramNotificationTarget, error) {
	token := os.Getenv(config.BotTokenEnv)
	if token == "" {
		return notifycore.TelegramNotificationTarget{}, fmt.Errorf("telegram bot token is required")
	}

	if config.ChatID == "" {
		return notifycore.TelegramNotificationTarget{}, fmt.Errorf("telegram chat id is required")
	}

	payload, err := json.Marshal(struct {
		Config     notifycore.TelegramNotificationConfig
		SenderHash string
	}{config, recordDigest([]byte(token))})
	if err != nil {
		return notifycore.TelegramNotificationTarget{}, err
	}

	return notifycore.ResolveTelegramNotificationTarget(projectcore.Project{ID: project.ProjectID, Path: project.ProjectDir, Enabled: true}, core.DefaultOpsNamespace, config.ChatID, recordDigest(payload))
}

func telegramNotificationInput(record notifycore.NotificationAuditRecord, project commandProjectContext, config notifycore.TelegramNotificationConfig) (notifycore.TelegramNotificationInput, error) {
	input := notifycore.TelegramNotificationInput{Record: record}
	if telegramNotificationEventBytes(record.Event) > 2<<20 {
		return input, tgmessage.ErrDocumentUnavailable
	}

	target, err := telegramNotificationTarget(project, config)
	if err != nil {
		return input, err
	}

	input.Target = target

	name := project.ProjectName
	if name == "" {
		name = project.ProjectID
	}

	input.Message = tgmessage.Blocks{Family: "47", InitialText: record.Event.Payload.Summary + " in " + name, Content: record.Event.Payload.Detail, SecondaryText: "This reports work activity. Integration requires its canonical receipt. Complete notification event export attached."}
	if record.Event.Type == notifycore.NotificationEventManualTest {
		input.Message = tgmessage.Blocks{Family: "48", InitialText: "Telegram notification test for " + name + "."}

		return input, nil
	}

	payload, err := json.MarshalIndent(record.Event, "", "  ")
	if err != nil {
		return input, err
	}

	payload = append(payload, '\n')
	document := tgmessage.Document{ProjectID: target.ProjectID, Namespace: target.Namespace, ProjectBindingDigest: target.ProjectBindingDigest, SourceRef: "exports/notifications/" + record.Event.ID + ".json", SourceHash: recordDigest(payload), Filename: "notification-event.json", MediaType: "application/json", Payload: payload}
	input.Document = &document

	return input, nil
}

func telegramNotificationEventBytes(event notifycore.NotificationEvent) int {
	total := 0
	for _, value := range []string{event.ID, event.ProjectID, string(event.Type), string(event.Source), event.Payload.Summary, event.Payload.Detail, event.FailureReason, event.References.TicketKey, event.References.ExecutionID, event.References.LoopRunID} {
		total += len(value)
	}

	return total
}
