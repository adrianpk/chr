package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramAuthoringDescriptor struct {
	Name, Arguments, Help string
	Action                core.AuthoringContinuationAction
	Target                core.AuthoringArtifactKind
}

type telegramAuthoringCommand struct {
	Input        core.TelegramConversationInput
	Action       core.AuthoringContinuationAction
	Target       core.AuthoringArtifactKind
	Instructions string
}

func telegramAuthoringRegistry() []telegramAuthoringDescriptor {
	return []telegramAuthoringDescriptor{
		{"/quiz", "[conversation]", "Ask one real clarification question", core.AuthoringContinuationAsk, core.AuthoringArtifactKindQuiz},
		{"/spec", "[conversation]", "Generate a draft specification or ask missing information", core.AuthoringContinuationGenerate, core.AuthoringArtifactKindSpec},
		{"/plan", "[conversation]", "Generate a plan from canonical context", core.AuthoringContinuationGenerate, core.AuthoringArtifactKindPlan},
		{"/refine", "<dump|spec|plan> <conversation> <instructions>", "Refine an observed artifact with explicit replacement authority", core.AuthoringContinuationRefine, ""},
	}
}

func parseTelegramAuthoring(update core.TelegramReceivedUpdate) (telegramAuthoringCommand, bool, error) {
	command := telegramAuthoringCommand{Input: core.TelegramConversationInput{Update: update}}
	if update.Kind != "message" {
		return command, false, nil
	}

	name, body := telegramCommandWord(strings.TrimSpace(update.Text))
	for _, descriptor := range telegramAuthoringRegistry() {
		if name != descriptor.Name {
			continue
		}

		command.Action, command.Target = descriptor.Action, descriptor.Target
		if descriptor.Action == core.AuthoringContinuationRefine {
			kind, rest := telegramCommandWord(body)
			command.Target = core.AuthoringArtifactKind(kind)
			command.Input.Conversation, command.Instructions = telegramCommandWord(rest)
		} else {
			command.Input.Conversation = body
		}

		return command, true, validateTelegramAuthoringCommand(command)
	}

	return command, false, nil
}

func validateTelegramAuthoringCommand(command telegramAuthoringCommand) error {
	if len(strings.Fields(command.Input.Conversation)) > 1 || len(command.Input.Conversation) > 128 {
		return ErrRecordUnsupported
	}

	if command.Action == core.AuthoringContinuationRefine {
		if command.Input.Conversation == "" || strings.TrimSpace(command.Instructions) == "" || len(command.Instructions) > 16384 {
			return ErrRecordUnsupported
		}

		if command.Target != core.AuthoringArtifactKindDump && command.Target != core.AuthoringArtifactKindSpec && command.Target != core.AuthoringArtifactKindPlan {
			return ErrRecordUnsupported
		}
	}

	return nil
}

func telegramIntentAuthoring(update core.TelegramReceivedUpdate, intent ConversationIntent) (telegramAuthoringCommand, bool, error) {
	a := intent.Arguments
	command := telegramAuthoringCommand{Input: core.TelegramConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}}

	for _, descriptor := range telegramAuthoringRegistry() {
		if intent.OperationID != strings.TrimPrefix(descriptor.Name, "/") {
			continue
		}

		command.Action, command.Target = descriptor.Action, descriptor.Target
		if a.Namespace != "" || a.RecordJSON != "" {
			return command, true, ErrRecordUnsupported
		}

		if command.Action == core.AuthoringContinuationRefine {
			if a.Text != update.Text {
				return command, true, ErrRecordUnsupported
			}

			command.Target, command.Instructions = core.AuthoringArtifactKind(a.Action), a.Text
		} else if a.Action != "" || a.Text != "" {
			return command, true, ErrRecordUnsupported
		}

		return command, true, validateTelegramAuthoringCommand(command)
	}

	return command, false, nil
}

func telegramAuthoringUsage(update core.TelegramReceivedUpdate) string {
	name, _ := telegramCommandWord(strings.TrimSpace(update.Text))
	for _, descriptor := range telegramAuthoringRegistry() {
		if descriptor.Name == name {
			return "Usage: " + telegramCommandSyntax(name, descriptor.Arguments)
		}
	}

	return "Send /help to choose an authoring command."
}
