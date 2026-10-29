package app

import (
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

func tuiGuidedResult(observed GuidedUseCaseInspection) *tui.GuidedResultView {
	if observed.ResultProblem != "" || observed.Result == nil {
		return nil
	}

	result := observed.Result

	view := &tui.GuidedResultView{Complete: result.Complete, ClosingCount: len(result.Closures), Summary: "Delivery is unconfirmed for the remaining work. Native waits and required actions remain in effect."}
	if result.Complete {
		view.Summary = "All admitted work has confirmed delivery and required closing evidence."
	}

	if result.Invocation.Status == workflowstate.InvocationCancelled {
		view.Summary = "Cancelled. Confirmed deliveries are preserved; the remaining work has not been delivered."
	}

	for _, evidence := range result.Receipts {
		receipt := evidence.Receipt
		view.Deliveries = append(view.Deliveries, tui.GuidedDeliveryView{Repository: receipt.Repository, Branch: evidence.Proof.Git.TargetBranch, Commit: receipt.ResultCommit, PullRequest: receipt.PullRequest, Method: receipt.Method, Reference: evidence.Reference.Path, Hash: evidence.Reference.Hash})
	}

	for _, validation := range result.Validations {
		view.Validation = append(view.Validation, tuiGuidedValidation(validation.Title+" ("+validation.Ticket+")", validation.Outcome, validation.Reference.Path, validation.Waiver, validation.Commands))
	}

	for _, closure := range result.Closures {
		view.Validation = append(view.Validation, tuiGuidedValidation("Set closing", closure.Receipt.Outcome, closure.Reference.Ref, closure.Receipt.Waiver, closure.Receipt.Validation))
	}

	return view
}

func tuiGuidedValidation(label, outcome, reference, waiver string, commands []workflowstate.CommandResult) tui.GuidedValidationView {
	view := tui.GuidedValidationView{Label: label, Outcome: outcome, Reference: reference, Waiver: waiver}
	for _, command := range commands {
		view.Commands = append(view.Commands, tui.GuidedValidationCommand{Command: strings.Join(command.Command, " "), ExitCode: command.ExitCode})
	}

	return view
}
