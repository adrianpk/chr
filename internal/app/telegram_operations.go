package app

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramReadOperation int

const (
	telegramReadHelp telegramReadOperation = iota
	telegramReadProjects
	telegramReadStatus
)

type telegramOperationDescriptor struct {
	ID      telegramReadOperation
	Command string
	Help    string
}

func telegramReadRegistry() []telegramOperationDescriptor {
	return []telegramOperationDescriptor{
		{telegramReadHelp, "/help", "Show available commands"},
		{telegramReadProjects, "/projects", "List your enabled registered projects"},
		{telegramReadStatus, "/status", "Show receiver health and retained polling state"},
	}
}

type telegramReadService interface {
	ListRegisteredProjects(ids []string) ([]core.Project, error)
	InspectTelegramReceiver(botID int64) (core.TelegramReceiverState, error)
}

func telegramReadResponse(operations telegramReadService, user core.TelegramAllowedUser, update core.TelegramReceivedUpdate) (telegramReadReply, error) {
	if update.Kind == "callback_query" {
		return telegramReadReply{Text: "This action is unavailable. Use /help for current commands."}, nil
	}

	for _, operation := range telegramReadRegistry() {
		if strings.TrimSpace(update.Text) != operation.Command {
			continue
		}

		return telegramReadOperationResponse(operations, user, update, operation.ID)
	}

	return telegramReadReply{Text: "I could not use that command. Send /help to see available commands."}, nil
}

func telegramReadOperationResponse(operations telegramReadService, user core.TelegramAllowedUser, update core.TelegramReceivedUpdate, operation telegramReadOperation) (telegramReadReply, error) {
	switch operation {
	case telegramReadProjects:
		text, err := telegramProjectResponse(operations, user.ProjectIDs)

		return telegramReadReply{Text: text}, err
	case telegramReadStatus:
		state, err := operations.InspectTelegramReceiver(update.Peer.BotID)
		if err != nil {
			return telegramReadReply{}, err
		}

		return telegramReadReply{Text: telegramReceiverHealthText(state)}, nil
	case telegramReadHelp:
		return telegramBasicHelp(), nil
	}

	return telegramReadReply{Text: "I could not use that command. Send /help to see available commands."}, nil
}

func telegramProjectResponse(operations telegramReadService, ids []string) (string, error) {
	projects, err := operations.ListRegisteredProjects(ids)
	if err != nil {
		return "", err
	}

	var lines []string

	for _, project := range projects {
		if project.Enabled {
			lines = append(lines, "- "+project.Name)
		}
	}

	if len(lines) == 0 {
		return "No projects are available to you. Project access must be configured locally.", nil
	}

	return core.RenderTelegramMessage(core.TelegramMessageBlocks{InitialText: "Available projects", ContentKind: "list", Content: strings.Join(lines, "\n"), SecondaryText: "Select one with /project <name>."}).PlainFallback, nil
}

func telegramReceiverHealthText(state core.TelegramReceiverState) string {
	text := fmt.Sprintf("Receiver: %s", state.Health)
	if state.RetryAttempt > 0 {
		text += fmt.Sprintf("\nPolling retry attempt: %d of 5.", state.RetryAttempt)
	}

	if state.LastErrorCode != "" {
		text += "\nLast error: " + state.LastErrorCode + "\n\nA local operator can inspect the receiver diagnostics."
	}

	return text
}
