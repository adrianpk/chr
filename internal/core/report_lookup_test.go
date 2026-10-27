package core

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

func TestReportLookupFindsReportByExecutionID(t *testing.T) {
	projectDir := t.TempDir()
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-ticket-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHR-1\n- status: succeeded\n")

	result, err := NewReportLookupService().LookupReport(context.Background(), projectDir, ReportLookupRequest{
		ExecutionID: "exec-1",
	})
	if err != nil {
		t.Fatalf("LookupReport returned error: %v", err)
	}

	if result.Entry.ExecutionID != "exec-1" {
		t.Fatalf("ExecutionID = %q, want exec-1", result.Entry.ExecutionID)
	}

	if !strings.Contains(result.Body, "# Execution Report") {
		t.Fatalf("Body = %q, want execution report body", result.Body)
	}
}

func TestReportLookupFindsReportByRelativePath(t *testing.T) {
	projectDir := t.TempDir()
	relativePath := filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-agent-work-loop",
		"evidence",
		"loop-report.md",
	)
	writeReportListingFile(t, projectDir, relativePath, "# Loop Report\n\n## Summary\n\n- loop: loop-1\n")

	result, err := NewReportLookupService().LookupReport(context.Background(), projectDir, ReportLookupRequest{
		ReportID: filepath.ToSlash(relativePath),
	})
	if err != nil {
		t.Fatalf("LookupReport returned error: %v", err)
	}

	if result.Entry.LoopID != "loop-1" {
		t.Fatalf("LoopID = %q, want loop-1", result.Entry.LoopID)
	}
}

func TestReportLookupRejectsMissingReport(t *testing.T) {
	projectDir := t.TempDir()

	_, err := NewReportLookupService().LookupReport(context.Background(), projectDir, ReportLookupRequest{
		ExecutionID: "missing",
	})
	if err == nil {
		t.Fatal("LookupReport returned nil error")
	}

	if !strings.Contains(err.Error(), "report not found") {
		t.Fatalf("error = %v, want report not found", err)
	}
}

func TestReportLookupRejectsAmbiguousReport(t *testing.T) {
	projectDir := t.TempDir()
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-ticket-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHR-1\n")
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00002-ticket-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-2\n- ticket: CHR-1\n")

	_, err := NewReportLookupService().LookupReport(context.Background(), projectDir, ReportLookupRequest{
		TicketKey: "CHR-1",
	})
	if err == nil {
		t.Fatal("LookupReport returned nil error")
	}

	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("error = %v, want ambiguous lookup", err)
	}
}
