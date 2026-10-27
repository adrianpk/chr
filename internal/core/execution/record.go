package execution

import (
	"fmt"
	"time"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Status defines a package boundary type.
type Status string

const (
	// StatusPending defines a package boundary constant.
	StatusPending Status = "pending"
	// StatusRunning defines a package boundary constant.
	StatusRunning Status = "running"
	// StatusSucceeded defines a package boundary constant.
	StatusSucceeded Status = "succeeded"
	// StatusFailed defines a package boundary constant.
	StatusFailed Status = "failed"
	// StatusBlocked defines a package boundary constant.
	StatusBlocked Status = "blocked"
)

// Record describes a package boundary value.
type Record struct {
	ID            string
	ProjectID     string
	TicketKey     string
	TicketPath    string
	Status        Status
	Runtime       RuntimeRef
	Workspace     WorkspaceRef
	Artifacts     Artifacts
	FailureReason string
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	UpdatedAt     time.Time
}

// RuntimeRef describes a package boundary value.
type RuntimeRef struct {
	Image             string
	DockerfilePath    string
	ContainerWorkdir  string
	ValidationCommand string
}

// WorkspaceRef describes where the agent work ran.
type WorkspaceRef struct {
	Mode                  workspacecore.ExecutionMode
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	BaseRef               string
	BaseSHA               string
	TicketBranch          string
}

// Artifacts describes a package boundary value.
type Artifacts struct {
	ArtifactID     string
	RunDir         string
	ResultPath     string
	LogPath        string
	ValidationPath string
}

// RecordFromRuntimePlan executes the corresponding package operation.
func RecordFromRuntimePlan(record Record, plan harnessruntime.Plan) Record {
	record.Runtime = RuntimeRef{
		Image:             plan.Image,
		DockerfilePath:    plan.DockerfilePath,
		ContainerWorkdir:  plan.ContainerWorkdir,
		ValidationCommand: plan.ValidationCommand,
	}
	record.Workspace = WorkspaceRef{
		Mode:                  plan.WorkspaceMode,
		ProjectCheckoutDir:    plan.ProjectCheckoutDir,
		ExecutionWorkspaceDir: plan.ExecutionWorkspaceDir,
		BaseRef:               plan.WorkspaceBaseRef,
		BaseSHA:               plan.WorkspaceBaseSHA,
		TicketBranch:          plan.TicketBranch,
	}
	record.Artifacts = Artifacts{
		ArtifactID:     plan.ArtifactID,
		RunDir:         plan.RunArtifactDir,
		ResultPath:     plan.ResultPath,
		LogPath:        plan.LogPath,
		ValidationPath: plan.ValidationPath,
	}

	return record
}

// ValidateRecord validates the corresponding package boundary value.
func ValidateRecord(record Record) error {
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
func (s Status) Valid() bool {
	switch s {
	case StatusPending,
		StatusRunning,
		StatusSucceeded,
		StatusFailed,
		StatusBlocked:
		return true
	default:
		return false
	}
}
