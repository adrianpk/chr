package app

import (
	"encoding/json"
	"fmt"
	"os"

	"forge.adrianpk.com/chr/internal/core"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramNotificationTarget(project commandProjectContext, config notifycore.TelegramConfig) (notifycore.TelegramTarget, error) {
	token := os.Getenv(config.BotTokenEnv)
	if token == "" {
		return notifycore.TelegramTarget{}, fmt.Errorf("telegram bot token is required")
	}

	if config.ChatID == "" {
		return notifycore.TelegramTarget{}, fmt.Errorf("telegram chat id is required")
	}

	payload, err := json.Marshal(struct {
		Config     notifycore.TelegramConfig
		SenderHash string
	}{config, recordDigest([]byte(token))})
	if err != nil {
		return notifycore.TelegramTarget{}, err
	}

	return notifycore.ResolveTelegramTarget(projectcore.Project{ID: project.ProjectID, Path: project.ProjectDir, Enabled: true}, core.DefaultOpsNamespace, config.ChatID, recordDigest(payload))
}

func telegramNotificationInput(record notifycore.AuditRecord, project commandProjectContext, config notifycore.TelegramConfig) (notifycore.TelegramInput, error) {
	input := notifycore.TelegramInput{Record: record}
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
	if record.Event.Type == notifycore.EventManualTest {
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

func telegramNotificationEventBytes(event notifycore.Event) int {
	total := 0
	for _, value := range []string{event.ID, event.ProjectID, string(event.Type), string(event.Source), event.Payload.Summary, event.Payload.Detail, event.FailureReason, event.References.TicketKey, event.References.ExecutionID, event.References.LoopRunID} {
		total += len(value)
	}

	return total
}
