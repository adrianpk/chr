package app

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramWorkObservation(status WorkflowWorkStatus) core.TelegramWorkObservation {
	phase, kind := "executing", "progress"

	switch status.Status {
	case "completed":
		phase, kind = "completed", "result"
	case "cancelled":
		phase, kind = "canceled", "result"
	case "paused", "blocked", "waiting_capacity", "waiting_review":
		phase = "waiting"
	}

	if kind != "result" && retainedWorkAwaitingAuthority(status) {
		phase, kind = "waiting", "diagnostic"
	}

	return core.TelegramWorkObservation{Revision: status.Revision, Phase: phase, Kind: kind, Text: telegramWorkProgressText(status)}
}

func telegramWorkProgressText(status WorkflowWorkStatus) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Native work: %s %s\nInvocation: %s\nRevision: %d; state: %s; outcome: %s\nHost-validated tasks: %d/%d; confirmed tickets: %d; integrated units: %d; remaining tickets: %d.\n", status.Selector.Extent, status.Selector.Identity, status.ID, status.Revision, status.Status, status.Outcome, status.ValidatedTasks, status.Total, status.ConfirmedTickets, status.IntegratedUnits, status.Remaining)
	fmt.Fprintf(&text, "Publication policy: %s; review: %s; next action: %s.\n", status.Policy.Mode, status.ReviewState, status.NextActionOwner)

	if status.CandidateRef != "" {
		fmt.Fprintf(&text, "Checked candidate: %s\nCandidate hash: %s\n", status.CandidateRef, status.CandidateHash)
	}

	if status.PullRequest != nil {
		fmt.Fprintf(&text, "Pull request: %s\n", status.PullRequest.URL)
	}

	if status.DeliveryRef != "" {
		fmt.Fprintf(&text, "Confirmed receipt: %s\nReceipt hash: %s\n", status.DeliveryRef, status.DeliveryHash)
	} else {
		text.WriteString("No confirmed integration receipt. Worker exit, validation and an open PR do not establish delivery.\n")
	}

	fmt.Fprintf(&text, "Project: %s; namespace: %s.\nInspect in that project: chr work status --json %s", status.ProjectID, status.Namespace, status.ID)

	return text.String()
}
