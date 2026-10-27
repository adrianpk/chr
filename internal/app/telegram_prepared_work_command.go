package app

import (
	"strings"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramPreparedWorkRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "prepare-work", Description: "Prepare one native backlog ticket from canonical dump/spec context with the conversation's retained intent", Arguments: "Optional project/conversation; no text, action, namespace or record_json. This cannot supply development authority."}}
}

func telegramPreparedWorkHelp() string {
	return telegramPreparedWorkUsage() + " — " + telegramPreparedWorkRegistry()[0].Description
}

func telegramPreparedWorkUsage() string {
	return "/" + telegramPreparedWorkRegistry()[0].ID + " [conversation]"
}

func parseTelegramPreparedWork(update telegramcore.ReceivedUpdate) (telegramcore.ConversationInput, bool, error) {
	input := telegramcore.ConversationInput{Update: update}

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

func telegramIntentPreparedWork(update telegramcore.ReceivedUpdate, intent ConversationIntent) (telegramcore.ConversationInput, bool, error) {
	a := intent.Arguments

	input := telegramcore.ConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}
	if intent.OperationID != telegramPreparedWorkRegistry()[0].ID {
		return input, false, nil
	}

	if a.Text != "" || a.Action != "" || a.Namespace != "" || a.RecordJSON != "" {
		return input, true, ErrRecordUnsupported
	}

	return input, true, nil
}
