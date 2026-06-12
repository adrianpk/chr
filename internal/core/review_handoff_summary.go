package core

import (
	"fmt"
	"strings"
)

const reviewHandoffChangedFileLimit = 25

// ReviewHandoffSummary describes a package boundary value.
type ReviewHandoffSummary struct {
	ExecutionID    string
	ProjectID      string
	TicketKey      string
	TicketPath     string
	Status         ExecutionStatus
	ResultStatus   ExecutionResultStatus
	ResultSummary  string
	Validation     ExecutionReportValidation
	ChangedFiles   []ExecutionReportChangedFile
	ChangedSummary ExecutionReportChangedFilesSummary
	FailureKind    ExecutionFailureKind
	FailureReason  string
	ResultPath     string
	LogPath        string
	ReportPath     string
	NextAction     string
}

// NewReviewHandoffSummary constructs the corresponding package service or value.
func NewReviewHandoffSummary(report ExecutionReport) ReviewHandoffSummary {
	return ReviewHandoffSummary{
		ExecutionID:    report.ExecutionID,
		ProjectID:      report.ProjectID,
		TicketKey:      report.TicketKey,
		TicketPath:     report.TicketPath,
		Status:         report.Execution.Status,
		ResultStatus:   report.Result.Status,
		ResultSummary:  report.Result.Summary,
		Validation:     report.Validation,
		ChangedFiles:   append([]ExecutionReportChangedFile(nil), report.ChangedFiles...),
		ChangedSummary: report.ChangedSummary,
		FailureKind:    report.Execution.FailureKind,
		FailureReason:  report.FailureReason,
		ResultPath:     report.Artifacts.ResultPath,
		LogPath:        report.Artifacts.LogPath,
		ReportPath:     report.Artifacts.ReportPath,
		NextAction:     reviewHandoffNextAction(report),
	}
}

// RenderReviewHandoffSummaryMarkdown renders the corresponding package data.
func RenderReviewHandoffSummaryMarkdown(summary ReviewHandoffSummary) string {
	var output strings.Builder

	output.WriteString("# Review Handoff Summary\n\n")
	writeReviewHandoffOverview(&output, summary)
	writeReviewHandoffValidation(&output, summary.Validation)
	writeReviewHandoffChangedFiles(&output, summary)
	writeReviewHandoffBlocker(&output, summary)
	writeReviewHandoffArtifacts(&output, summary)
	writeReviewHandoffNextAction(&output, summary.NextAction)

	return output.String()
}

func writeReviewHandoffOverview(output *strings.Builder, summary ReviewHandoffSummary) {
	output.WriteString("## Overview\n\n")
	writeReviewHandoffField(output, "execution", summary.ExecutionID)
	writeReviewHandoffField(output, "project", summary.ProjectID)
	writeReviewHandoffField(output, "ticket", summary.TicketKey)
	writeReviewHandoffField(output, "ticket_path", summary.TicketPath)
	writeReviewHandoffField(output, "status", string(summary.Status))
	writeReviewHandoffField(output, "result_status", string(summary.ResultStatus))
	writeReviewHandoffField(output, "summary", summary.ResultSummary)
	output.WriteString("\n")
}

func writeReviewHandoffValidation(output *strings.Builder, validation ExecutionReportValidation) {
	output.WriteString("## Validation\n\n")
	writeReviewHandoffField(output, "status", string(validation.Status))
	writeReviewHandoffField(output, "command", validation.Command)
	writeReviewHandoffField(output, "path", validation.Path)
	output.WriteString("\n")
}

func writeReviewHandoffChangedFiles(output *strings.Builder, summary ReviewHandoffSummary) {
	output.WriteString("## Changed Files\n\n")
	writeReviewHandoffField(output, "total", fmt.Sprintf("%d", summary.ChangedSummary.Total))
	writeReviewHandoffField(output, "rendered", fmt.Sprintf("%d", summary.ChangedSummary.Rendered))
	writeReviewHandoffField(output, "truncated", fmt.Sprintf("%t", summary.ChangedSummary.Truncated))
	output.WriteString("\n")

	if len(summary.ChangedFiles) == 0 {
		output.WriteString("No changed files were reported.\n\n")

		return
	}

	changedFiles := summary.ChangedFiles
	if len(changedFiles) > reviewHandoffChangedFileLimit {
		changedFiles = changedFiles[:reviewHandoffChangedFileLimit]
	}

	for _, changedFile := range changedFiles {
		output.WriteString("- ")
		output.WriteString(emptyReviewHandoffValue(changedFile.Path))

		description := strings.TrimSpace(changedFile.Summary)
		if description != "" {
			output.WriteString(": ")
			output.WriteString(description)
		}

		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeReviewHandoffBlocker(output *strings.Builder, summary ReviewHandoffSummary) {
	output.WriteString("## Blocker\n\n")

	if strings.TrimSpace(summary.FailureReason) == "" {
		output.WriteString("No blocker or failure was recorded.\n\n")

		return
	}

	writeReviewHandoffField(output, "kind", string(summary.FailureKind))
	writeReviewHandoffField(output, "reason", summary.FailureReason)
	output.WriteString("\n")
}

func writeReviewHandoffArtifacts(output *strings.Builder, summary ReviewHandoffSummary) {
	output.WriteString("## Artifacts\n\n")
	writeReviewHandoffField(output, "result", summary.ResultPath)
	writeReviewHandoffField(output, "log", summary.LogPath)
	writeReviewHandoffField(output, "report", summary.ReportPath)
	output.WriteString("\n")
}

func writeReviewHandoffNextAction(output *strings.Builder, nextAction string) {
	output.WriteString("## Next Action\n\n")
	output.WriteString(emptyReviewHandoffValue(nextAction))
	output.WriteString("\n")
}

func reviewHandoffNextAction(report ExecutionReport) string {
	switch report.Execution.Status {
	case ExecutionStatusSucceeded:
		if report.Validation.Status == ExecutionReportValidationSucceeded {
			return "Review changed files and validation evidence."
		}

		return "Review execution result and confirm validation evidence."
	case ExecutionStatusFailed,
		ExecutionStatusBlocked:
		return "Inspect blocker details, runtime log, validation output, and result artifact."
	case ExecutionStatusPending,
		ExecutionStatusRunning:
		return "Wait for execution to finish before review."
	default:
		return "Inspect execution report before deciding the next action."
	}
}

func writeReviewHandoffField(output *strings.Builder, name, value string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": ")
	output.WriteString(emptyReviewHandoffValue(value))
	output.WriteString("\n")
}

func emptyReviewHandoffValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}

	return trimmed
}
