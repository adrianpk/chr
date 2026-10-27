package app

import (
	"context"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramConversationCommand struct {
	Name      string
	Arguments string
	Help      string
	Action    telegramcore.ConversationAction
	Intent    string
}

func telegramConversationRegistry() []telegramConversationCommand {
	return []telegramConversationCommand{
		{"/project", "<project>", "Select an exact allowed project ID or unique name", telegramcore.ConversationSelect, ""},
		{"/prepare", "[--project <id>] <topic>", "Start a preparation-only capture", telegramcore.ConversationStart, "prepare"},
		{"/develop", "[--project <id>] <topic>", "Start a development-intent capture", telegramcore.ConversationStart, "develop"},
		{"/sessions", "", "List your retained sessions", telegramcore.ConversationList, ""},
		{"/focus", "<conversation>", "Focus a collecting session", telegramcore.ConversationFocus, ""},
		{"/reply", "<conversation> <text>", "Append text to an exact session", telegramcore.ConversationCapture, ""},
		{"/dump", "[conversation]", "Show bounded retained capture", telegramcore.ConversationShow, ""},
		{"/done", "[conversation]", "Finish capture and prepare a canonical dump", telegramcore.ConversationDone, ""},
		{"/cancel", "[conversation]", "Discard an unstarted session's capture", telegramcore.ConversationCancel, ""},
	}
}

func telegramConversationInput(update telegramcore.ReceivedUpdate) (telegramcore.ConversationInput, bool, string) {
	input := telegramcore.ConversationInput{Update: update}
	if update.Kind != "message" {
		return input, false, ""
	}

	text := strings.TrimSpace(update.Text)
	if !strings.HasPrefix(text, "/") {
		input.Action, input.Text = telegramcore.ConversationCapture, update.Text

		return input, true, ""
	}

	name, body := telegramCommandWord(text)
	for _, command := range telegramConversationRegistry() {
		if name != command.Name {
			continue
		}

		input.Action, input.Intent = command.Action, command.Intent
		if !parseTelegramConversationArguments(&input, body) {
			return input, true, telegramConversationUsage(command)
		}

		return input, true, ""
	}

	return input, false, ""
}

func telegramCommandWord(text string) (string, string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return "", ""
	}

	return words[0], strings.TrimSpace(strings.TrimPrefix(text, words[0]))
}

func parseTelegramConversationArguments(input *telegramcore.ConversationInput, body string) bool {
	switch input.Action {
	case telegramcore.ConversationSelect:
		input.Project = body

		return body != ""
	case telegramcore.ConversationStart:
		return parseTelegramConversationStart(input, body)
	case telegramcore.ConversationList:
		return body == ""
	case telegramcore.ConversationCapture:
		input.Conversation, input.Text = telegramCommandWord(body)

		return input.Conversation != "" && input.Text != ""
	default:
		input.Conversation = body

		return len(strings.Fields(body)) <= 1 && (body != "" || input.Action != telegramcore.ConversationFocus)
	}
}

func parseTelegramConversationStart(input *telegramcore.ConversationInput, body string) bool {
	first, rest := telegramCommandWord(body)
	if first == "--project" {
		input.Project, body = telegramCommandWord(rest)
		if input.Project == "" || strings.HasPrefix(input.Project, "--") {
			return false
		}
	} else if strings.HasPrefix(first, "--") {
		return false
	}

	input.Text = body

	return input.Text != ""
}

func (r telegramReceiver) respondConversation(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	input, recognized, diagnostic := telegramConversationInput(update)
	if !recognized {
		return false, nil
	}

	if diagnostic != "" {
		return true, r.store.QueueReadResponse(ctx, update, diagnostic, config.Digest(), time.Now())
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, r.store.QueueReadResponse(ctx, update, "The project registry is unavailable. Use local diagnostics before continuing.", config.Digest(), time.Now())
	}

	input.Config, input.Projects = config, projects
	if handled, err := r.dispatchAuthoringConversation(ctx, user, input); handled || err != nil {
		return true, err
	}

	return true, r.store.ApplyConversation(ctx, input, time.Now())
}

func telegramConversationUsage(command telegramConversationCommand) string {
	usage := "Usage: " + telegramCommandSyntax(command.Name, command.Arguments)
	if command.Action == telegramcore.ConversationStart {
		usage += "\n\nExample: " + command.Name + " update the README"
	}

	return usage
}
