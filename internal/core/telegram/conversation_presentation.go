package telegram

import (
	"fmt"
	"strings"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func conversationLabel(text string) string {
	return tgmessage.BoundBlock(strings.Join(strings.Fields(text), " "), 160)
}

func (op conversationTransaction) projectLabel(id string) string {
	for _, project := range op.allowedProjects() {
		if project.ID == id {
			return conversationLabel(project.Name)
		}
	}

	return "an unavailable project"
}

func conversationState(phase ConversationPhase) string {
	states := map[ConversationPhase]string{
		ConversationCollecting: "collecting details", ConversationClarifying: "waiting for your answer",
		ConversationPreparing: "capture saved", ConversationPrepared: "prepared", ConversationAwaitingConfirmation: "waiting for confirmation",
		ConversationCompleted: "completed", ConversationCanceled: "canceled", ConversationFailed: "failed",
	}
	if state, ok := states[phase]; ok {
		return state
	}

	return string(phase)
}

func (op conversationTransaction) presentation(text string, c *Conversation) tgmessage.Blocks {
	families := map[ConversationAction]string{
		ConversationSelect: "04", ConversationStart: "07", ConversationList: "09", ConversationFocus: "10", ConversationCapture: "12", ConversationShow: "13", ConversationDone: "15", ConversationCancel: "16",
	}

	message := tgmessage.Blocks{Family: families[op.input.Action], InitialText: text}
	if c == nil || op.input.Action != ConversationShow {
		return message
	}

	message.InitialText = conversationLabel(c.Topic) + " in " + op.projectLabel(c.ProjectID) + "\nState: " + conversationState(c.Phase)

	message.ContentKind, message.Content = "markdown", c.Capture
	if message.Content == "" {
		message.ContentKind, message.Content = "text", "No additional details collected."
	}

	message.SecondaryText = "Saved details are unchanged.\nReference: " + c.ID

	return message
}

func (op conversationTransaction) conversationListRow(c Conversation, focus string) string {
	project := op.projectLabel(c.ProjectID)

	label := conversationLabel(c.Topic) + " in " + project
	if project == "an unavailable project" {
		label = "Unavailable conversation; project access changed"
	}

	state := conversationState(c.Phase)
	if c.ID == focus {
		state += "; focused"
	}

	command := "/dump " + c.ID
	if captureEligible(c) {
		command = "/focus " + c.ID
	}

	return "\n" + label + " — " + state + "\n" + command + "\n"
}

func captureSavedText(c Conversation, parts int) string {
	response := "Details saved. Send more details, or /done " + c.ID + " to finish."
	if parts >= 28 || len(c.Capture) >= 56*1024 {
		response += fmt.Sprintf("\n\nNear the capture limit: %d of 32 parts; %d of 65536 bytes.", parts, len(c.Capture))
	}

	return response
}
