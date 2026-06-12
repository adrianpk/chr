package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewExecutionReportBuildsSummaryFromRecord(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())

	report, err := NewExecutionReport(ExecutionReportOptions{Execution: record})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	if report.ExecutionID != record.ID {
		t.Fatalf("ExecutionID = %q, want %q", report.ExecutionID, record.ID)
	}

	if report.Runtime.ValidationCommand != "make verify" {
		t.Fatalf("Runtime.ValidationCommand = %q, want make verify", report.Runtime.ValidationCommand)
	}

	if report.Artifacts.ReportPath != filepath.Join(record.Artifacts.RunDir, "execution-report.md") {
		t.Fatalf("Artifacts.ReportPath = %q, want execution report path", report.Artifacts.ReportPath)
	}

	if report.Result.Available {
		t.Fatal("Result.Available = true, want false")
	}
}

func TestNewExecutionReportIncludesResultEvidence(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	result := ExecutionResult{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Status:      ExecutionResultStatusSucceeded,
		Summary:     "ready for review",
		ChangedFiles: []ExecutionResultChangedFile{
			{Path: "internal/core/execution_report.go", Summary: "add report model"},
		},
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	if !report.Result.Available {
		t.Fatal("Result.Available = false, want true")
	}

	if report.Result.Summary != "ready for review" {
		t.Fatalf("Result.Summary = %q, want ready for review", report.Result.Summary)
	}

	if report.Validation.Status != ExecutionReportValidationSucceeded {
		t.Fatalf("Validation.Status = %q, want succeeded", report.Validation.Status)
	}

	if len(report.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles len = %d, want 1", len(report.ChangedFiles))
	}
}

func TestNewExecutionReportRejectsMismatchedResult(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	result := ExecutionResult{
		ExecutionID: "exec-other",
		TicketKey:   record.TicketKey,
		Status:      ExecutionResultStatusSucceeded,
	}

	_, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err == nil {
		t.Fatal("NewExecutionReport returned nil error")
	}
}

func TestRenderExecutionReportMarkdownIncludesEvidenceSections(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	result := ExecutionResult{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Status:      ExecutionResultStatusSucceeded,
		Summary:     "ready for review",
		ChangedFiles: []ExecutionResultChangedFile{
			{Path: "internal/core/execution_report.go", Summary: "add report model"},
		},
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	markdown := RenderExecutionReportMarkdown(report)

	for _, want := range []string{
		"# Execution Report",
		"## Summary",
		"- execution: exec-1",
		"- ticket: CHC-1",
		"- status: succeeded",
		"## Runtime",
		"- validation_command: make verify",
		"## Artifacts",
		"- result: " + record.Artifacts.ResultPath,
		"- report: " + record.Artifacts.RunDir + "/execution-report.md",
		"## Validation",
		"- status: succeeded",
		"- path: " + record.Artifacts.ValidationPath,
		"## Result",
		"- summary: ready for review",
		"## Changed Files",
		"- total: 1",
		"- rendered: 1",
		"- truncated: false",
		"- internal/core/execution_report.go: add report model",
		"## Failure",
		"No failure was recorded.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}

func TestNewExecutionReportSortsAndBoundsChangedFiles(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	result := ExecutionResult{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Status:      ExecutionResultStatusSucceeded,
		Summary:     "ready for review",
	}

	for index := 29; index >= 0; index-- {
		result.ChangedFiles = append(result.ChangedFiles, ExecutionResultChangedFile{
			Path:    filepath.Join("file", string(rune('a'+index%26))+".go"),
			Summary: "changed",
		})
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	if len(report.ChangedFiles) != executionReportChangedFileLimit {
		t.Fatalf("ChangedFiles len = %d, want %d", len(report.ChangedFiles), executionReportChangedFileLimit)
	}

	if report.ChangedSummary.Total != 30 {
		t.Fatalf("ChangedSummary.Total = %d, want 30", report.ChangedSummary.Total)
	}

	if !report.ChangedSummary.Truncated {
		t.Fatal("ChangedSummary.Truncated = false, want true")
	}

	for index := 1; index < len(report.ChangedFiles); index++ {
		previous := report.ChangedFiles[index-1].Path
		current := report.ChangedFiles[index].Path

		if previous > current {
			t.Fatalf("changed files are not sorted: %q before %q", previous, current)
		}
	}
}

func TestExecutionReportWriterWritesMarkdownReport(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())
	result := ExecutionResult{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Status:      ExecutionResultStatusSucceeded,
		Summary:     "ready for review",
	}

	report, err := NewExecutionReport(ExecutionReportOptions{
		Execution: record,
		Result:    &result,
	})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	path, err := NewExecutionReportWriter().WriteExecutionReport(context.Background(), report)
	if err != nil {
		t.Fatalf("WriteExecutionReport returned error: %v", err)
	}

	if path != filepath.Join(record.Artifacts.RunDir, "execution-report.md") {
		t.Fatalf("path = %q, want execution report path", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(content), "- execution: exec-1") {
		t.Fatalf("report content missing execution id:\n%s", string(content))
	}
}

func TestExecutionReportWriterRefusesNonReportOverwrite(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())

	report, err := NewExecutionReport(ExecutionReportOptions{Execution: record})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	err = os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(report.Artifacts.ReportPath, []byte("# Something Else\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err = NewExecutionReportWriter().WriteExecutionReport(context.Background(), report)
	if err == nil {
		t.Fatal("WriteExecutionReport returned nil error")
	}
}

func TestRenderExecutionReportMarkdownHandlesMissingResult(t *testing.T) {
	record := testExecutionReportRecord(t.TempDir())

	report, err := NewExecutionReport(ExecutionReportOptions{Execution: record})
	if err != nil {
		t.Fatalf("NewExecutionReport returned error: %v", err)
	}

	markdown := RenderExecutionReportMarkdown(report)

	for _, want := range []string{
		"No execution result was available.",
		"No changed files were reported.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
}

func testExecutionReportRecord(dir string) ExecutionRecord {
	runDir := filepath.Join(dir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	startedAt := now.Add(time.Minute)
	finishedAt := now.Add(2 * time.Minute)

	return ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(dir, ".chc", "user", "tickets", "CHC-1.md"),
		Status:     ExecutionStatusSucceeded,
		Runtime: ExecutionRuntimeRef{
			Image:             "chc-agent-runtime:local",
			DockerfilePath:    filepath.Join(dir, ".chc", "user", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt:  now,
		StartedAt:  &startedAt,
		FinishedAt: &finishedAt,
		UpdatedAt:  finishedAt,
	}
}
