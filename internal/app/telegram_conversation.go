package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramConversationCommand struct {
	Name      string
	Arguments string
	Help      string
	Action    core.TelegramConversationAction
	Intent    string
}

func telegramConversationRegistry() []telegramConversationCommand {
	return []telegramConversationCommand{
		{"/project", "<project>", "Select an exact allowed project ID or unique name", core.TelegramConversationSelect, ""},
		{"/prepare", "[--project <id>] <topic>", "Start a preparation-only capture", core.TelegramConversationStart, "prepare"},
		{"/develop", "[--project <id>] <topic>", "Start a development-intent capture", core.TelegramConversationStart, "develop"},
		{"/sessions", "", "List your retained sessions", core.TelegramConversationList, ""},
		{"/focus", "<conversation>", "Focus a collecting session", core.TelegramConversationFocus, ""},
		{"/reply", "<conversation> <text>", "Append text to an exact session", core.TelegramConversationCapture, ""},
		{"/dump", "[conversation]", "Show bounded retained capture", core.TelegramConversationShow, ""},
		{"/done", "[conversation]", "Finish capture and retain it for authoring", core.TelegramConversationDone, ""},
		{"/cancel", "[conversation]", "Discard an unstarted session's capture", core.TelegramConversationCancel, ""},
	}
}

func telegramConversationHelp() string {
	var text strings.Builder
	for _, command := range telegramConversationRegistry() {
		fmt.Fprintf(&text, "%s %s — %s\n", command.Name, command.Arguments, command.Help)
	}

	return strings.TrimSpace(text.String())
}

func telegramConversationInput(update core.TelegramReceivedUpdate) (core.TelegramConversationInput, bool, string) {
	input := core.TelegramConversationInput{Update: update}
	if update.Kind != "message" {
		return input, false, ""
	}

	text := strings.TrimSpace(update.Text)
	if !strings.HasPrefix(text, "/") {
		input.Action, input.Text = core.TelegramConversationCapture, update.Text

		return input, true, ""
	}

	name, body := telegramCommandWord(text)
	for _, command := range telegramConversationRegistry() {
		if name != command.Name {
			continue
		}

		input.Action, input.Intent = command.Action, command.Intent
		if !parseTelegramConversationArguments(&input, body) {
			return input, true, "Usage: " + command.Name + " " + command.Arguments
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

func parseTelegramConversationArguments(input *core.TelegramConversationInput, body string) bool {
	switch input.Action {
	case core.TelegramConversationSelect:
		input.Project = body

		return body != ""
	case core.TelegramConversationStart:
		return parseTelegramConversationStart(input, body)
	case core.TelegramConversationList:
		return body == ""
	case core.TelegramConversationCapture:
		input.Conversation, input.Text = telegramCommandWord(body)

		return input.Conversation != "" && input.Text != ""
	default:
		input.Conversation = body

		return len(strings.Fields(body)) <= 1 && (body != "" || input.Action != core.TelegramConversationFocus)
	}
}

func parseTelegramConversationStart(input *core.TelegramConversationInput, body string) bool {
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

	return input.Text != "" && len(input.Text) <= 512
}

func (r telegramReceiver) respondConversation(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
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

	return true, r.store.ApplyTelegramConversation(ctx, input, time.Now())
}
