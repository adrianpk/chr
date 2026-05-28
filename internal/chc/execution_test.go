package chc

import (
	"path/filepath"
	"testing"
	"time"
)

func TestValidateExecutionRecordAcceptsValidRecord(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	record := ExecutionRecord{
		ID:        "exec-1",
		ProjectID: "project-1",
		TicketKey: "CHC-1",
		Status:    ExecutionStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := ValidateExecutionRecord(record)
	if err != nil {
		t.Fatalf("ValidateExecutionRecord returned error: %v", err)
	}
}

func TestValidateExecutionRecordRejectsMissingRequiredFields(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	for _, test := range missingRequiredExecutionRecordTests(now) {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateExecutionRecord(test.record)
			if err == nil {
				t.Fatal("ValidateExecutionRecord returned nil error")
			}
		})
	}
}

func missingRequiredExecutionRecordTests(now time.Time) []struct {
	name   string
	record ExecutionRecord
} {
	return []struct {
		name   string
		record ExecutionRecord
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
	record ExecutionRecord
} {
	return struct {
		name   string
		record ExecutionRecord
	}{
		name: "id",
		record: ExecutionRecord{
			ProjectID: "project-1",
			TicketKey: "CHC-1",
			Status:    ExecutionStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func missingExecutionProjectIDTest(now time.Time) struct {
	name   string
	record ExecutionRecord
} {
	return struct {
		name   string
		record ExecutionRecord
	}{
		name: "project id",
		record: ExecutionRecord{
			ID:        "exec-1",
			TicketKey: "CHC-1",
			Status:    ExecutionStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func missingExecutionTicketKeyTest(now time.Time) struct {
	name   string
	record ExecutionRecord
} {
	return struct {
		name   string
		record ExecutionRecord
	}{
		name: "ticket key",
		record: ExecutionRecord{
			ID:        "exec-1",
			ProjectID: "project-1",
			Status:    ExecutionStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func missingExecutionCreatedAtTest(now time.Time) struct {
	name   string
	record ExecutionRecord
} {
	return struct {
		name   string
		record ExecutionRecord
	}{
		name: "created at",
		record: ExecutionRecord{
			ID:        "exec-1",
			ProjectID: "project-1",
			TicketKey: "CHC-1",
			Status:    ExecutionStatusPending,
			UpdatedAt: now,
		},
	}
}

func missingExecutionUpdatedAtTest(now time.Time) struct {
	name   string
	record ExecutionRecord
} {
	return struct {
		name   string
		record ExecutionRecord
	}{
		name: "updated at",
		record: ExecutionRecord{
			ID:        "exec-1",
			ProjectID: "project-1",
			TicketKey: "CHC-1",
			Status:    ExecutionStatusPending,
			CreatedAt: now,
		},
	}
}

func TestValidateExecutionRecordRejectsInvalidStatus(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	record := ExecutionRecord{
		ID:        "exec-1",
		ProjectID: "project-1",
		TicketKey: "CHC-1",
		Status:    ExecutionStatus("unknown"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := ValidateExecutionRecord(record)
	if err == nil {
		t.Fatal("ValidateExecutionRecord returned nil error")
	}
}

func TestValidateExecutionRecordRejectsFinishedBeforeStarted(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	startedAt := now
	finishedAt := now.Add(-time.Minute)

	record := ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHC-1",
		Status:     ExecutionStatusFailed,
		CreatedAt:  now,
		StartedAt:  &startedAt,
		FinishedAt: &finishedAt,
		UpdatedAt:  now,
	}

	err := ValidateExecutionRecord(record)
	if err == nil {
		t.Fatal("ValidateExecutionRecord returned nil error")
	}
}

func TestExecutionRecordFromRuntimePlanCopiesRuntimeReferences(t *testing.T) {
	dir := t.TempDir()
	plan := RuntimePlan{
		Image:             "chc-agent-runtime:local",
		DockerfilePath:    filepath.Join(dir, ".chc", "user", "runtime", "agent", "Dockerfile"),
		ContainerWorkdir:  "/workspace",
		RunArtifactDir:    filepath.Join(dir, ".chc", "user", "runtime", "runs", "exec-1"),
		WorkerContextPath: filepath.Join(dir, ".chc", "user", "runtime", "runs", "exec-1", "context.md"),
		ResultPath:        filepath.Join(dir, ".chc", "user", "runtime", "runs", "exec-1", "result.json"),
		LogPath:           filepath.Join(dir, ".chc", "user", "runtime", "runs", "exec-1", "runtime.log"),
		ValidationPath:    filepath.Join(dir, ".chc", "user", "runtime", "runs", "exec-1", "validation.log"),
		ValidationCommand: "make verify",
	}

	record := ExecutionRecordFromRuntimePlan(ExecutionRecord{ID: "exec-1"}, plan)

	if record.Runtime.Image != plan.Image {
		t.Fatalf("Runtime.Image = %q, want %q", record.Runtime.Image, plan.Image)
	}

	if record.Artifacts.RunDir != plan.RunArtifactDir {
		t.Fatalf("Artifacts.RunDir = %q, want %q", record.Artifacts.RunDir, plan.RunArtifactDir)
	}

	if record.Artifacts.ResultPath != plan.ResultPath {
		t.Fatalf("Artifacts.ResultPath = %q, want %q", record.Artifacts.ResultPath, plan.ResultPath)
	}

	if record.Artifacts.ValidationPath != plan.ValidationPath {
		t.Fatalf("Artifacts.ValidationPath = %q, want %q", record.Artifacts.ValidationPath, plan.ValidationPath)
	}
}

func TestExecutionStatusValid(t *testing.T) {
	validStatuses := []ExecutionStatus{
		ExecutionStatusPending,
		ExecutionStatusRunning,
		ExecutionStatusSucceeded,
		ExecutionStatusFailed,
		ExecutionStatusBlocked,
	}

	for _, status := range validStatuses {
		if !status.Valid() {
			t.Fatalf("status %q is invalid, want valid", status)
		}
	}

	if ExecutionStatus("unknown").Valid() {
		t.Fatal("unknown status is valid, want invalid")
	}
}
