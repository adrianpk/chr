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

	return telegramHelpReply("Unknown or unavailable command.\n"), nil
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

		return telegramReadReply{Text: fmt.Sprintf("Receiver: %s\nBot: %d\nCursor epoch: %d\nPoll attempts in current failure cycle: %d\nLast error: %s", state.Health, state.BotID, state.Epoch, state.RetryAttempt, state.LastErrorCode)}, nil
	case telegramReadHelp:
		return telegramHelpReply(""), nil
	}

	return telegramHelpReply("Unknown or unavailable command.\n"), nil
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
		return "No enabled registered projects are available for your identity.", nil
	}

	return core.BoundTelegramReceiverText("Registered projects. Select one with /project <name>.\n" + strings.Join(lines, "\n")), nil
}
