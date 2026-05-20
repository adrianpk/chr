package chc

import (
	"fmt"
	"time"
)

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSucceeded ExecutionStatus = "succeeded"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusBlocked   ExecutionStatus = "blocked"
)

type ExecutionRecord struct {
	ID            string
	ProjectID     string
	TicketKey     string
	TicketPath    string
	Status        ExecutionStatus
	Runtime       ExecutionRuntimeRef
	Artifacts     ExecutionArtifacts
	FailureReason string
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	UpdatedAt     time.Time
}

type ExecutionRuntimeRef struct {
	Image             string
	DockerfilePath    string
	ContainerWorkdir  string
	ValidationCommand string
}

type ExecutionArtifacts struct {
	RunDir         string
	ResultPath     string
	LogPath        string
	ValidationPath string
}

func ExecutionRecordFromRuntimePlan(record ExecutionRecord, plan RuntimePlan) ExecutionRecord {
	record.Runtime = ExecutionRuntimeRef{
		Image:             plan.Image,
		DockerfilePath:    plan.DockerfilePath,
		ContainerWorkdir:  plan.ContainerWorkdir,
		ValidationCommand: plan.ValidationCommand,
	}
	record.Artifacts = ExecutionArtifacts{
		RunDir:         plan.RunArtifactDir,
		ResultPath:     plan.ResultPath,
		LogPath:        plan.LogPath,
		ValidationPath: plan.ValidationPath,
	}

	return record
}

func ValidateExecutionRecord(record ExecutionRecord) error {
	if record.ID == "" {
		return fmt.Errorf("execution id is required")
	}

	if record.ProjectID == "" {
		return fmt.Errorf("execution project id is required")
	}

	if record.TicketKey == "" {
		return fmt.Errorf("execution ticket key is required")
	}

	if !record.Status.Valid() {
		return fmt.Errorf("execution status is invalid: %s", record.Status)
	}

	if record.CreatedAt.IsZero() {
		return fmt.Errorf("execution created time is required")
	}

	if record.UpdatedAt.IsZero() {
		return fmt.Errorf("execution updated time is required")
	}

	if record.StartedAt != nil && record.FinishedAt != nil && record.FinishedAt.Before(*record.StartedAt) {
		return fmt.Errorf("execution finished before it started")
	}

	return nil
}

func (s ExecutionStatus) Valid() bool {
	switch s {
	case ExecutionStatusPending,
		ExecutionStatusRunning,
		ExecutionStatusSucceeded,
		ExecutionStatusFailed,
		ExecutionStatusBlocked:
		return true
	default:
		return false
	}
}
