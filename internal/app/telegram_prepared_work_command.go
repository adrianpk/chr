package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramPreparedWorkRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "prepare-work", Description: "Prepare one native backlog ticket from canonical dump/spec context with the conversation's retained intent", Arguments: "Optional project/conversation; no text, action, namespace or record_json. This cannot supply development authority."}}
}

func telegramPreparedWorkHelp() string {
	return "/" + telegramPreparedWorkRegistry()[0].ID + " [conversation] — " + telegramPreparedWorkRegistry()[0].Description
}

func parseTelegramPreparedWork(update core.TelegramReceivedUpdate) (core.TelegramConversationInput, bool, error) {
	input := core.TelegramConversationInput{Update: update}

	name, body := telegramCommandWord(strings.TrimSpace(update.Text))
	if update.Kind != "message" || name != "/"+telegramPreparedWorkRegistry()[0].ID {
		return input, false, nil
	}

	input.Conversation = body
	if len(strings.Fields(body)) > 1 || len(body) > 128 {
		return input, true, ErrRecordUnsupported
	}

	return input, true, nil
}

func telegramIntentPreparedWork(update core.TelegramReceivedUpdate, intent ConversationIntent) (core.TelegramConversationInput, bool, error) {
	a := intent.Arguments

	input := core.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}
	if intent.OperationID != telegramPreparedWorkRegistry()[0].ID {
		return input, false, nil
	}

	if a.Text != "" || a.Action != "" || a.Namespace != "" || a.RecordJSON != "" {
		return input, true, ErrRecordUnsupported
	}

	return input, true, nil
}
