package app

import (
	"strings"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramRecordCommand struct {
	Input    telegramcore.TelegramConversationInput
	Kind     RecordKind
	Action   string
	Query    RecordQuery
	Mutation RecordMutationRequest
}

type telegramRecordDescriptor struct {
	Name string
	Kind RecordKind
}

func telegramRecordRegistry() []telegramRecordDescriptor {
	return []telegramRecordDescriptor{{"/tickets", RecordTicket}, {"/ticket", RecordTicket}, {"/slices", RecordSlice}, {"/slice-sets", RecordSliceSet}, {"/use-cases", RecordUseCase}}
}

func telegramRecordUsage(command telegramRecordDescriptor) string {
	return command.Name + " [--project <id>] [--conversation <id>] [--namespace <name>] list|show|create|update <json>"
}

func parseTelegramRecord(update telegramcore.TelegramReceivedUpdate) (telegramRecordCommand, bool, error) {
	command := telegramRecordCommand{Input: telegramcore.TelegramConversationInput{Update: update}}
	if update.Kind != "message" {
		return command, false, nil
	}

	name, body := telegramCommandWord(strings.TrimSpace(update.Text))
	for _, descriptor := range telegramRecordRegistry() {
		if name == descriptor.Name {
			command.Kind = descriptor.Kind

			break
		}
	}

	if command.Kind == "" {
		return command, false, nil
	}

	body, err := parseTelegramRecordScope(&command.Input, body)
	if err != nil {
		return command, true, err
	}

	command.Action, body = telegramCommandWord(body)
	if command.Action == "" {
		command.Action = "list"
	}

	err = decodeTelegramRecord(&command, body)

	return command, true, err
}

func parseTelegramRecordScope(input *telegramcore.TelegramConversationInput, body string) (string, error) {
	seen := map[string]bool{}

	for strings.HasPrefix(body, "--") {
		name, rest := telegramCommandWord(body)

		value, next := telegramCommandWord(rest)
		if value == "" || strings.HasPrefix(value, "--") || seen[name] {
			return "", ErrRecordUnsupported
		}

		seen[name] = true
		switch name {
		case "--project":
			input.Project = value
		case "--conversation":
			input.Conversation = value
		case "--namespace":
			input.Namespace = value
		default:
			return "", ErrRecordUnsupported
		}

		body = next
	}

	return body, nil
}

func decodeTelegramRecord(command *telegramRecordCommand, body string) error {
	switch command.Action {
	case "list", "show":
		return decodeTelegramRecordQuery(command, body)
	case "create", "update":
		if err := decodeOperationJSON(body, &command.Mutation); err != nil {
			return err
		}

		if command.Mutation.Kind != "" || command.Mutation.Action != "" {
			return ErrRecordUnsupported
		}

		command.Mutation.Kind, command.Mutation.Action = command.Kind, RecordMutationAction(command.Action)
	default:
		return ErrRecordUnsupported
	}

	return nil
}

func decodeTelegramRecordQuery(command *telegramRecordCommand, body string) error {
	command.Query.Kind = command.Kind
	if body != "" {
		if err := decodeOperationJSON(body, &command.Query); err != nil {
			return err
		}
	}

	if command.Query.Kind != command.Kind || (command.Action == "show" && command.Query.Identity == "") || (command.Action == "list" && command.Query.Identity != "") {
		return ErrRecordUnsupported
	}

	return nil
}
