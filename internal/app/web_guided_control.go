package app

import (
	"crypto/rand"
	"strconv"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func guidedWebCommand(request webui.WriteActionRequest) (core.GuidedUseCaseCommand, error) {
	command := core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: request.ID, Namespace: request.Fields["namespace"]}, OperationID: request.Fields["request_id"], Action: request.Fields["step"], Content: request.Fields["content"], QuestionID: request.Fields["question_id"], InputDigest: request.Fields["input_digest"]}

	var err error

	command.ExpectedRevision, err = strconv.ParseInt(request.Revision, 10, 64)
	if err == nil && request.Fields["native_revision"] != "" {
		command.NativeRevision, err = strconv.ParseUint(request.Fields["native_revision"], 10, 64)
	}

	return command, err
}

func guidedWebPending(session core.AuthoringSession) bool {
	return session.Guided != nil && (session.Guided.PendingOperation != nil || session.Guided.Control != nil && session.Guided.Control.PendingOperation != nil)
}

func guidedWebControls(view *webui.GuidedUseCaseView, session core.AuthoringSession, work *WorkflowWorkStatus) {
	view.Suspended = !core.GuidedDispatchAllowed(session)
	if command, ready := guidedControlCommand(session); ready {
		view.Controls = []webui.GuidedUseCaseControl{guidedWebControl(command, "Retry "+command.Action)}
		view.ControlStatus = "The control result is uncertain. Retry the retained request to reconcile it."

		return
	}

	if guidedWebControlTerminal(session.Guided.Status, work) {
		return
	}

	actions := guidedWebControlActions(view, session.Guided)

	for _, action := range actions {
		command := core.GuidedUseCaseCommand{Action: action, OperationID: action + "-" + rand.Text(), ExpectedRevision: session.Revision}
		if work != nil {
			command.NativeRevision = work.Revision
		}

		view.Controls = append(view.Controls, guidedWebControl(command, guidedControlLabel(action)))
	}
}

func guidedWebControl(command core.GuidedUseCaseCommand, label string) webui.GuidedUseCaseControl {
	return webui.GuidedUseCaseControl{Action: command.Action, Label: label, RequestID: command.OperationID, Revision: strconv.FormatInt(command.ExpectedRevision, 10), NativeRevision: strconv.FormatUint(command.NativeRevision, 10), Content: command.Content, QuestionID: command.QuestionID, InputDigest: command.InputDigest}
}

func guidedControlLabel(action string) string {
	switch action {
	case "pause":
		return "Pause use case"
	case "resume":
		return "Resume use case"
	default:
		return "Cancel use case"
	}
}

func guidedWebControlTerminal(status string, work *WorkflowWorkStatus) bool {
	return status == "cancelled" || status == "completed" || status == "failed" || work != nil && work.Status == "completed"
}

func guidedWebControlActions(view *webui.GuidedUseCaseView, state *core.GuidedUseCaseState) []string {
	if state.Control == nil || state.Control.StopIntent == "" {
		return []string{"pause", "cancel"}
	}

	view.ControlStatus = "Stop requested; waiting for native work to become quiescent."
	if state.Status == "paused" {
		view.ControlStatus = "Paused. Retained inputs, work membership and attempt limits are preserved."

		return []string{"resume", "cancel"}
	}

	return []string{"cancel"}
}
