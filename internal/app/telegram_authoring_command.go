package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramAuthoringDescriptor struct {
	Name, Arguments, Help string
	Action                core.AuthoringContinuationAction
	Target                core.AuthoringArtifactKind
}

type telegramAuthoringCommand struct {
	Input        telegramcore.ConversationInput
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

func parseTelegramAuthoring(update telegramcore.ReceivedUpdate) (telegramAuthoringParsing, error) {
	command := telegramAuthoringCommand{Input: telegramcore.ConversationInput{Update: update}}
	if update.Kind != "message" {
		return telegramAuthoringParsing{Command: command, Recognized: false}, nil
	}

	word := telegramCommandWord(strings.TrimSpace(update.Text))

	name, body := word.Command, word.Body
	for _, descriptor := range telegramAuthoringRegistry() {
		if name != descriptor.Name {
			continue
		}

		command.Action, command.Target = descriptor.Action, descriptor.Target
		if descriptor.Action == core.AuthoringContinuationRefine {
			word2 := telegramCommandWord(body)
			kind, rest := word2.Command, word2.Body
			command.Target = core.AuthoringArtifactKind(kind)
			word3 := telegramCommandWord(rest)
			command.Input.Conversation, command.Instructions = word3.Command, word3.Body
		} else {
			command.Input.Conversation = body
		}

		return telegramAuthoringParsing{Command: command, Recognized: true}, validateTelegramAuthoringCommand(command)
	}

	return telegramAuthoringParsing{Command: command, Recognized: false}, nil
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

func telegramIntentAuthoring(update telegramcore.ReceivedUpdate, intent ConversationIntent) (telegramAuthoringParsing, error) {
	a := intent.Arguments
	command := telegramAuthoringCommand{Input: telegramcore.ConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation}}

	for _, descriptor := range telegramAuthoringRegistry() {
		if intent.OperationID != strings.TrimPrefix(descriptor.Name, "/") {
			continue
		}

		command.Action, command.Target = descriptor.Action, descriptor.Target
		if a.Namespace != "" || a.RecordJSON != "" {
			return telegramAuthoringParsing{Command: command, Recognized: true}, ErrRecordUnsupported
		}

		if command.Action == core.AuthoringContinuationRefine {
			if a.Text != update.Text {
				return telegramAuthoringParsing{Command: command, Recognized: true}, ErrRecordUnsupported
			}

			command.Target, command.Instructions = core.AuthoringArtifactKind(a.Action), a.Text
		} else if a.Action != "" || a.Text != "" {
			return telegramAuthoringParsing{Command: command, Recognized: true}, ErrRecordUnsupported
		}

		return telegramAuthoringParsing{Command: command, Recognized: true}, validateTelegramAuthoringCommand(command)
	}

	return telegramAuthoringParsing{Command: command, Recognized: false}, nil
}

func telegramAuthoringUsage(update telegramcore.ReceivedUpdate) string {
	word4 := telegramCommandWord(strings.TrimSpace(update.Text))

	name := word4.Command
	for _, descriptor := range telegramAuthoringRegistry() {
		if descriptor.Name == name {
			return "Usage: " + telegramCommandSyntax(name, descriptor.Arguments)
		}
	}

	return "Send /help to choose an authoring command."
}
