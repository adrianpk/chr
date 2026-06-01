package chc

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
	artifacts := NewGateReportArtifacts(projectDir, createdAt)
	report := NewGateReport(createdAt, GateGitEvidence{
		Base:      GateGitRef{Ref: "main", SHA: "base-sha"},
		Head:      GateGitRef{Ref: "dev", SHA: "head-sha"},
		MergeBase: "base-sha",
		Commits: []GateGitCommit{
			{SHA: "head-sha", Subject: "add gate"},
		},
		Files: []GateGitChangedFile{
			{Status: "M", Path: "internal/chc/gate.go"},
		},
		DiffStat: "1 file changed, 1 insertion(+)",
	}, GateValidationResult{
		Command: DefaultGateValidationCommand,
		Passed:  true,
	}, artifacts)

	path, err := NewGateReportWriter().WriteJSON(context.Background(), report)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	if path != filepath.Join(projectDir, ".chc", "user", "reports", "gates", "dev-main-20260601T123045Z.json") {
		t.Fatalf("path = %q, want timestamped JSON report path", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	var decoded GateReport
	err = json.Unmarshal(content, &decoded)
	if err != nil {
		t.Fatalf("decode report: %v", err)
	}

	if decoded.Status != GateReportStatusPassed {
		t.Fatalf("Status = %q, want passed", decoded.Status)
	}

	if decoded.Base.Ref != "main" || decoded.Head.Ref != "dev" {
		t.Fatalf("decoded refs = %#v %#v, want main/dev", decoded.Base, decoded.Head)
	}
}

func TestGateReportWriterRejectsDirectoryTarget(t *testing.T) {
	projectDir := t.TempDir()
	path := filepath.Join(projectDir, ".chc", "user", "reports", "gates", "dev-main.json")

	err := os.MkdirAll(path, 0o755)
	if err != nil {
		t.Fatalf("create report directory target: %v", err)
	}

	report := GateReport{
		Artifacts: GateReportArtifacts{JSONPath: path},
	}
	_, err = NewGateReportWriter().WriteJSON(context.Background(), report)
	if err == nil {
		t.Fatal("WriteJSON returned nil error")
	}

	if !strings.Contains(err.Error(), "directory") {
		t.Fatalf("error = %q, want directory context", err.Error())
	}
}

func TestNewFailedGateReportRecordsRefsAndFailures(t *testing.T) {
	report := NewFailedGateReport(
		time.Date(2026, 6, 1, 12, 30, 45, 0, time.UTC),
		" main ",
		" dev ",
		GateReportArtifacts{},
		[]string{"resolve base ref"},
	)

	if report.Status != GateReportStatusFailed {
		t.Fatalf("Status = %q, want failed", report.Status)
	}

	if report.Base.Ref != "main" || report.Head.Ref != "dev" {
		t.Fatalf("refs = %#v %#v, want trimmed refs", report.Base, report.Head)
	}

	if len(report.Failures) != 1 {
		t.Fatalf("Failures = %#v, want one failure", report.Failures)
	}
}
