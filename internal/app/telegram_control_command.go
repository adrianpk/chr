package app

import (
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramControlCommand struct {
	Input      core.TelegramConversationInput
	Action     string
	Invocation string
}

func telegramControlRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "control", Description: "Inspect or control an exact retained native invocation", Arguments: "Optional project/conversation/namespace; action=status|pause|cancel|resume|continue|reconcile; text=exact invocation (empty only for a bound conversation); no record_json"}}
}

func telegramControlHelp() string {
	return "/status|/pause|/cancel|/resume|/continue|/reconcile [--project <id>] [--conversation <id>] [--namespace <name>] [execution] — Inspect or confirm an exact retained invocation and revision"
}

func parseTelegramControl(update core.TelegramReceivedUpdate) (telegramControlCommand, bool, error) {
	command := telegramControlCommand{Input: core.TelegramConversationInput{Update: update}}
	name, body := telegramCommandWord(strings.TrimSpace(update.Text))

	command.Action = strings.TrimPrefix(name, "/")
	if update.Kind != "message" || !strings.HasPrefix(name, "/") || !validTelegramControlAction(command.Action) {
		return command, false, nil
	}

	if body == "" && update.ReplyToMessageID == 0 && (command.Action == "cancel" || command.Action == "status") {
		return command, false, nil
	}

	body, err := parseTelegramRecordScope(&command.Input, body)

	command.Invocation = body
	if err != nil {
		return command, true, err
	}

	return command, true, validateTelegramControlCommand(command)
}

func validTelegramControlAction(action string) bool {
	return slices.Contains([]string{"status", "pause", "cancel", "resume", "continue", "reconcile"}, action)
}

func validateTelegramControlCommand(command telegramControlCommand) error {
	if !validTelegramControlAction(command.Action) || command.Invocation != "" && (!validWorkflowControlIdentity(command.Invocation) || len(strings.Fields(command.Invocation)) != 1) {
		return ErrRecordUnsupported
	}

	return nil
}

func telegramIntentControl(update core.TelegramReceivedUpdate, intent ConversationIntent) (telegramControlCommand, bool, error) {
	a := intent.Arguments

	command := telegramControlCommand{Input: core.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation, Namespace: a.Namespace}, Action: a.Action, Invocation: a.Text}
	if intent.OperationID != "control" {
		return command, false, nil
	}

	if a.RecordJSON != "" {
		return command, true, ErrRecordUnsupported
	}

	return command, true, validateTelegramControlCommand(command)
}
