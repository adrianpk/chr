package chc

import (
	"strings"
	"testing"
	"time"
)

func TestNewLoopReportBuildsSummaryFromLoopRun(t *testing.T) {
	record := testLoopReportRecord()

	report, err := NewLoopReport(LoopReportOptions{Loop: record})
	if err != nil {
		t.Fatalf("NewLoopReport returned error: %v", err)
	}

	if report.LoopID != record.ID {
		t.Fatalf("LoopID = %q, want %q", report.LoopID, record.ID)
	}

	if report.Counts.Processed != 2 {
		t.Fatalf("Counts.Processed = %d, want 2", report.Counts.Processed)
	}

	if report.ExecutionSummary.Total != 0 {
		t.Fatalf("ExecutionSummary.Total = %d, want 0", report.ExecutionSummary.Total)
	}
}

func TestNewLoopReportIncludesRelatedExecutions(t *testing.T) {
	record := testLoopReportRecord()
	executions := []ExecutionRecord{
		testLoopReportExecution("exec-1", "CHC-1", ExecutionStatusSucceeded),
		testLoopReportExecution("exec-2", "CHC-2", ExecutionStatusBlocked),
	}

	report, err := NewLoopReport(LoopReportOptions{
		Loop:       record,
		Executions: executions,
	})
	if err != nil {
		t.Fatalf("NewLoopReport returned error: %v", err)
	}

	if len(report.Executions) != 2 {
		t.Fatalf("Executions len = %d, want 2", len(report.Executions))
	}

	if report.ExecutionSummary.Total != 2 {
		t.Fatalf("ExecutionSummary.Total = %d, want 2", report.ExecutionSummary.Total)
	}
}

func TestNewLoopReportBoundsExecutions(t *testing.T) {
	record := testLoopReportRecord()
	executions := []ExecutionRecord{}

	for index := 0; index < 30; index++ {
		executions = append(executions, testLoopReportExecution(
			"exec-"+string(rune('a'+index%26)),
			"CHC-1",
			ExecutionStatusSucceeded,
		))
	}

	report, err := NewLoopReport(LoopReportOptions{
		Loop:       record,
		Executions: executions,
	})
	if err != nil {
		t.Fatalf("NewLoopReport returned error: %v", err)
	}

	if len(report.Executions) != loopReportExecutionLimit {
		t.Fatalf("Executions len = %d, want %d", len(report.Executions), loopReportExecutionLimit)
	}

	if report.ExecutionSummary.Total != 30 {
		t.Fatalf("ExecutionSummary.Total = %d, want 30", report.ExecutionSummary.Total)
	}

	if !report.ExecutionSummary.Truncated {
		t.Fatal("ExecutionSummary.Truncated = false, want true")
	}
}

func TestRenderLoopReportMarkdownIncludesLoopEvidence(t *testing.T) {
	record := testLoopReportRecord()
	executions := []ExecutionRecord{
		testLoopReportExecution("exec-1", "CHC-1", ExecutionStatusSucceeded),
		testLoopReportExecution("exec-2", "CHC-2", ExecutionStatusBlocked),
	}

	report, err := NewLoopReport(LoopReportOptions{
		Loop:       record,
		Executions: executions,
	})
	if err != nil {
		t.Fatalf("NewLoopReport returned error: %v", err)
	}

	markdown := RenderLoopReportMarkdown(report)

	for _, want := range []string{
		"# Loop Report",
		"## Summary",
		"- loop: loop-1",
		"- project: project-1",
		"- status: blocked",
		"## Counts",
		"- eligible: 3",
		"- processed: 2",
		"- blocked: 1",
		"## Executions",
		"- total: 2",
		"- exec-1 CHC-1 succeeded",
		"- exec-2 CHC-2 blocked",
		"## Failure",
		"- reason: ticket blocked",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}

func TestRenderLoopReportMarkdownHandlesMissingExecutions(t *testing.T) {
	record := testLoopReportRecord()
	record.FailureReason = ""

	report, err := NewLoopReport(LoopReportOptions{Loop: record})
	if err != nil {
		t.Fatalf("NewLoopReport returned error: %v", err)
	}

	markdown := RenderLoopReportMarkdown(report)

	for _, want := range []string{
		"No executions were recorded.",
		"No failure was recorded.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}

func testLoopReportRecord() LoopRunRecord {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	startedAt := now.Add(time.Minute)
	finishedAt := now.Add(3 * time.Minute)

	return LoopRunRecord{
		ID:             "loop-1",
		ProjectID:      "project-1",
		Status:         LoopRunStatusBlocked,
		EligibleCount:  3,
		ProcessedCount: 2,
		BlockedCount:   1,
		FailedCount:    0,
		FailureReason:  "ticket blocked",
		CreatedAt:      now,
		StartedAt:      &startedAt,
		FinishedAt:     &finishedAt,
		UpdatedAt:      finishedAt,
	}
}

func testLoopReportExecution(id, ticketKey string, status ExecutionStatus) ExecutionRecord {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)

	return ExecutionRecord{
		ID:        id,
		ProjectID: "project-1",
		TicketKey: ticketKey,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
