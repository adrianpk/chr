package app

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramGuidedPrompt(observed GuidedUseCaseInspection) string {
	session := observed.Session
	state := session.Guided

	text := "Guided use case: " + session.Topic + "\nStage: " + state.Stage + " · " + state.Status
	if state.Blocker != "" {
		text += "\n" + state.Blocker
	}

	if !core.GuidedDispatchAllowed(*session) {
		return text + "\n\nSaved inputs and work limits are retained. Send Show controls for the available controls, or Show current step to refresh."
	}

	text += telegramGuidedInputPrompt(observed)
	if len(state.Checkpoints) > 0 {
		text += "\nSend Idea: followed by a later request to save it separately for review after completion."
	}

	return text + "\nSend Show controls for available actions, or Show current step to refresh."
}

func telegramGuidedInputPrompt(observed GuidedUseCaseInspection) string {
	state := observed.Session.Guided
	if state.PendingOperation != nil {
		return "\n\nThe accepted step is retained. Reply Retry to reconcile its original request."
	}

	switch state.Stage {
	case "name":
		return "\n\nWhat is the use-case name? Reply here."
	case "dump":
		return "\n\nAdd everything known about this use case. Reply Finish capture when the context is complete."
	case "quiz":
		if observed.Question != "" {
			return "\n\n" + observed.Question + "\nReply here to answer this current question."
		}

		return "\n\nClarification needs recovery. Reply Retry to continue its retained step."
	case "specification", "preparation":
		return "\n\nThe next step needs its existing prerequisites. Reply Retry after correcting them."
	case "execution":
		return telegramGuidedWorkPrompt(observed.Work)
	default:
		return "\n\nInspect the confirmed result before requesting further work."
	}
}

func telegramGuidedWorkPrompt(work *WorkflowWorkStatus) string {
	if work == nil {
		return "\n\nNative work status is unavailable; delivery is unconfirmed."
	}

	text := fmt.Sprintf("\n\nWork: %s. %d of %d validated; %d confirmed; %d remaining.", strings.ReplaceAll(work.Status, "_", " "), work.ValidatedTasks, work.Total, work.ConfirmedTickets, work.Remaining)
	if work.Blocker != nil {
		text += "\n" + work.Blocker.Reason + ". Next action owner: " + work.Blocker.NextActionOwner + "."
	}

	if work.ReviewState == "waiting_approval" || work.ReviewState == "human_required" {
		text += "\nThe native review requires approval before execution can continue."
	}

	if work.Integration != nil && work.Integration.Blocker != nil {
		text += "\nRequired integration action: " + work.Integration.Blocker.RequiredAction
	}

	return text
}
