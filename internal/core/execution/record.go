package execution

import (
	"fmt"
	"time"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// ExecutionStatus defines a package boundary type.
type ExecutionStatus string

const (
	// ExecutionStatusPending defines a package boundary constant.
	ExecutionStatusPending ExecutionStatus = "pending"
	// ExecutionStatusRunning defines a package boundary constant.
	ExecutionStatusRunning ExecutionStatus = "running"
	// ExecutionStatusSucceeded defines a package boundary constant.
	ExecutionStatusSucceeded ExecutionStatus = "succeeded"
	// ExecutionStatusFailed defines a package boundary constant.
	ExecutionStatusFailed ExecutionStatus = "failed"
	// ExecutionStatusBlocked defines a package boundary constant.
	ExecutionStatusBlocked ExecutionStatus = "blocked"
)

// ExecutionRecord describes a package boundary value.
type ExecutionRecord struct {
	ID            string
	ProjectID     string
	TicketKey     string
	TicketPath    string
	Status        ExecutionStatus
	Runtime       ExecutionRuntimeRef
	Workspace     ExecutionWorkspaceRef
	Artifacts     ExecutionArtifacts
	FailureReason string
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	UpdatedAt     time.Time
}

// ExecutionRuntimeRef describes a package boundary value.
type ExecutionRuntimeRef struct {
	Image             string
	DockerfilePath    string
	ContainerWorkdir  string
	ValidationCommand string
}

// ExecutionWorkspaceRef describes where the agent work ran.
type ExecutionWorkspaceRef struct {
	Mode                  workspacecore.ExecutionMode
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	BaseRef               string
	BaseSHA               string
	TicketBranch          string
}

// ExecutionArtifacts describes a package boundary value.
type ExecutionArtifacts struct {
	ArtifactID     string
	RunDir         string
	ResultPath     string
	LogPath        string
	ValidationPath string
}

// ExecutionRecordFromRuntimePlan executes the corresponding package operation.
func ExecutionRecordFromRuntimePlan(record ExecutionRecord, plan harnessruntime.Plan) ExecutionRecord {
	record.Runtime = ExecutionRuntimeRef{
		Image:             plan.Image,
		DockerfilePath:    plan.DockerfilePath,
		ContainerWorkdir:  plan.ContainerWorkdir,
		ValidationCommand: plan.ValidationCommand,
	}
	record.Workspace = ExecutionWorkspaceRef{
		Mode:                  plan.WorkspaceMode,
		ProjectCheckoutDir:    plan.ProjectCheckoutDir,
		ExecutionWorkspaceDir: plan.ExecutionWorkspaceDir,
		BaseRef:               plan.WorkspaceBaseRef,
		BaseSHA:               plan.WorkspaceBaseSHA,
		TicketBranch:          plan.TicketBranch,
	}
	record.Artifacts = ExecutionArtifacts{
		ArtifactID:     plan.ArtifactID,
		RunDir:         plan.RunArtifactDir,
		ResultPath:     plan.ResultPath,
		LogPath:        plan.LogPath,
		ValidationPath: plan.ValidationPath,
	}

	return record
}

// ValidateExecutionRecord validates the corresponding package boundary value.
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

// Valid reports whether the value is one of the supported constants.
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
