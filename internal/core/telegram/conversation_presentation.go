package telegram

import (
	"fmt"
	"strings"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramConversationLabel(text string) string {
	return tgmessage.BoundBlock(strings.Join(strings.Fields(text), " "), 160)
}

func (op telegramConversationTransaction) projectLabel(id string) string {
	for _, project := range op.allowedProjects() {
		if project.ID == id {
			return telegramConversationLabel(project.Name)
		}
	}

	return "an unavailable project"
}

func telegramConversationState(phase string) string {
	states := map[string]string{
		"collecting": "collecting details", "clarifying": "waiting for your answer",
		"preparing": "capture saved", "prepared": "prepared", "awaiting_confirmation": "waiting for confirmation",
		"completed": "completed", "canceled": "canceled", "failed": "failed",
	}
	if state, ok := states[phase]; ok {
		return state
	}

	return phase
}

func (op telegramConversationTransaction) presentation(text string, c *TelegramConversation) tgmessage.Blocks {
	families := map[TelegramConversationAction]string{
		TelegramConversationSelect: "04", TelegramConversationStart: "07", TelegramConversationList: "09", TelegramConversationFocus: "10", TelegramConversationCapture: "12", TelegramConversationShow: "13", TelegramConversationDone: "15", TelegramConversationCancel: "16",
	}

	message := tgmessage.Blocks{Family: families[op.input.Action], InitialText: text}
	if c == nil || op.input.Action != TelegramConversationShow {
		return message
	}

	message.InitialText = telegramConversationLabel(c.Topic) + " in " + op.projectLabel(c.ProjectID) + "\nState: " + telegramConversationState(c.Phase)

	message.ContentKind, message.Content = "markdown", c.Capture
	if message.Content == "" {
		message.ContentKind, message.Content = "text", "No additional details collected."
	}

	message.SecondaryText = "Saved details are unchanged.\nReference: " + c.ID

	return message
}

func (op telegramConversationTransaction) conversationListRow(c TelegramConversation, focus string) string {
	project := op.projectLabel(c.ProjectID)

	label := telegramConversationLabel(c.Topic) + " in " + project
	if project == "an unavailable project" {
		label = "Unavailable conversation; project access changed"
	}

	state := telegramConversationState(c.Phase)
	if c.ID == focus {
		state += "; focused"
	}

	command := "/dump " + c.ID
	if telegramCaptureEligible(c) {
		command = "/focus " + c.ID
	}

	return "\n" + label + " — " + state + "\n" + command + "\n"
}

func telegramCaptureSavedText(c TelegramConversation, parts int) string {
	response := "Details saved. Send more details, or /done " + c.ID + " to finish."
	if parts >= 28 || len(c.Capture) >= 56*1024 {
		response += fmt.Sprintf("\n\nNear the capture limit: %d of 32 parts; %d of 65536 bytes.", parts, len(c.Capture))
	}

	return response
}
