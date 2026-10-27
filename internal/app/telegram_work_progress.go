package app

import (
	"fmt"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramWorkObservation(status WorkflowWorkStatus, projectName string) telegramcore.WorkObservation {
	phase, kind := "executing", "progress"

	switch status.Status {
	case "completed":
		phase, kind = "completed", "result"
	case "cancelled":
		phase, kind = "canceled", "result"
	case "failed":
		phase, kind = "failed", "result"
	case "paused", "blocked", "waiting_capacity", "waiting_review", "waiting_integration":
		phase = "waiting"
	}

	if kind != "result" && retainedWorkAwaitingAuthority(status) {
		phase, kind = "waiting", "diagnostic"
	}

	message := telegramWorkProgressMessage(status, projectName)

	return telegramcore.WorkObservation{Revision: status.Revision, Phase: phase, Kind: kind, Text: tgmessage.Render(message).PlainFallback, Message: &message}
}

func telegramWorkProgressText(status WorkflowWorkStatus) string {
	return tgmessage.Render(telegramWorkProgressMessage(status, "")).PlainFallback
}

func telegramWorkProgressMessage(status WorkflowWorkStatus, projectName string) tgmessage.Blocks {
	title := telegramWorkLabel(status.WorkTitle)
	if title == "" {
		title = "Selected " + strings.ReplaceAll(status.Selector.Extent, "_", " ") + " work"
	}

	if projectName != "" {
		title += " in " + telegramWorkLabel(projectName)
	}

	content := fmt.Sprintf("Validated tasks: %d/%d; confirmed tickets: %d; remaining tickets: %d.", status.ValidatedTasks, status.Total, status.ConfirmedTickets, status.Remaining)
	if status.IntegratedUnits > 0 {
		content += fmt.Sprintf("\nIntegrated units: %d.", status.IntegratedUnits)
	}

	if status.Blocker != nil {
		content += "\nBlocker: " + telegramControlExcerpt(status.Blocker.Reason)
	}

	if status.PullRequest != nil {
		content += "\nPull request: " + status.PullRequest.URL
	}

	secondary := "No confirmed integration receipt. Validation or an open pull request alone does not confirm integration."
	if status.DeliveryRef != "" && status.DeliveryHash != "" {
		secondary = "Integration confirmed by retained native receipts."
	}

	secondary += "\n" + telegramWorkNextAction(status)
	secondary += "\nInspect complete retained state locally: chr work status --json " + status.ID

	return tgmessage.Blocks{Family: "33", InitialText: telegramWorkStateText(status) + "\n" + title, Content: content, SecondaryText: secondary}
}

func telegramWorkStateText(status WorkflowWorkStatus) string {
	if status.StopIntent != "" && status.Status != "paused" && status.Status != "cancelled" {
		return "Stop requested; an active worker may still be stopping."
	}

	states := map[string]string{"pending": "Work is admitted; execution has not started.", "running": "Work is running.", "admitted": "Work is admitted; execution has not started.", "waiting_capacity": "Work is waiting for execution capacity.", "waiting_review": "Work is waiting for review.", "waiting_integration": "Work is waiting for integration.", "paused": "Work is paused; saved results are retained.", "blocked": "Work is blocked; saved results are retained.", "completed": "Work has finished.", "cancelled": "Work was canceled; saved results are retained.", "failed": "Work failed; saved results are retained."}
	if state, ok := states[status.Status]; ok {
		return state
	}

	return "Current work state: " + status.Status + "."
}

func telegramWorkNextAction(status WorkflowWorkStatus) string {
	if status.Outcome == "recovered" {
		return "Recovered the same saved work. Inspect its current state before authorizing another step."
	}

	owners := map[string]string{"host": "Next: the existing workflow host continues when eligible.", "current_owner": "Next: wait for the current worker or host to release capacity.", "operator": "Next: inspect the current state with /status " + status.ID + " before approving a control.", "none": "No further action is requested."}
	if text, ok := owners[status.NextActionOwner]; ok {
		return text
	}

	return "Inspect the current state with /status " + status.ID + " before continuing."
}

func telegramWorkLabel(value string) string {
	runes := []rune(value)
	if len(runes) > 80 {
		return string(runes[:80]) + "… [shortened]"
	}

	return value
}

func telegramWorkProjectName(projects []projectcore.Project, id string) string {
	for _, project := range projects {
		if project.ID == id {
			return project.Name
		}
	}

	return ""
}
