package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestReportListerListsProjectReports(t *testing.T) {
	projectDir := t.TempDir()
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-ticket-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHR-1\n- status: succeeded\n")
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00002-agent-work-loop",
		"evidence",
		"loop-report.md",
	), "# Loop Report\n\n## Summary\n\n- loop: loop-1\n- status: blocked\n")
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-ticket-run",
		"evidence",
		"runtime.log",
	), "log")

	reports, err := NewReportLister().ListProjectReports(context.Background(), projectDir, ReportListFilter{})
	if err != nil {
		t.Fatalf("ListProjectReports returned error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("reports len = %d, want 2: %#v", len(reports), reports)
	}

	if reports[0].Kind != ReportKindExecution {
		t.Fatalf("reports[0].Kind = %q, want execution", reports[0].Kind)
	}

	if reports[0].ExecutionID != "exec-1" {
		t.Fatalf("reports[0].ExecutionID = %q, want exec-1", reports[0].ExecutionID)
	}
}

func TestReportListerFiltersReports(t *testing.T) {
	projectDir := t.TempDir()
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-ticket-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHR-1\n- status: succeeded\n")
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00002-ticket-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-2\n- ticket: CHR-2\n- status: blocked\n")

	reports, err := NewReportLister().ListProjectReports(context.Background(), projectDir, ReportListFilter{
		Kind:      ReportKindExecution,
		TicketKey: "CHR-2",
		Status:    "blocked",
	})
	if err != nil {
		t.Fatalf("ListProjectReports returned error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("reports len = %d, want 1: %#v", len(reports), reports)
	}

	if reports[0].ExecutionID != "exec-2" {
		t.Fatalf("ExecutionID = %q, want exec-2", reports[0].ExecutionID)
	}
}

func TestReportListerReturnsEmptyWhenRunDirectoryIsMissing(t *testing.T) {
	projectDir := t.TempDir()

	reports, err := NewReportLister().ListProjectReports(context.Background(), projectDir, ReportListFilter{})
	if err != nil {
		t.Fatalf("ListProjectReports returned error: %v", err)
	}

	if len(reports) != 0 {
		t.Fatalf("reports len = %d, want 0", len(reports))
	}
}

func TestReportListerUsesActiveOpsNamespace(t *testing.T) {
	projectDir := t.TempDir()
	t.Setenv(opsnamespace.UserEnv, "adrian")

	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-default-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: default-exec\n")
	writeReportListingFile(t, projectDir, filepath.Join(
		opscore.DirectoryName,
		"adrian",
		"run",
		"00002-adrian-run",
		"evidence",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: adrian-exec\n")
	writeReportListingFile(t, projectDir, filepath.Join(
		workspacecore.DirectoryName,
		UserDirectoryName,
		"runtime",
		"runs",
		"legacy-exec",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: legacy-exec\n")

	reports, err := NewReportLister().ListProjectReports(context.Background(), projectDir, ReportListFilter{})
	if err != nil {
		t.Fatalf("ListProjectReports returned error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("reports len = %d, want 1: %#v", len(reports), reports)
	}

	if reports[0].ExecutionID != "adrian-exec" {
		t.Fatalf("ExecutionID = %q, want adrian-exec", reports[0].ExecutionID)
	}
}

func writeReportListingFile(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, relativePath)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}
