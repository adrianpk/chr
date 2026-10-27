package core

import (
	"strings"
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestReviewHandoffSummaryRendersSucceededExecution(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	result := executioncore.ExecutionResult{
		ExecutionID:     record.ID,
		TicketKey:       record.TicketKey,
		Status:          executioncore.ExecutionResultStatusSucceeded,
		Summary:         "ready for review",
		EvidenceSummary: testExecutionResultEvidenceSummary(),
		ChangedFiles: []executioncore.ExecutionResultChangedFile{
			{Path: "internal/core/review_handoff_summary.go", Summary: "add handoff summary"},
		},
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	summary := NewReviewHandoffSummary(report)
	markdown := RenderReviewHandoffSummaryMarkdown(summary)

	for _, want := range []string{
		"# Review Handoff Summary",
		"- execution: exec-1",
		"- ticket: CHR-1",
		"- status: succeeded",
		"- result_status: succeeded",
		"- summary: ready for review",
		"- status: succeeded",
		"- internal/core/review_handoff_summary.go: add handoff summary",
		"No blocker or failure was recorded.",
		"Review changed files and validation evidence.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}

func TestReviewHandoffSummaryRendersFailedExecution(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	record.Status = executioncore.ExecutionStatusFailed
	record.FailureReason = "validation: make verify failed"
	result := executioncore.ExecutionResult{
		ExecutionID:     record.ID,
		TicketKey:       record.TicketKey,
		Status:          executioncore.ExecutionResultStatusFailed,
		Summary:         "validation failed",
		EvidenceSummary: testExecutionResultEvidenceSummary(),
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	markdown := RenderReviewHandoffSummaryMarkdown(NewReviewHandoffSummary(report))

	for _, want := range []string{
		"- status: failed",
		"- result_status: failed",
		"- status: failed",
		"- kind: validation",
		"- reason: validation: make verify failed",
		"Inspect blocker details, runtime log, validation output, and result artifact.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}

func TestReviewHandoffSummaryRendersBlockedExecution(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	record.Status = executioncore.ExecutionStatusBlocked
	record.FailureReason = "worker: codex could not resolve compile error"
	result := executioncore.ExecutionResult{
		ExecutionID:     record.ID,
		TicketKey:       record.TicketKey,
		Status:          executioncore.ExecutionResultStatusFailed,
		Summary:         "blocked by compile error",
		EvidenceSummary: testExecutionResultEvidenceSummary(),
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	markdown := RenderReviewHandoffSummaryMarkdown(NewReviewHandoffSummary(report))

	for _, want := range []string{
		"- status: blocked",
		"- kind: worker",
		"- reason: worker: codex could not resolve compile error",
		"Inspect blocker details, runtime log, validation output, and result artifact.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}
