package app

import (
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/tui"
)

func tuiWorkflowView(status WorkflowWorkStatus) tui.WorkflowView {
	lines := []string{
		fmt.Sprintf("Invocation: %s revision %d", status.ID, status.Revision),
		fmt.Sprintf("Extent: %s %s; authority: %s", status.Selector.Extent, status.Selector.Identity, status.Authority),
		fmt.Sprintf("State: %s; outcome: %s; stop: %s", status.Status, status.Outcome, status.StopIntent),
		fmt.Sprintf("Host-validated tasks: %d/%d; confirmed tickets: %d; integrated units: %d; remaining: %d", status.ValidatedTasks, status.Total, status.ConfirmedTickets, status.IntegratedUnits, status.Remaining),
		fmt.Sprintf("Selected ticket: %s; slice: %s; order: %s", status.SelectedTicket, status.SelectedSlice, status.SetOrderRule),
		fmt.Sprintf("Policy: %s; method: %s; source: %s", status.Policy.Mode, status.Policy.Method, status.Policy.Source),
		fmt.Sprintf("Frozen limits: worker attempts %d; integration attempts %d; checkpoint steps %d", status.FrozenPolicy.MaxWorkerAttempts, status.FrozenPolicy.MaxIntegrationAttempts, workflowDisplayStepLimit(status.FrozenPolicy.CheckpointStepLimit)),
		fmt.Sprintf("Target: %s at %s", status.Target.Repository, status.Target.Commit),
		fmt.Sprintf("Next action owner: %s; capacity resumes: %s", status.NextActionOwner, status.CapacityResumeStatus),
	}
	lines = appendWorkflowExclusions(lines, status.Excluded)
	lines = appendWorkflowBindingLines(lines, status)
	lines = append(lines, workflowIntegrationLines(status.Integration)...)
	lines = appendWorkflowMemberLines(lines, status)

	for _, evidence := range status.Evidence[max(0, len(status.Evidence)-8):] {
		lines = append(lines, fmt.Sprintf("Evidence: %s %s %s %s; slot %d; %s %s", evidence.ID, evidence.Purpose, evidence.TicketKey, evidence.Outcome, evidence.Slot, evidence.ResultRef, evidence.ResultHash))
	}

	view := tui.WorkflowView{ID: status.ID, Title: status.WorkTitle, State: status.Status, Extent: status.Selector.Extent, Identity: status.Selector.Identity, TicketKey: status.SelectedTicket, Revision: status.Revision, ApprovalHash: status.ApprovalHash, Lines: lines, Members: status.Members, NextOffset: status.NextOffset}
	for _, action := range workflowControlOptions(status) {
		view.Actions = append(view.Actions, tui.WorkflowActionView{Action: tui.WorkActionID("workflow_" + action.action), Label: action.label, Reason: action.reason})
	}

	return view
}

func appendWorkflowExclusions(lines []string, excluded map[string]int) []string {
	keys := make([]string, 0, len(excluded))
	for key := range excluded {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("Excluded %s: %d", key, excluded[key]))
	}

	return lines
}

func appendWorkflowBindingLines(lines []string, status WorkflowWorkStatus) []string {
	if status.Blocker != nil {
		lines = append(lines, fmt.Sprintf("Blocker: %s %s; %s; owner: %s", status.Blocker.Dependency.Kind, status.Blocker.Dependency.ID, status.Blocker.Reason, status.Blocker.NextActionOwner))
	}

	if status.CandidateHash != "" {
		lines = append(lines, "Candidate: "+status.CandidateRef+" "+status.CandidateHash)
	}

	if status.PullRequest != nil {
		lines = append(lines, fmt.Sprintf("PR: %s; repository %s; state %s; head %s", status.PullRequest.URL, status.PullRequest.Request.Repository, status.PullRequest.State, status.PullRequest.SourceCommit))
	}

	if status.ApprovalHash != "" {
		lines = append(lines, "Approval binding: "+status.ApprovalHash+"; state: "+status.ReviewState)
	}

	if status.DeliveryHash != "" {
		lines = append(lines, "Delivery receipt: "+status.DeliveryRef+" "+status.DeliveryHash)
	}

	if status.Continuation != nil {
		lines = append(lines, fmt.Sprintf("Continuation: %d %s; evidence %s %s", status.Continuation.Sequence, status.Continuation.Reason, status.Continuation.EvidenceRef, status.Continuation.EvidenceHash))
	}

	return lines
}

func appendWorkflowMemberLines(lines []string, status WorkflowWorkStatus) []string {
	for _, member := range status.Members {
		lines = append(lines, fmt.Sprintf("Member: %s revision %d; set %s; slice %s; order %d; phase %s; readiness %s", member.Key, member.Revision, member.SetID, member.SliceID, member.TaskOrder, member.Phase, member.Readiness))
		if member.Blocker != nil {
			lines = append(lines, fmt.Sprintf("Member blocker: %s; owner %s", member.Blocker.Reason, member.Blocker.NextActionOwner))
		}
	}

	if status.NextOffset >= 0 {
		lines = append(lines, fmt.Sprintf("More members: chr work status %s --offset %d --revision %d", status.ID, status.NextOffset, status.Revision))
	}

	return lines
}

func nativeWorkPreviewLines(bound WorkflowBoundPreview) []string {
	view := bound.View

	lines := []string{
		fmt.Sprintf("Extent: %s %s; authority: %s", view.Preview.Selector.Extent, view.Preview.Selector.Identity, view.Preview.Authority),
		fmt.Sprintf("Admitted tickets: %d; order: %s", view.Preview.Total, view.Preview.SetOrderRule),
		"Snapshot: " + view.Preview.SnapshotHash,
		fmt.Sprintf("Target: %s at %s", view.Target.Repository, view.Target.Commit),
		fmt.Sprintf("Policy: %s; method: %s; source: %s", view.Policy.Mode, view.Policy.Method, view.Policy.Source),
		fmt.Sprintf("Frozen limits: worker attempts %d; integration attempts %d; checkpoint steps %d", view.ExecutionPolicy.MaxWorkerAttempts, view.ExecutionPolicy.MaxIntegrationAttempts, workflowDisplayStepLimit(view.ExecutionPolicy.CheckpointStepLimit)),
		fmt.Sprintf("Timeouts: worker %ds; validation %ds; token budget %d", view.ExecutionPolicy.WorkerTimeoutSeconds, view.ExecutionPolicy.ValidationTimeoutSeconds, view.ExecutionPolicy.AgentTokenBudget),
		"Execute at most one shared checkpoint; continue from retained status.",
	}
	if bound.Selected != nil {
		lines = append(lines, fmt.Sprintf("Selected ticket: %s revision %d", bound.Selected.Key, bound.Selected.Revision))
	}

	lines = appendWorkflowExclusions(lines, view.Preview.Excluded)
	status := WorkflowWorkStatus{WorkflowWorkPage: WorkflowWorkPage{Members: view.Preview.Members[:min(8, len(view.Preview.Members))], NextOffset: -1}}

	lines = appendWorkflowMemberLines(lines, status)
	if view.Preview.Total > len(status.Members) {
		lines = append(lines, fmt.Sprintf("Showing first %d of %d selected members; the snapshot binds the complete extent.", len(status.Members), view.Preview.Total))
	}

	return lines
}

func workflowDisplayStepLimit(limit int) int {
	if limit == 0 {
		return 128
	}

	return limit
}
