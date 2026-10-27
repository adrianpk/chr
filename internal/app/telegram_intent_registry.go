package app

import (
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// The model and deterministic parser obtain operation identities from the same
// registries; neither a model result nor a command can introduce another owner.
func telegramIntentRegistry() []ConversationIntentOperation {
	result := telegramExplanationRegistry()

	result = append(result, telegramPreparedWorkRegistry()...)

	result = append(result, telegramWorkRegistry()...)
	result = append(result, telegramControlRegistry()...)

	for _, operation := range telegramReadRegistry() {
		result = append(result, ConversationIntentOperation{ID: strings.TrimPrefix(operation.Command, "/"), Description: operation.Help, Arguments: "No arguments"})
	}

	for _, command := range telegramConversationRegistry() {
		result = append(result, ConversationIntentOperation{ID: strings.TrimPrefix(command.Name, "/"), Description: command.Help, Arguments: command.Arguments})
	}

	for _, descriptor := range telegramAuthoringRegistry() {
		arguments := "Optional project/conversation; no text, action, namespace or record_json"
		if descriptor.Action == core.AuthoringContinuationRefine {
			arguments = "conversation=exact session; action=dump|spec|plan; text=the original message"
		}

		result = append(result, ConversationIntentOperation{ID: strings.TrimPrefix(descriptor.Name, "/"), Description: descriptor.Help, Arguments: arguments})
	}

	result = append(result, ConversationIntentOperation{ID: "capture", Description: "Append the original message to the targeted collecting conversation", Arguments: "Optional conversation; text must equal the original message"})
	for _, command := range telegramRecordRegistry() {
		result = append(result, ConversationIntentOperation{ID: strings.TrimPrefix(command.Name, "/"), Description: "Read or explicitly create/update a guarded " + string(command.Kind) + " record", Arguments: "Optional project/conversation/namespace; action=list|show|create|update; record_json=strict JSON with actual observed revision/source_hash guards"})
	}

	return result
}

func telegramIntentConversation(update telegramcore.TelegramReceivedUpdate, intent ConversationIntent) (telegramcore.TelegramConversationInput, bool, error) {
	a := intent.Arguments

	input := telegramcore.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation, Namespace: a.Namespace, Text: a.Text}
	if intent.OperationID == "capture" {
		input.Action = telegramcore.TelegramConversationCapture
		if !validTelegramIntentCapture(a, update.Text) {
			return input, true, ErrRecordUnsupported
		}

		return input, true, nil
	}

	for _, command := range telegramConversationRegistry() {
		if intent.OperationID != strings.TrimPrefix(command.Name, "/") {
			continue
		}

		input.Action, input.Intent = command.Action, command.Intent
		if a.Action != "" || a.RecordJSON != "" || a.Namespace != "" {
			return input, true, ErrRecordUnsupported
		}

		return input, true, validateTelegramIntentConversation(input)
	}

	return input, false, nil
}

func validateTelegramIntentConversation(input telegramcore.TelegramConversationInput) error {
	var valid bool

	switch input.Action {
	case telegramcore.TelegramConversationStart:
		valid = validTelegramIntentStart(input)
	case telegramcore.TelegramConversationSelect:
		valid = input.Project != "" && input.Conversation == "" && input.Text == ""
	case telegramcore.TelegramConversationCapture:
		valid = input.Conversation != "" && input.Text == input.Update.Text && input.Project == ""
	default:
		valid = validTelegramIntentControl(input)
	}

	if !valid {
		return ErrRecordUnsupported
	}

	return nil
}

func validTelegramIntentCapture(a ConversationIntentArguments, text string) bool {
	return a.Text == text && a.Project == "" && a.Namespace == "" && a.Action == "" && a.RecordJSON == ""
}

func validTelegramIntentStart(input telegramcore.TelegramConversationInput) bool {
	return input.Text != "" && len(input.Text) <= 512 && input.Conversation == ""
}

func validTelegramIntentControl(input telegramcore.TelegramConversationInput) bool {
	if input.Text != "" || input.Project != "" {
		return false
	}

	switch input.Action {
	case telegramcore.TelegramConversationList:
		return input.Conversation == ""
	case telegramcore.TelegramConversationFocus:
		return input.Conversation != ""
	default:
		return true
	}
}

func telegramIntentRecord(update telegramcore.TelegramReceivedUpdate, intent ConversationIntent) (telegramRecordCommand, bool, error) {
	a := intent.Arguments
	command := telegramRecordCommand{Input: telegramcore.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation, Namespace: a.Namespace}, Action: a.Action}

	for _, descriptor := range telegramRecordRegistry() {
		if intent.OperationID != strings.TrimPrefix(descriptor.Name, "/") {
			continue
		}

		command.Kind = descriptor.Kind
		if a.Text != "" {
			return command, true, errors.New("unexpected record text")
		}

		return command, true, decodeTelegramRecord(&command, a.RecordJSON)
	}

	return command, false, nil
}
