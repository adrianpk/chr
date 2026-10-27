package execution_test

import (
	"path/filepath"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestValidateExecutionRecordAcceptsValidRecord(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	record := executioncore.Record{
		ID:        "exec-1",
		ProjectID: "project-1",
		TicketKey: "CHR-1",
		Status:    executioncore.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := executioncore.ValidateRecord(record)
	if err != nil {
		t.Fatalf("ValidateExecutionRecord returned error: %v", err)
	}
}

func TestValidateExecutionRecordRejectsMissingRequiredFields(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	for _, test := range missingRequiredExecutionRecordTests(now) {
		t.Run(test.name, func(t *testing.T) {
			err := executioncore.ValidateRecord(test.record)
			if err == nil {
				t.Fatal("ValidateExecutionRecord returned nil error")
			}
		})
	}
}

func missingRequiredExecutionRecordTests(now time.Time) []struct {
	name   string
	record executioncore.Record
} {
	return []struct {
		name   string
		record executioncore.Record
	}{
		missingExecutionIDTest(now),
		missingExecutionProjectIDTest(now),
		missingExecutionTicketKeyTest(now),
		missingExecutionCreatedAtTest(now),
		missingExecutionUpdatedAtTest(now),
	}
}

func missingExecutionIDTest(now time.Time) struct {
	name   string
	record executioncore.Record
} {
	return struct {
		name   string
		record executioncore.Record
	}{
		name: "id",
		record: executioncore.Record{
			ProjectID: "project-1",
			TicketKey: "CHR-1",
			Status:    executioncore.StatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func missingExecutionProjectIDTest(now time.Time) struct {
	name   string
	record executioncore.Record
} {
	return struct {
		name   string
		record executioncore.Record
	}{
		name: "project id",
		record: executioncore.Record{
			ID:        "exec-1",
			TicketKey: "CHR-1",
			Status:    executioncore.StatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func missingExecutionTicketKeyTest(now time.Time) struct {
	name   string
	record executioncore.Record
} {
	return struct {
		name   string
		record executioncore.Record
	}{
		name: "ticket key",
		record: executioncore.Record{
			ID:        "exec-1",
			ProjectID: "project-1",
			Status:    executioncore.StatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func missingExecutionCreatedAtTest(now time.Time) struct {
	name   string
	record executioncore.Record
} {
	return struct {
		name   string
		record executioncore.Record
	}{
		name: "created at",
		record: executioncore.Record{
			ID:        "exec-1",
			ProjectID: "project-1",
			TicketKey: "CHR-1",
			Status:    executioncore.StatusPending,
			UpdatedAt: now,
		},
	}
}

func missingExecutionUpdatedAtTest(now time.Time) struct {
	name   string
	record executioncore.Record
} {
	return struct {
		name   string
		record executioncore.Record
	}{
		name: "updated at",
		record: executioncore.Record{
			ID:        "exec-1",
			ProjectID: "project-1",
			TicketKey: "CHR-1",
			Status:    executioncore.StatusPending,
			CreatedAt: now,
		},
	}
}

func TestValidateExecutionRecordRejectsInvalidStatus(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	record := executioncore.Record{
		ID:        "exec-1",
		ProjectID: "project-1",
		TicketKey: "CHR-1",
		Status:    executioncore.Status("unknown"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := executioncore.ValidateRecord(record)
	if err == nil {
		t.Fatal("ValidateExecutionRecord returned nil error")
	}
}

func TestValidateExecutionRecordRejectsFinishedBeforeStarted(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	startedAt := now
	finishedAt := now.Add(-time.Minute)

	record := executioncore.Record{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHR-1",
		Status:     executioncore.StatusFailed,
		CreatedAt:  now,
		StartedAt:  &startedAt,
		FinishedAt: &finishedAt,
		UpdatedAt:  now,
	}

	err := executioncore.ValidateRecord(record)
	if err == nil {
		t.Fatal("ValidateExecutionRecord returned nil error")
	}
}

func TestExecutionRecordFromRuntimePlanCopiesRuntimeReferences(t *testing.T) {
	dir := t.TempDir()
	plan := harnessruntime.Plan{
		ArtifactID:            "00001-ticket-run",
		Image:                 "chr-agent-runtime:local",
		DockerfilePath:        filepath.Join(dir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		ContainerWorkdir:      "/workspace",
		WorkspaceMode:         workspacecore.ExecutionWorktree,
		ProjectCheckoutDir:    dir,
		ExecutionWorkspaceDir: filepath.Join(dir, ".chr", "default", "worktrees", "CHR-1"),
		WorkspaceBaseRef:      "origin/dev",
		WorkspaceBaseSHA:      "abc123",
		TicketBranch:          "chr/CHR-1",
		RunArtifactDir:        filepath.Join(dir, ".chr", "default", "runtime", "runs", "exec-1"),
		WorkerContextPath:     filepath.Join(dir, ".chr", "default", "runtime", "runs", "exec-1", "context.md"),
		ResultPath:            filepath.Join(dir, ".chr", "default", "runtime", "runs", "exec-1", "result.json"),
		LogPath:               filepath.Join(dir, ".chr", "default", "runtime", "runs", "exec-1", "runtime.log"),
		ValidationPath:        filepath.Join(dir, ".chr", "default", "runtime", "runs", "exec-1", "validation.log"),
		ValidationCommand:     "make verify",
	}

	record := executioncore.RecordFromRuntimePlan(executioncore.Record{ID: "exec-1"}, plan)

	if record.Runtime.Image != plan.Image {
		t.Fatalf("Runtime.Image = %q, want %q", record.Runtime.Image, plan.Image)
	}

	if record.Artifacts.RunDir != plan.RunArtifactDir {
		t.Fatalf("Artifacts.RunDir = %q, want %q", record.Artifacts.RunDir, plan.RunArtifactDir)
	}

	if record.Artifacts.ArtifactID != plan.ArtifactID {
		t.Fatalf("Artifacts.ArtifactID = %q, want %q", record.Artifacts.ArtifactID, plan.ArtifactID)
	}

	if record.Artifacts.ResultPath != plan.ResultPath {
		t.Fatalf("Artifacts.ResultPath = %q, want %q", record.Artifacts.ResultPath, plan.ResultPath)
	}

	if record.Artifacts.ValidationPath != plan.ValidationPath {
		t.Fatalf("Artifacts.ValidationPath = %q, want %q", record.Artifacts.ValidationPath, plan.ValidationPath)
	}

	if record.Workspace.Mode != workspacecore.ExecutionWorktree {
		t.Fatalf("Workspace.Mode = %q, want worktree", record.Workspace.Mode)
	}

	if record.Workspace.ExecutionWorkspaceDir != plan.ExecutionWorkspaceDir {
		t.Fatalf("Workspace.ExecutionWorkspaceDir = %q, want %q", record.Workspace.ExecutionWorkspaceDir, plan.ExecutionWorkspaceDir)
	}

	if record.Workspace.TicketBranch != "chr/CHR-1" {
		t.Fatalf("Workspace.TicketBranch = %q, want chr/CHR-1", record.Workspace.TicketBranch)
	}
}

func TestExecutionStatusValid(t *testing.T) {
	validStatuses := []executioncore.Status{
		executioncore.StatusPending,
		executioncore.StatusRunning,
		executioncore.StatusSucceeded,
		executioncore.StatusFailed,
		executioncore.StatusBlocked,
	}

	for _, status := range validStatuses {
		if !status.Valid() {
			t.Fatalf("status %q is invalid, want valid", status)
		}
	}

	if executioncore.Status("unknown").Valid() {
		t.Fatal("unknown status is valid, want invalid")
	}
}
