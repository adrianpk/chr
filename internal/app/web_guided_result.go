package app

import (
	"crypto/rand"
	"strconv"
	"strings"
	"time"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	"forge.adrianpk.com/chr/internal/webui"
)

func guidedWebIdeas(view *webui.GuidedUseCaseView, observed GuidedUseCaseInspection) {
	view.CanRetainIdea = len(observed.Session.Guided.Checkpoints) != 0

	view.IdeaRequestID, view.IdeaRevision = "idea-"+rand.Text(), strconv.FormatInt(observed.Session.Revision, 10)
	for _, idea := range observed.Ideas {
		view.Ideas = append(view.Ideas, webui.GuidedUseCaseIdea{Content: idea.Content, Actor: idea.Actor, CreatedAt: idea.CreatedAt.Format(time.RFC3339)})
	}
}

func guidedWebResult(view *webui.GuidedUseCaseView, observed GuidedUseCaseInspection) {
	if observed.ResultProblem != "" {
		if view.Status == "completed" {
			view.Status, view.Blocker = "blocked", observed.ResultProblem
		}

		return
	}

	result := observed.Result
	if result == nil {
		return
	}

	projected := &webui.GuidedUseCaseResult{Complete: result.Complete, ClosingCount: len(result.Closures), Summary: "Delivery is unconfirmed for the remaining work. Native waits and required actions remain in effect."}
	if result.Complete {
		projected.Summary = "All admitted work has confirmed delivery and required closing evidence."
	}

	if result.Invocation.Status == "cancelled" {
		projected.Summary = "Cancelled. Confirmed deliveries are preserved; the remaining work has not been delivered."
	}

	for _, evidence := range result.Receipts {
		receipt := evidence.Receipt
		projected.Deliveries = append(projected.Deliveries, webui.GuidedUseCaseDelivery{Repository: receipt.Repository, Branch: evidence.Proof.Git.TargetBranch, Commit: receipt.ResultCommit, PullRequest: receipt.PullRequest, Method: receipt.Method, Reference: evidence.Reference.Path, Hash: evidence.Reference.Hash})
	}

	for _, validation := range result.Validations {
		projected.Validation = append(projected.Validation, guidedWebValidation(validation.Title+" ("+validation.Ticket+")", validation.Outcome, validation.Reference.Path, validation.Waiver, validation.Commands))
	}

	for _, closure := range result.Closures {
		projected.Validation = append(projected.Validation, guidedWebValidation("Set closing", closure.Receipt.Outcome, closure.Reference.Ref, closure.Receipt.Waiver, closure.Receipt.Validation))
	}

	view.Result = projected
}

func guidedWebValidation(label, outcome, reference, waiver string, commands []workflowcore.CommandResult) webui.GuidedUseCaseValidation {
	result := webui.GuidedUseCaseValidation{Label: label, Outcome: outcome, Reference: reference, Waiver: waiver}
	for _, command := range commands {
		result.Commands = append(result.Commands, webui.GuidedUseCaseValidationCommand{Command: strings.Join(command.Command, " "), ExitCode: command.ExitCode})
	}

	return result
}
