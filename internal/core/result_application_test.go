package core

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestNewResultApplicationMapsSucceededExecutionToReviewing(t *testing.T) {
	record := testResultApplicationExecutionRecord(t.TempDir())
	record.Status = ExecutionStatusSucceeded

	result := testResultApplicationResult(record)

	application, err := NewResultApplication(ResultApplicationOptions{
		Execution:     record,
		Result:        result,
		CommentAuthor: "chr",
	})
	if err != nil {
		t.Fatalf("NewResultApplication returned error: %v", err)
	}

	if application.TicketStatus != TicketResultStatusReviewing {
		t.Fatalf("TicketStatus = %q, want reviewing", application.TicketStatus)
	}

	if application.Comment.Author != "chr" {
		t.Fatalf("Comment.Author = %q, want chr", application.Comment.Author)
	}

	for _, want := range []string{
		"Execution `exec-1` finished with status `succeeded`.",
		"Summary: ready for review",
	} {
		if !strings.Contains(application.Comment.Body, want) {
			t.Fatalf("Comment.Body does not contain %q:\n%s", want, application.Comment.Body)
		}
	}

	if application.Evidence.ValidationCommand != "make verify" {
		t.Fatalf("Evidence.ValidationCommand = %q, want make verify", application.Evidence.ValidationCommand)
	}

	if application.BlockedReason.Reason != "" {
		t.Fatalf("BlockedReason.Reason = %q, want empty", application.BlockedReason.Reason)
	}
}

func TestNewResultApplicationMapsFailedExecutionToBlocked(t *testing.T) {
	record := testResultApplicationExecutionRecord(t.TempDir())
	record.Status = ExecutionStatusFailed
	record.FailureReason = "harness: docker unavailable"

	result := testResultApplicationResult(record)
	result.Status = ExecutionResultStatusFailed
	result.Summary = "harness failed"

	application, err := NewResultApplication(ResultApplicationOptions{
		Execution: record,
		Result:    result,
	})
	if err != nil {
		t.Fatalf("NewResultApplication returned error: %v", err)
	}

	if application.TicketStatus != TicketResultStatusBlocked {
		t.Fatalf("TicketStatus = %q, want blocked", application.TicketStatus)
	}

	if application.BlockedReason.Kind != ExecutionFailureHarness {
		t.Fatalf("BlockedReason.Kind = %q, want harness", application.BlockedReason.Kind)
	}

	if application.BlockedReason.LogPath != record.Artifacts.LogPath {
		t.Fatalf("BlockedReason.LogPath = %q, want log path", application.BlockedReason.LogPath)
	}
}

func TestNewResultApplicationBlocksFailedEvidenceReadiness(t *testing.T) {
	record := testResultApplicationExecutionRecord(t.TempDir())
	result := testResultApplicationResult(record)
	readiness := EvidenceReadinessResult{
		Allowed: false,
		Diagnostics: []EvidenceReadinessDiagnostic{
			{
				Code:   EvidenceStrategyDiagnosticMissing,
				Reason: "ticket evidence strategy checklist is missing",
			},
		},
	}

	application, err := NewResultApplication(ResultApplicationOptions{
		Execution:         record,
		Result:            result,
		EvidenceReadiness: &readiness,
	})
	if err != nil {
		t.Fatalf("NewResultApplication returned error: %v", err)
	}

	if application.TicketStatus != TicketResultStatusBlocked {
		t.Fatalf("TicketStatus = %q, want blocked", application.TicketStatus)
	}

	if application.BlockedReason.Kind != ExecutionFailureValidation {
		t.Fatalf("BlockedReason.Kind = %q, want validation", application.BlockedReason.Kind)
	}

	if !strings.Contains(application.BlockedReason.Reason, "evidence_strategy_missing") {
		t.Fatalf("BlockedReason.Reason = %q, want evidence diagnostic", application.BlockedReason.Reason)
	}
}

