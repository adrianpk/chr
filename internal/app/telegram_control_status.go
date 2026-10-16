package app

import (
	"fmt"
	"strings"
)

func telegramControlStatusText(status WorkflowWorkStatus) string {
	var text strings.Builder

	integration := status.Integration
	if integration == nil || integration.EntryID == "" {
		return ""
	}

	fmt.Fprintf(&text, "Integration: %s; entry revision: %d; budget revision: %d; candidate: %s.\n", integration.EntryID, integration.EntryRevision, integration.Budget.Revision, integration.CandidateHash)

	decision := integration.Decision
	if decision == nil {
		return text.String()
	}

	fmt.Fprintf(&text, "Questions: %s; revision: %d; hash: %s; state: %s.\n", decision.ID, decision.Revision, decision.Hash, decision.State)

	for _, question := range decision.Requests[:min(3, len(decision.Requests))] {
		fmt.Fprintf(&text, "Question %s: %s\n", question.ID, telegramControlExcerpt(question.Issue))

		for index, option := range question.Options {
			fmt.Fprintf(&text, "  %d. %s\n", index+1, telegramControlExcerpt(option))
		}
	}

	if len(decision.Requests) > 3 {
		text.WriteString("More retained questions are available in canonical invocation status.\n")
	}

	return text.String()
}

func telegramControlExcerpt(value string) string {
	runes := []rune(value)
	if len(runes) <= 256 {
		return value
	}

	return string(runes[:256]) + "… [full text in canonical invocation status]"
}
