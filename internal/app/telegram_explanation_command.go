package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramExplanationCommand struct {
	Input    core.TelegramConversationInput
	Question string
}

func telegramExplanationRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "explain", Description: "Explain canonical project or session context without preparing or starting work", Arguments: "Optional project/conversation; text must equal the original message; no namespace, action or record_json"}}
}

func telegramExplanationHelp() string {
	return telegramExplanationUsage() + " — " + telegramExplanationRegistry()[0].Description
}

func telegramExplanationUsage() string {
	return "/" + telegramExplanationRegistry()[0].ID + " [--project <id>|--conversation <id>] <question>"
}

func parseTelegramExplanation(update core.TelegramReceivedUpdate) (telegramExplanationCommand, bool, error) {
	command := telegramExplanationCommand{Input: core.TelegramConversationInput{Update: update}}

	name, body := telegramCommandWord(strings.TrimSpace(update.Text))
	if update.Kind != "message" || name != "/"+telegramExplanationRegistry()[0].ID {
		return command, false, nil
	}

	flag, rest := telegramCommandWord(body)
	switch flag {
	case "--project":
		command.Input.Project, body = telegramCommandWord(rest)
	case "--conversation":
		command.Input.Conversation, body = telegramCommandWord(rest)
	}

	command.Question = body
	if strings.TrimSpace(body) == "" || len(body) > 16384 || (strings.HasPrefix(flag, "--") && flag != "--project" && flag != "--conversation") {
		return command, true, ErrRecordUnsupported
	}

	return command, true, nil
}

func telegramIntentExplanation(update core.TelegramReceivedUpdate, intent ConversationIntent) (telegramExplanationCommand, bool, error) {
	a := intent.Arguments

	command := telegramExplanationCommand{Input: core.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}, Question: update.Text}
	if intent.OperationID != telegramExplanationRegistry()[0].ID {
		return command, false, nil
	}

	if a.Text != update.Text || a.Namespace != "" || a.Action != "" || a.RecordJSON != "" || len(a.Text) > 16384 {
		return command, true, ErrRecordUnsupported
	}

	return command, true, nil
}
