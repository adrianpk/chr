package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramWorkCommand struct {
	Input    telegramcore.ConversationInput
	Selector core.WorkflowSelector
}

func telegramWorkRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "work", Description: "Preview and confirm an exact native work extent", Arguments: "Optional project/conversation/namespace; action=ticket|slice|slice_set|use_case|loose; text=exact reference (empty only for loose); no record_json"}}
}

func telegramWorkHelp() string {
	return telegramWorkUsage() + " — Preview the exact native selection and policy before admission"
}

func telegramWorkUsage() string {
	return "/" + telegramWorkRegistry()[0].ID + " [--project <id>] [--conversation <id>] [--namespace <name>] ticket|slice|slice-set|use-case <reference> or loose"
}

func parseTelegramWork(update telegramcore.ReceivedUpdate) (telegramWorkParsing, error) {
	command := telegramWorkCommand{Input: telegramcore.ConversationInput{Update: update}}

	word := telegramCommandWord(strings.TrimSpace(update.Text))

	name, body := word.Command, word.Body
	if update.Kind != "message" || name != "/"+telegramWorkRegistry()[0].ID {
		return telegramWorkParsing{Command: command, Recognized: false}, nil
	}

	body, err := parseTelegramRecordScope(&command.Input, body)
	if err != nil {
		return telegramWorkParsing{Command: command, Recognized: true}, err
	}

	word2 := telegramCommandWord(body)
	command.Selector.Extent, command.Selector.Identity = word2.Command, word2.Body
	command.Selector.Extent = strings.ReplaceAll(command.Selector.Extent, "-", "_")

	return telegramWorkParsing{Command: command, Recognized: true}, validateTelegramWorkSelector(command.Selector)
}

func validateTelegramWorkSelector(selector core.WorkflowSelector) error {
	if selector.Extent == "loose" && selector.Identity == "" {
		return nil
	}

	if selector.Extent != "ticket" && selector.Extent != "slice" && selector.Extent != "slice_set" && selector.Extent != "use_case" {
		return ErrRecordUnsupported
	}

	if selector.Identity == "" || len(selector.Identity) > 128 || len(strings.Fields(selector.Identity)) != 1 {
		return ErrRecordUnsupported
	}

	return nil
}

func telegramIntentWork(update telegramcore.ReceivedUpdate, intent ConversationIntent) (telegramWorkParsing, error) {
	a := intent.Arguments

	command := telegramWorkCommand{Input: telegramcore.ConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation, Namespace: a.Namespace}, Selector: core.WorkflowSelector{Extent: a.Action, Identity: a.Text}}
	if intent.OperationID != telegramWorkRegistry()[0].ID {
		return telegramWorkParsing{Command: command, Recognized: false}, nil
	}

	if a.RecordJSON != "" {
		return telegramWorkParsing{Command: command, Recognized: true}, ErrRecordUnsupported
	}

	return telegramWorkParsing{Command: command, Recognized: true}, validateTelegramWorkSelector(command.Selector)
}
