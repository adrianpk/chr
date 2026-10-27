package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramExplanationCommand struct {
	Input        telegramcore.TelegramConversationInput
	Question     string
	DocumentKind core.AuthoringArtifactKind
}

func telegramExplanationRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "explain", Description: "Explain canonical context, or show an existing original document without preparing or starting work", Arguments: "To show, read, open or download an existing dump/spec/plan, use action=dump|spec|plan and an exact observed conversation; for questions or explanations leave action empty. Optional project; text must equal the original message; no namespace or record_json. Clarify an ambiguous document or conversation."}}
}

func telegramExplanationHelp() string {
	return telegramExplanationUsage() + " — " + telegramExplanationRegistry()[0].Description
}

func telegramExplanationUsage() string {
	return "/" + telegramExplanationRegistry()[0].ID + " [--project <id>|--conversation <id>] <question>"
}

func parseTelegramExplanation(update telegramcore.TelegramReceivedUpdate) (telegramExplanationCommand, bool, error) {
	command := telegramExplanationCommand{Input: telegramcore.TelegramConversationInput{Update: update}}

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

func telegramIntentExplanation(update telegramcore.TelegramReceivedUpdate, intent ConversationIntent) (telegramExplanationCommand, bool, error) {
	a := intent.Arguments

	command := telegramExplanationCommand{Input: telegramcore.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}, Question: update.Text, DocumentKind: core.AuthoringArtifactKind(a.Action)}
	if intent.OperationID != telegramExplanationRegistry()[0].ID {
		return command, false, nil
	}

	if a.Text != update.Text || a.Namespace != "" || a.RecordJSON != "" || len(a.Text) > 16384 || (a.Action != "" && !validAuthoringDocumentKind(command.DocumentKind)) {
		return command, true, ErrRecordUnsupported
	}

	return command, true, nil
}
