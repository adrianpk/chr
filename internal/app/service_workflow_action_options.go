package app

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowControlOption struct{ action, label, reason string }

// Options are display guidance. The canonical control owner rechecks current
// revision, policy, budget and claims under exclusion before any mutation.
func workflowControlOptions(status WorkflowWorkStatus) []workflowControlOption {
	options := []workflowControlOption{{action: "grant", label: "Extend allowance"}, {action: "answer_decision", label: "Answer decision"}, {action: "continue", label: "Continue"}, {action: "pause", label: "Pause"}, {action: "cancel", label: "Cancel"}, {action: "resume", label: "Resume"}, {action: "approve", label: "Approve"}, {action: "reconcile", label: "Reconcile"}}
	for index := range options {
		options[index].reason = workflowControlUnavailable(status, options[index].action)
	}

	return options
}

func workflowControlUnavailable(status WorkflowWorkStatus, action string) string {
	if workflowstate.InvocationTerminal(workflowstate.InvocationStatus(status.Status)) {
		return "This workflow is terminal; inspect its retained evidence."
	}

	switch action {
	case "continue":
		if status.Status == "running" || !retainedWorkRunnable(status) {
			return "The current owner, stop intent or retained wait must be resolved first."
		}
	case "resume":
		if !slices.Contains([]string{"paused", "blocked"}, status.Status) {
			return "Resume requires a paused or blocked workflow."
		}
	case "grant", "answer_decision":
		return workflowIntegrationControlUnavailable(status, action)
	case "approve":
		return workflowApprovalUnavailable(status)
	case "reconcile":
		return workflowReconcileUnavailable(status)
	}

	return ""
}

func workflowApprovalUnavailable(status WorkflowWorkStatus) string {
	if status.ApprovalHash == "" || !slices.Contains([]string{"waiting_approval", "human_required"}, status.ReviewState) {
		return "No exact approval binding is awaiting an operator."
	}

	return ""
}

func workflowIntegrationControlUnavailable(status WorkflowWorkStatus, action string) string {
	current := status.Integration
	if current == nil || current.EntryID == "" || current.Budget.Revision == 0 {
		return "Inspect an admitted integration unit with a retained allowance."
	}

	if action == "answer_decision" && (current.Decision == nil || current.Decision.State != "waiting" || len(current.Decision.Requests) == 0) {
		return "No retained integration question is awaiting an answer."
	}

	return ""
}

func workflowReconcileUnavailable(status WorkflowWorkStatus) string {
	if status.Status != "running" {
		return "Reconciliation is for an interrupted running owner."
	}

	return ""
}
