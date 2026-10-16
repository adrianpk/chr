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

func telegramReadResponse(operations telegramReadService, user core.TelegramAllowedUser, update core.TelegramReceivedUpdate) (string, error) {
	if update.Kind == "callback_query" {
		return "This action is unavailable. Use /help for current commands.", nil
	}

	for _, operation := range telegramReadRegistry() {
		if strings.TrimSpace(update.Text) != operation.Command {
			continue
		}

		switch operation.ID {
		case telegramReadProjects:
			return telegramProjectResponse(operations, user.ProjectIDs)
		case telegramReadStatus:
			state, err := operations.InspectTelegramReceiver(update.Peer.BotID)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("Receiver: %s\nBot: %d\nCursor epoch: %d\nPoll attempts in current failure cycle: %d\nLast error: %s", state.Health, state.BotID, state.Epoch, state.RetryAttempt, state.LastErrorCode), nil
		case telegramReadHelp:
			return telegramReadHelpText(), nil
		}
	}

	return "Unknown or unavailable command.\n" + telegramReadHelpText(), nil
}

func telegramReadHelpText() string {
	var text strings.Builder
	for _, operation := range telegramReadRegistry() {
		fmt.Fprintf(&text, "%s — %s\n", operation.Command, operation.Help)
	}

	return strings.TrimSpace(text.String()) + "\n" + telegramConversationHelp() + "\n" + telegramRecordHelp() + "\n" + telegramActionHelp()
}

func telegramProjectResponse(operations telegramReadService, ids []string) (string, error) {
	projects, err := operations.ListRegisteredProjects(ids)
	if err != nil {
		return "", err
	}

	var lines []string

	for _, project := range projects {
		if project.Enabled {
			lines = append(lines, project.ID+" — "+project.Name)
		}
	}

	if len(lines) == 0 {
		return "No enabled registered projects are available for your identity.", nil
	}

	return core.BoundTelegramReceiverText(strings.Join(lines, "\n")), nil
}