func TestNewResultApplicationBlocksEvidenceReadinessWithWorkspaceMetadata(t *testing.T) {
	projectDir := t.TempDir()
	worktreeDir := t.TempDir()
	record := testResultApplicationExecutionRecord(projectDir)
	record.Workspace = ExecutionWorkspaceRef{
		Mode:                  workspacecore.ExecutionWorktree,
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: worktreeDir,
		BaseRef:               "origin/dev",
		BaseSHA:               "abc123",
		TicketBranch:          "chr/CHR-1",
	}
	result := testResultApplicationResult(record)
	readiness := EvidenceReadinessResult{
		Allowed: false,
		Diagnostics: []EvidenceReadinessDiagnostic{
			{
				Code:   EvidenceStrategyDiagnosticValidationMissing,
				Reason: "validation evidence is missing",
			},
		},
	}

	application, err := NewResultApplication(ResultApplicationOptions{
		Execution:         record,
		Result:            result,
		EvidenceReadiness: &readiness,
	})
	if err != nil {
		t.Fatalf("NewResultApplication returned error: %v", err)
	}

	if application.TicketStatus != TicketResultStatusBlocked {
		t.Fatalf("TicketStatus = %q, want blocked", application.TicketStatus)
	}

	if application.BlockedReason.Kind != ExecutionFailureValidation {
		t.Fatalf("BlockedReason.Kind = %q, want validation", application.BlockedReason.Kind)
	}

	if !strings.Contains(application.BlockedReason.Reason, "validation_missing") {
		t.Fatalf("BlockedReason.Reason = %q, want validation evidence diagnostic", application.BlockedReason.Reason)
	}

	if application.Evidence.ValidationPath != record.Artifacts.ValidationPath {
		t.Fatalf("Evidence.ValidationPath = %q, want %q", application.Evidence.ValidationPath, record.Artifacts.ValidationPath)
	}

	if application.BlockedReason.ResultPath != record.Artifacts.ResultPath {
		t.Fatalf("BlockedReason.ResultPath = %q, want %q", application.BlockedReason.ResultPath, record.Artifacts.ResultPath)
	}
}

func TestNewResultApplicationPreservesChangedFiles(t *testing.T) {
	record := testResultApplicationExecutionRecord(t.TempDir())
	result := testResultApplicationResult(record)

	application, err := NewResultApplication(ResultApplicationOptions{
		Execution: record,
		Result:    result,
		ChangedFiles: []ResultApplicationChangedFile{
			{Path: "internal/core/result_application.go", Summary: "add model"},
		},
	})
	if err != nil {
		t.Fatalf("NewResultApplication returned error: %v", err)
	}

	if len(application.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles len = %d, want 1", len(application.ChangedFiles))
	}

	if application.ChangedFiles[0].Path != "internal/core/result_application.go" {
		t.Fatalf("ChangedFiles[0].Path = %q, want model path", application.ChangedFiles[0].Path)
	}
}

func TestNewResultApplicationMapsResultChangedFiles(t *testing.T) {
	record := testResultApplicationExecutionRecord(t.TempDir())
	result := testResultApplicationResult(record)
	result.ChangedFiles = []ExecutionResultChangedFile{
		{Path: "internal/core/execution_result.go", Summary: "add changed files"},
	}

	application, err := NewResultApplication(ResultApplicationOptions{
		Execution: record,
		Result:    result,
	})
	if err != nil {
		t.Fatalf("NewResultApplication returned error: %v", err)
	}

	if len(application.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles len = %d, want 1", len(application.ChangedFiles))
	}

	if application.ChangedFiles[0].Summary != "add changed files" {
		t.Fatalf("ChangedFiles[0].Summary = %q, want summary", application.ChangedFiles[0].Summary)
	}
}

func TestNewResultApplicationRejectsMismatchedExecutionID(t *testing.T) {
	record := testResultApplicationExecutionRecord(t.TempDir())
	result := testResultApplicationResult(record)
	result.ExecutionID = "exec-other"

	_, err := NewResultApplication(ResultApplicationOptions{
		Execution: record,
		Result:    result,
	})
	if err == nil {
		t.Fatal("NewResultApplication returned nil error")
	}
}

func testResultApplicationExecutionRecord(dir string) ExecutionRecord {
	runDir := filepath.Join(dir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	return ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(dir, ".chr", "default", "tickets", "CHR-1.md"),
		Status:     ExecutionStatusSucceeded,
		Runtime: ExecutionRuntimeRef{
			Image:             "chr-agent-runtime:local",
			DockerfilePath:    filepath.Join(dir, ".chr", "default", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func testResultApplicationResult(record ExecutionRecord) ExecutionResult {
	return ExecutionResult{
		ExecutionID:     record.ID,
		TicketKey:       record.TicketKey,
		Status:          ExecutionResultStatusSucceeded,
		Summary:         "ready for review",
		EvidenceSummary: testExecutionResultEvidenceSummary(),
	}
}
