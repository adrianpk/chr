package app

import "forge.adrianpk.com/chr/internal/core"

func guidedControlChoices(observed GuidedUseCaseInspection) []string {
	session := *observed.Session

	state := session.Guided
	if guidedControlTerminal(state.Status, observed.Work) {
		return nil
	}

	if _, ready := guidedControlCommand(session); ready {
		return nil
	}

	if state.Control != nil && state.Control.StopIntent != "" {
		if state.Status == "paused" {
			return []string{"resume", "cancel"}
		}

		return []string{"cancel"}
	}

	choices := []string{"pause", "cancel"}
	if state.Stage == "dump" && state.PendingOperation == nil {
		choices = append([]string{"finish"}, choices...)
	}

	return choices
}

func guidedControlTerminal(status string, work *WorkflowWorkStatus) bool {
	return status == "cancelled" || status == "completed" || status == "failed" || work != nil && work.Status == "completed"
}

func guidedVisibleStatus(observed GuidedUseCaseInspection) string {
	if observed.ResultProblem != "" {
		return "blocked"
	}

	complete := observed.Result != nil && observed.Result.Complete

	terminal := observed.Session.Guided.Status == "completed" || observed.Work != nil && observed.Work.Status == "completed"
	if terminal && !complete {
		return "blocked"
	}

	return observed.Session.Guided.Status
}

type guidedResultExport struct {
	Name, Stage, Status, Problem string
	Complete                     bool
	Deliveries                   []core.WorkflowResultReceipt
	Validations                  []core.WorkflowResultValidation
	Closures                     []core.WorkflowResultClosure
	Ideas                        []core.GuidedLaterIdea
}

func guidedResultSnapshot(observed GuidedUseCaseInspection) guidedResultExport {
	snapshot := guidedResultExport{Name: observed.Session.Topic, Stage: observed.Session.Guided.Stage, Status: guidedVisibleStatus(observed), Problem: observed.ResultProblem, Ideas: observed.Ideas}
	if observed.Result != nil {
		snapshot.Complete = observed.Result.Complete
		snapshot.Deliveries = observed.Result.Receipts
		snapshot.Validations = observed.Result.Validations
		snapshot.Closures = observed.Result.Closures
	}

	return snapshot
}
