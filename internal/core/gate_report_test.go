package core

import "forge.adrianpk.com/chr/internal/core/gate"

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGateReportWriterWritesJSON(t *testing.T) {
	projectDir := t.TempDir()

	createdAt := time.Date(2026, 6, 1, 12, 30, 45, 0, time.UTC)

	artifacts, err := newGateTestArtifacts(projectDir, createdAt)
	if err != nil {
		t.Fatalf("NewGateReportArtifacts returned error: %v", err)
	}

	report := gate.NewGateReport(createdAt, gate.GateGitEvidence{
		Base:      gate.GateGitRef{Ref: "main", SHA: "base-sha"},
		Head:      gate.GateGitRef{Ref: "dev", SHA: "head-sha"},
		MergeBase: "base-sha",
		Commits: []gate.GateGitCommit{
			{SHA: "head-sha", Subject: "add gate"},
		},
		Files: []gate.GateGitChangedFile{
			{Status: "M", Path: "internal/core/gate.go"},
		},
		DiffStat: "1 file changed, 1 insertion(+)",
	}, gate.GateValidationResult{
		Command: gate.DefaultGateValidationCommand,
		Passed:  true,
	}, gate.GateEvidenceClassification{}, artifacts)

	path, err := gate.NewGateReportWriter().WriteJSON(context.Background(), report)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	wantPath := filepath.Join(projectDir, "ops", "default", "report", "00001-dev-main-gate-2026-06-01-12-30-45z", "evidence", "gate-report.json")
	if path != wantPath {
		t.Fatalf("path = %q, want canonical JSON report path", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	var decoded gate.GateReport

	err = json.Unmarshal(content, &decoded)
	if err != nil {
		t.Fatalf("decode report: %v", err)
	}

	if decoded.Status != gate.GateReportStatusPassed {
		t.Fatalf("Status = %q, want passed", decoded.Status)
	}

	if decoded.Base.Ref != "main" || decoded.Head.Ref != "dev" {
		t.Fatalf("decoded refs = %#v %#v, want main/dev", decoded.Base, decoded.Head)
	}
}

func TestGateReportWriterWritesMarkdown(t *testing.T) {
	projectDir := t.TempDir()

	createdAt := time.Date(2026, 6, 1, 12, 30, 45, 0, time.UTC)

	artifacts, err := newGateTestArtifacts(projectDir, createdAt)
	if err != nil {
		t.Fatalf("NewGateReportArtifacts returned error: %v", err)
	}

	report := gateReportMarkdownFixture(createdAt, artifacts)
	report.Failures = []string{"validation failed"}

	path, err := gate.NewGateReportWriter().WriteMarkdown(context.Background(), report)
	if err != nil {
		t.Fatalf("WriteMarkdown returned error: %v", err)
	}

	wantPath := filepath.Join(projectDir, "ops", "default", "report", "00001-dev-main-gate-2026-06-01-12-30-45z", "artifact.md")
	if path != wantPath {
		t.Fatalf("path = %q, want canonical Markdown report path", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	for _, want := range []string{
		"# Dev to Main Gate Report",
		"- status: failed",
		"- base: main base-sha",
		"## Agent-Managed Work",
		"ticket=CHR-1 execution=exec-1",
		"## Manual or External Changes",
		"- M README.md",
		"- command: go test ./...",
		"validation failed",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content missing %q:\n%s", want, string(content))
		}
	}
}

func gateReportMarkdownFixture(createdAt time.Time, artifacts gate.GateReportArtifacts) gate.GateReport {
	return gate.NewGateReport(createdAt, gate.GateGitEvidence{
		Base:      gate.GateGitRef{Ref: "main", SHA: "base-sha"},
		Head:      gate.GateGitRef{Ref: "dev", SHA: "head-sha"},
		MergeBase: "base-sha",
		Commits: []gate.GateGitCommit{
			{SHA: "head-sha", Subject: "add gate"},
		},
		Files: []gate.GateGitChangedFile{
			{Status: "M", Path: "internal/core/gate.go"},
		},
	}, gate.GateValidationResult{
		Command:  gate.DefaultGateValidationCommand,
		ExitCode: 1,
		Output:   "validation failed",
	}, gate.GateEvidenceClassification{
		AgentManagedWork: []gate.GateAgentManagedWork{
			{
				Path:        "internal/core/gate.go",
				Status:      "M",
				TicketKey:   "CHR-1",
				ExecutionID: "exec-1",
				ReportPath:  ".chr/default/runtime/runs/CHR-1/exec-1/execution-report.md",
			},
		},
		ManualOrExternalChanges: []gate.GateManualOrExternalChange{
			{Path: "README.md", Status: "M"},
		},
	}, artifacts)
}

func TestGateReportWriterRejectsDirectoryTarget(t *testing.T) {
	projectDir := t.TempDir()
	path := filepath.Join(projectDir, ".chr", "default", "reports", "gates", "dev-main.json")

	err := os.MkdirAll(path, 0o755)
	if err != nil {
		t.Fatalf("create report directory target: %v", err)
	}

	report := gate.GateReport{
		Artifacts: gate.GateReportArtifacts{JSONPath: path},
	}

	_, err = gate.NewGateReportWriter().WriteJSON(context.Background(), report)
	if err == nil {
		t.Fatal("WriteJSON returned nil error")
	}

	if !strings.Contains(err.Error(), "directory") {
		t.Fatalf("error = %q, want directory context", err.Error())
	}
}

func TestNewFailedGateReportRecordsRefsAndFailures(t *testing.T) {
	report := gate.NewFailedGateReport(
		time.Date(2026, 6, 1, 12, 30, 45, 0, time.UTC),
		" main ",
		" dev ",
		gate.GateReportArtifacts{},
		[]string{"resolve base ref"},
	)

	if report.Status != gate.GateReportStatusFailed {
		t.Fatalf("Status = %q, want failed", report.Status)
	}

	if report.Base.Ref != "main" || report.Head.Ref != "dev" {
		t.Fatalf("refs = %#v %#v, want trimmed refs", report.Base, report.Head)
	}

	if len(report.Failures) != 1 {
		t.Fatalf("Failures = %#v, want one failure", report.Failures)
	}
}

func newGateTestArtifacts(projectDir string, createdAt time.Time) (gate.GateReportArtifacts, error) {
	return gate.GateReportArtifacts{JSONPath: filepath.Join(projectDir, "ops", "default", "report", "00001-dev-main-gate-"+createdAt.UTC().Format("2006-01-02-15-04-05z"), "evidence", "gate-report.json"), MarkdownPath: filepath.Join(projectDir, "ops", "default", "report", "00001-dev-main-gate-"+createdAt.UTC().Format("2006-01-02-15-04-05z"), "artifact.md")}, nil
}
