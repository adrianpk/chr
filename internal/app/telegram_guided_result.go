package app

import (
	"fmt"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramGuidedResultSummary(observed GuidedUseCaseInspection) string {
	if observed.ResultProblem != "" {
		return observed.ResultProblem
	}

	result := observed.Result
	if result == nil {
		return "Delivery is unconfirmed; no native completion evidence is available."
	}

	text := "Delivery is unconfirmed for the remaining work. Native waits and required actions remain in effect."
	if result.Complete {
		text = "All admitted work has confirmed delivery and required closing evidence."
	}

	if result.Invocation.Status == workflowstate.InvocationCancelled {
		text = "Cancelled. Confirmed deliveries are preserved; remaining work has not been delivered."
	}

	text += fmt.Sprintf("\nConfirmed deliveries: %d. Validation receipts: %d. Closing receipts: %d.", len(result.Receipts), len(result.Validations), len(result.Closures))
	for _, evidence := range result.Receipts[:min(2, len(result.Receipts))] {
		receipt := evidence.Receipt

		text += "\nCode: " + receipt.Repository + " · " + evidence.Proof.Git.TargetBranch + " · " + receipt.ResultCommit
		if receipt.PullRequest != "" {
			text += "\n" + receipt.PullRequest
		}
	}

	return strings.TrimSpace(text)
}
