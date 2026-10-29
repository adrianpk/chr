package app

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/tui"
)

func projectTUIGuidedStatus(view *tui.GuidedView, observed GuidedUseCaseInspection) {
	view.Status = guidedVisibleStatus(observed)
	view.Result = tuiGuidedResult(observed)
	view.Controls = tuiGuidedControlChoices(observed)
	view.CanRetainIdea = len(observed.Session.Guided.Checkpoints) != 0

	command, ready := guidedControlCommand(*observed.Session)
	if !ready {
		command, ready = guidedProgressCommand(*observed.Session)
	}

	if ready {
		view.Retry = &tui.GuidedRequest{Action: "input", Command: command}
	}

	view.Progress = append(view.Progress, tuiGuidedOperationLines(observed)...)
	view.Progress = append(view.Progress, tuiGuidedAuthoringLines(observed)...)
	view.Progress = append(view.Progress, tuiGuidedWorkLines(observed)...)

	view.Progress = append(view.Progress, tuiGuidedResultLines(observed)...)
	for _, idea := range observed.Ideas {
		view.Progress = append(view.Progress, "Later idea, pending review: "+idea.Content)
	}

	if work := observed.Work; work != nil {
		view.NativeRevision = work.Revision
		view.CanApprove = workflowApprovalUnavailable(*work) == ""
		native := tuiWorkflowView(*work)
		view.Native = &native
	}
}

func tuiGuidedWorkLines(observed GuidedUseCaseInspection) []string {
	lines := tuiGuidedStopLines(observed)

	work := observed.Work
	if work == nil {
		return lines
	}

	lines = append(lines, fmt.Sprintf("Work: %s. %d of %d validated; %d confirmed; %d remaining.", strings.ReplaceAll(work.Status, "_", " "), work.ValidatedTasks, work.Total, work.ConfirmedTickets, work.Remaining))

	if work.PullRequest != nil {
		lines = append(lines, "Review PR: "+work.PullRequest.URL, "Review branch: "+work.PullRequest.Request.HeadBranch+". Delivery requires a confirmed receipt.")
	}

	lines = append(lines, fmt.Sprintf("Retained limits: %d worker attempts; %d integration attempts.", work.FrozenPolicy.MaxWorkerAttempts, work.FrozenPolicy.MaxIntegrationAttempts))
	if work.Blocker != nil {
		lines = append(lines, work.Blocker.Reason+". Next action owner: "+work.Blocker.NextActionOwner)
	}

	if work.ReviewState == "waiting_approval" || work.ReviewState == "human_required" {
		lines = append(lines, "Native review requires approval before execution can continue.")
	}

	if work.Integration != nil && work.Integration.Blocker != nil {
		lines = append(lines, "Required integration action: "+work.Integration.Blocker.RequiredAction)
	}

	return lines
}

func tuiGuidedStopLines(observed GuidedUseCaseInspection) []string {
	control := observed.Session.Guided.Control
	if control == nil || control.StopIntent == "" {
		return nil
	}

	lines := []string{"Dispatch stopped: " + control.StopIntent + ". Saved input and consumed work limits are retained."}
	if control.NativeRequest != nil && control.AcknowledgedRevision == 0 {
		lines = append(lines, "Native stop is requested; worker quiescence is not yet confirmed.")
	}

	return lines
}

func tuiGuidedResultLines(observed GuidedUseCaseInspection) []string {
	if observed.ResultProblem != "" {
		return []string{observed.ResultProblem}
	}

	result := observed.Result
	if result == nil {
		return []string{"Delivery is unconfirmed; native completion evidence is unavailable."}
	}

	text := "Delivery is unconfirmed for the remaining work."
	if result.Complete {
		text = "All admitted work has confirmed delivery and required closing evidence."
	}

	lines := []string{text, fmt.Sprintf("Verified deliveries: %d; validation receipts: %d; closing receipts: %d.", len(result.Receipts), len(result.Validations), len(result.Closures))}
	for _, evidence := range result.Receipts {
		receipt := evidence.Receipt

		lines = append(lines, "Code: "+receipt.Repository+" · "+evidence.Proof.Git.TargetBranch+" · "+receipt.ResultCommit)
		if receipt.PullRequest != "" {
			lines = append(lines, receipt.PullRequest)
		}
	}

	return lines
}
