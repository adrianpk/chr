package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramExplanationCommand struct {
	Input        telegramcore.ConversationInput
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

func parseTelegramExplanation(update telegramcore.ReceivedUpdate) (telegramExplanationParsing, error) {
	command := telegramExplanationCommand{Input: telegramcore.ConversationInput{Update: update}}

	word := telegramCommandWord(strings.TrimSpace(update.Text))

	name, body := word.Command, word.Body
	if update.Kind != "message" || name != "/"+telegramExplanationRegistry()[0].ID {
		return telegramExplanationParsing{Command: command, Recognized: false}, nil
	}

	word2 := telegramCommandWord(body)

	flag, rest := word2.Command, word2.Body
	switch flag {
	case "--project":
		word3 := telegramCommandWord(rest)
		command.Input.Project, body = word3.Command, word3.Body
	case "--conversation":
		word4 := telegramCommandWord(rest)
		command.Input.Conversation, body = word4.Command, word4.Body
	}

	command.Question = body
	if strings.TrimSpace(body) == "" || len(body) > 16384 || (strings.HasPrefix(flag, "--") && flag != "--project" && flag != "--conversation") {
		return telegramExplanationParsing{Command: command, Recognized: true}, ErrRecordUnsupported
	}

	return telegramExplanationParsing{Command: command, Recognized: true}, nil
}

func telegramIntentExplanation(update telegramcore.ReceivedUpdate, intent ConversationIntent) (telegramExplanationParsing, error) {
	a := intent.Arguments

	command := telegramExplanationCommand{Input: telegramcore.ConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}, Question: update.Text, DocumentKind: core.AuthoringArtifactKind(a.Action)}
	if intent.OperationID != telegramExplanationRegistry()[0].ID {
		return telegramExplanationParsing{Command: command, Recognized: false}, nil
	}

	if a.Text != update.Text || a.Namespace != "" || a.RecordJSON != "" || len(a.Text) > 16384 || (a.Action != "" && !validAuthoringDocumentKind(command.DocumentKind)) {
		return telegramExplanationParsing{Command: command, Recognized: true}, ErrRecordUnsupported
	}

	return telegramExplanationParsing{Command: command, Recognized: true}, nil
}
