package tui

import (
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/app/work"
)

func renderObservedWork(contexts []work.Context, limit int, empty string) []string {
	if len(contexts) == 0 {
		return []string{empty}
	}

	lines := []string{}
	for _, context := range contexts[:min(len(contexts), limit)] {
		lines = append(lines, observedWorkLines(context)...)
	}

	if len(contexts) > limit {
		lines = append(lines, fmt.Sprintf("+%d more contexts", len(contexts)-limit))
	}

	return lines
}

func observedWorkLines(context work.Context) []string {
	status := strings.ReplaceAll(context.Status, "_", " ")
	if context.Kind == "execution" && context.Status == "succeeded" {
		status = "Harness succeeded"
	}

	lines := []string{
		fmt.Sprintf("%s %s  %s  %s", context.Kind, status, context.TicketKey, context.Title),
		fmt.Sprintf("Context: %s; phase: %s", context.ID, context.Phase),
	}
	if context.Kind == "native" {
		lines = append(lines, fmt.Sprintf("Host-validated tasks: %d; confirmed tickets: %d; remaining: %d", context.Validated, context.Confirmed, context.Remaining))
	}

	if context.Kind == "loop" {
		lines = append(lines, fmt.Sprintf("%d eligible / %d processed / %d blocked / %d failed", context.Eligible, context.Processed, context.Blocked, context.Failed))
	}

	if context.Scope != "" {
		lines = append(lines, "Case: "+context.Scope)
	}

	if context.Message != "" {
		lines = append(lines, context.Message)
	}

	if !context.ActivityAt.IsZero() {
		lines = append(lines, "Activity: "+context.ActivityAt.UTC().Format(time.RFC3339))
	}

	return lines
}

func (m *Model) renderObservedActivity() []string {
	lines := []string{"Results"}
	lines = append(lines, renderObservedWork(m.workObservation.Results, 16, "No retained results.")...)
	lines = append(lines, "", "Recent activity")

	events := m.workObservation.Events
	if len(events) == 0 {
		lines = append(lines, "No recent activity.")
	}

	for _, event := range events[:min(len(events), 128)] {
		lines = append(lines, fmt.Sprintf("%s %s %s: %s", event.Time.UTC().Format(time.RFC3339), event.TicketKey, event.Phase, event.Message),
			"Context: "+event.ContextID+"; event: "+event.ID)
	}

	if len(events) > 128 {
		lines = append(lines, fmt.Sprintf("+%d more events", len(events)-128))
	}

	for _, warning := range m.workObservation.Warnings {
		lines = append(lines, "Warning: "+warning)
	}

	return lines
}
