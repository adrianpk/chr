package app

import (
	"fmt"
	"strings"
)

func telegramWorkEvidenceText(status WorkflowWorkStatus) string {
	var text strings.Builder
	if status.ApprovalHash != "" {
		fmt.Fprintf(&text, "Exact review binding: %s\n", status.ApprovalHash)
	}

	if status.Blocker != nil {
		fmt.Fprintf(&text, "Blocker: %s; next action owner: %s\n", status.Blocker.Reason, status.Blocker.NextActionOwner)
	}

	for _, evidence := range status.Evidence[max(0, len(status.Evidence)-3):] {
		fmt.Fprintf(&text, "Native evidence: %s; %s; %s\n", evidence.ID, evidence.Purpose, evidence.Outcome)

		if evidence.ResultRef != "" {
			fmt.Fprintf(&text, "Evidence reference: %s\nEvidence hash: %s\n", evidence.ResultRef, evidence.ResultHash)
		}
	}

	return text.String()
}
