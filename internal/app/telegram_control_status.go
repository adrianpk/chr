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
	decision := integration.Decision
	if decision == nil || len(decision.Requests) == 0 {
		return fmt.Sprintf("Integration state: %s.\n", integration.Phase)
	}
	question := decision.Requests[0]
	fmt.Fprintf(&text, "%s\n\n", telegramControlExcerpt(question.Issue))
	for index, option := range question.Options[:min(8, len(question.Options))] {
		fmt.Fprintf(&text, "%d. %s\n", index+1, telegramControlExcerpt(option))
	}
	fmt.Fprintf(&text, "Request a bound answer preview with /answer-decision %s %s resolve <option-number> <reason>.\n", status.ID, question.ID)
	if len(decision.Requests) > 1 || len(question.Options) > 8 {
		text.WriteString("This view is shortened. Inspect the complete retained questions and options before answering.\n")
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
