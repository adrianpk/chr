package core

import (
	"fmt"
	"strings"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

// LoopReport describes a package boundary value.
type LoopReport struct {
	LoopID           string
	ProjectID        string
	Status           executioncore.LoopRunStatus
	Counts           LoopReportCounts
	StopReason       string
	FailureReason    string
	Executions       []LoopReportExecution
	ExecutionSummary LoopReportExecutionSummary
	CreatedAt        time.Time
	StartedAt        *time.Time
	FinishedAt       *time.Time
	UpdatedAt        time.Time
}

// LoopReportCounts describes a package boundary value.
type LoopReportCounts struct {
	Eligible  int
	Processed int
	Blocked   int
	Failed    int
}

// LoopReportExecution describes a package boundary value.
type LoopReportExecution struct {
	ExecutionID string
	TicketKey   string
	Status      executioncore.Status
}

// LoopReportExecutionSummary describes a package boundary value.
type LoopReportExecutionSummary struct {
	Total     int
	Rendered  int
	Truncated bool
}

// LoopReportOptions describes a package boundary value.
type LoopReportOptions struct {
	Loop       executioncore.LoopRunRecord
	Executions []executioncore.Record
}

const loopReportExecutionLimit = 25

// NewLoopReport constructs the corresponding package service or value.
func NewLoopReport(opts LoopReportOptions) (LoopReport, error) {
	err := executioncore.ValidateLoopRunRecord(opts.Loop)
	if err != nil {
		return LoopReport{}, err
	}

	report := loopReportFromRecord(opts.Loop)
	entries := loopReportExecutions(opts.Executions)
	report.Executions, report.ExecutionSummary = entries.executions, entries.summary

	return report, nil
}

func loopReportFromRecord(record executioncore.LoopRunRecord) LoopReport {
	return LoopReport{
		LoopID:    record.ID,
		ProjectID: record.ProjectID,
		Status:    record.Status,
		Counts: LoopReportCounts{
			Eligible:  record.EligibleCount,
			Processed: record.ProcessedCount,
			Blocked:   record.BlockedCount,
			Failed:    record.FailedCount,
		},
		StopReason:    loopReportStopReason(record),
		FailureReason: record.FailureReason,
		CreatedAt:     record.CreatedAt,
		StartedAt:     record.StartedAt,
		FinishedAt:    record.FinishedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

type loopReportEntries struct {
	executions []LoopReportExecution
	summary    LoopReportExecutionSummary
}

func loopReportExecutions(records []executioncore.Record) loopReportEntries {
	executions := make([]LoopReportExecution, 0, len(records))

	for _, record := range records {
		executions = append(executions, LoopReportExecution{
			ExecutionID: record.ID,
			TicketKey:   record.TicketKey,
			Status:      record.Status,
		})
	}

	summary := LoopReportExecutionSummary{
		Total:    len(executions),
		Rendered: len(executions),
	}

	if len(executions) > loopReportExecutionLimit {
		executions = executions[:loopReportExecutionLimit]
		summary.Rendered = loopReportExecutionLimit
		summary.Truncated = true
	}

	return loopReportEntries{executions: executions, summary: summary}
}

// RenderLoopReportMarkdown renders the corresponding package data.
func RenderLoopReportMarkdown(report LoopReport) string {
	var output strings.Builder

	output.WriteString("# Loop Report\n\n")
	writeLoopReportSummary(&output, report)
	writeLoopReportCounts(&output, report.Counts)
	writeLoopReportStopReason(&output, report.StopReason)
	writeLoopReportTimes(&output, report)
	writeLoopReportExecutions(&output, report.Executions, report.ExecutionSummary)
	writeLoopReportFailure(&output, report.FailureReason)

	return output.String()
}

func loopReportStopReason(record executioncore.LoopRunRecord) string {
	switch {
	case record.Status == executioncore.LoopRunStatusFailed && strings.TrimSpace(record.FailureReason) != "":
		return record.FailureReason
	case record.Status == executioncore.LoopRunStatusFailed:
		return "loop failed before completing integration work"
	case record.BlockedCount > 0:
		return "loop stopped because the current ticket integration cycle blocked"
	case record.Status == executioncore.LoopRunStatusSucceeded && record.EligibleCount == 0:
		return "loop completed with no eligible tickets"
	case record.Status == executioncore.LoopRunStatusSucceeded:
		return "loop completed all selected integration cycles"
	case record.Status == executioncore.LoopRunStatusRunning:
		return "loop is still running"
	default:
		return "loop stop reason is unknown"
	}
}

func writeLoopReportSummary(output *strings.Builder, report LoopReport) {
	output.WriteString("## Summary\n\n")
	writeLoopReportField(output, "loop", report.LoopID)
	writeLoopReportField(output, "project", report.ProjectID)
	writeLoopReportField(output, "status", string(report.Status))
	output.WriteString("\n")
}

func writeLoopReportCounts(output *strings.Builder, counts LoopReportCounts) {
	output.WriteString("## Counts\n\n")
	writeLoopReportField(output, "eligible", fmt.Sprintf("%d", counts.Eligible))
	writeLoopReportField(output, "processed", fmt.Sprintf("%d", counts.Processed))
	writeLoopReportField(output, "blocked", fmt.Sprintf("%d", counts.Blocked))
	writeLoopReportField(output, "failed", fmt.Sprintf("%d", counts.Failed))
	output.WriteString("\n")
}

func writeLoopReportStopReason(output *strings.Builder, reason string) {
	output.WriteString("## Stop Reason\n\n")
	writeLoopReportField(output, "reason", reason)
	output.WriteString("\n")
}

func writeLoopReportTimes(output *strings.Builder, report LoopReport) {
	output.WriteString("## Times\n\n")
	writeLoopReportField(output, "created_at", formatLoopReportTime(report.CreatedAt))
	writeLoopReportField(output, "started_at", formatLoopReportOptionalTime(report.StartedAt))
	writeLoopReportField(output, "finished_at", formatLoopReportOptionalTime(report.FinishedAt))
	writeLoopReportField(output, "updated_at", formatLoopReportTime(report.UpdatedAt))
	output.WriteString("\n")
}

func writeLoopReportExecutions(output *strings.Builder, executions []LoopReportExecution, summary LoopReportExecutionSummary) {
	output.WriteString("## Executions\n\n")
	writeLoopReportField(output, "total", fmt.Sprintf("%d", summary.Total))
	writeLoopReportField(output, "rendered", fmt.Sprintf("%d", summary.Rendered))
	writeLoopReportField(output, "truncated", fmt.Sprintf("%t", summary.Truncated))
	output.WriteString("\n")

	if len(executions) == 0 {
		output.WriteString("No executions were recorded.\n\n")

		return
	}

	for _, execution := range executions {
		output.WriteString("- ")
		output.WriteString(emptyLoopReportValue(execution.ExecutionID))
		output.WriteString(" ")
		output.WriteString(emptyLoopReportValue(execution.TicketKey))
		output.WriteString(" ")
		output.WriteString(emptyLoopReportValue(string(execution.Status)))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeLoopReportFailure(output *strings.Builder, reason string) {
	output.WriteString("## Failure\n\n")

	if strings.TrimSpace(reason) == "" {
		output.WriteString("No failure was recorded.\n")

		return
	}

	writeLoopReportField(output, "reason", reason)
}

func writeLoopReportField(output *strings.Builder, name, value string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": ")
	output.WriteString(emptyLoopReportValue(value))
	output.WriteString("\n")
}

func formatLoopReportTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}

func formatLoopReportOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return formatLoopReportTime(*value)
}

func emptyLoopReportValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}

	return trimmed
}
