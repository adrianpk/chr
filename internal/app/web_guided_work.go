package app

import (
	"strings"

	"forge.adrianpk.com/chr/internal/webui"
)

func guidedWebWork(view *webui.GuidedUseCaseView, status *WorkflowWorkStatus) {
	if status == nil {
		return
	}

	work := &webui.GuidedUseCaseWork{Status: status.Status, Total: status.Total, Validated: status.ValidatedTasks, Confirmed: status.ConfirmedTickets, Integrated: status.IntegratedUnits, Remaining: status.Remaining, Attempts: status.TotalAttempts}

	work.MaxWorkerAttempts, work.WorkerTimeoutSeconds = status.FrozenPolicy.MaxWorkerAttempts, status.FrozenPolicy.WorkerTimeoutSeconds
	if status.PullRequest != nil {
		work.PullRequestURL = status.PullRequest.URL
	}

	if status.Blocker != nil {
		work.Action = "Native prerequisite: " + status.Blocker.Reason + ". Next action owner: " + status.Blocker.NextActionOwner + "."
	}

	if status.ReviewState == "waiting_approval" || status.ReviewState == "human_required" {
		work.Action = "Review and approve the native work before it can continue."
	}

	if status.Integration != nil && status.Integration.Blocker != nil {
		work.Action = "Required integration action: " + status.Integration.Blocker.RequiredAction
	}

	guidedWebIntegration(work, status.Integration)

	view.Work = work
}

func guidedWebIntegration(work *webui.GuidedUseCaseWork, integration *WorkflowIntegrationStatus) {
	if integration == nil {
		return
	}

	work.Publication, work.IntegrationPhase = integration.Publication, strings.ReplaceAll(integration.Phase, "_", " ")
	if integration.Decision != nil && integration.Decision.State == "waiting" {
		work.Action = "Answer the retained native integration decision before work can continue."
	}
}
