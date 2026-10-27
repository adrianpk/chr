package execution

import (
	"database/sql"

	"forge.adrianpk.com/chr/internal/core/database"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// executionRow pairs the stored column order with domain values and raw scan conversions.
type executionRow struct {
	record                Record
	status, workspaceMode string
	createdAt, updatedAt  string
	startedAt, finishedAt sql.NullString
}

const executionColumns = `id, project_id, ticket_key, ticket_path, status,
    runtime_image, runtime_dockerfile_path, runtime_container_workdir,
    runtime_validation_command, workspace_mode, workspace_project_checkout_dir,
    workspace_execution_dir, workspace_base_ref, workspace_base_sha,
    workspace_ticket_branch, artifact_id, artifact_run_dir, artifact_result_path,
    artifact_log_path, artifact_validation_path, failure_reason,
    created_at, started_at, finished_at, updated_at`

func (row executionRow) values() []any {
	record := row.record

	return []any{
		record.ID,
		record.ProjectID,
		record.TicketKey,
		record.TicketPath,
		string(record.Status),
		record.Runtime.Image,
		record.Runtime.DockerfilePath,
		record.Runtime.ContainerWorkdir,
		record.Runtime.ValidationCommand,
		string(record.Workspace.Mode),
		record.Workspace.ProjectCheckoutDir,
		record.Workspace.ExecutionWorkspaceDir,
		record.Workspace.BaseRef,
		record.Workspace.BaseSHA,
		record.Workspace.TicketBranch,
		record.Artifacts.ArtifactID,
		record.Artifacts.RunDir,
		record.Artifacts.ResultPath,
		record.Artifacts.LogPath,
		record.Artifacts.ValidationPath,
		record.FailureReason,
		database.FormatTime(record.CreatedAt),
		formatOptionalExecutionTime(record.StartedAt),
		formatOptionalExecutionTime(record.FinishedAt),
		database.FormatTime(record.UpdatedAt),
	}
}

func (row *executionRow) scanTargets() []any {
	record := &row.record

	return []any{
		&record.ID,
		&record.ProjectID,
		&record.TicketKey,
		&record.TicketPath,
		&row.status,
		&record.Runtime.Image,
		&record.Runtime.DockerfilePath,
		&record.Runtime.ContainerWorkdir,
		&record.Runtime.ValidationCommand,
		&row.workspaceMode,
		&record.Workspace.ProjectCheckoutDir,
		&record.Workspace.ExecutionWorkspaceDir,
		&record.Workspace.BaseRef,
		&record.Workspace.BaseSHA,
		&record.Workspace.TicketBranch,
		&record.Artifacts.ArtifactID,
		&record.Artifacts.RunDir,
		&record.Artifacts.ResultPath,
		&record.Artifacts.LogPath,
		&record.Artifacts.ValidationPath,
		&record.FailureReason,
		&row.createdAt,
		&row.startedAt,
		&row.finishedAt,
		&row.updatedAt,
	}
}

func (row executionRow) decoded() (Record, error) {
	times, err := parseExecutionRecordTimes(row.createdAt, row.updatedAt, row.startedAt, row.finishedAt)
	if err != nil {
		return Record{}, err
	}

	record := row.record
	record.Status = Status(row.status)
	record.Workspace.Mode = workspacecore.ExecutionMode(row.workspaceMode)
	record.CreatedAt, record.UpdatedAt = times.CreatedAt, times.UpdatedAt
	record.StartedAt, record.FinishedAt = times.StartedAt, times.FinishedAt

	return record, nil
}

func insertExecutionSQL() string {
	return `INSERT INTO execution_records (` + executionColumns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}
func selectExecutionSQL() string { return `SELECT ` + executionColumns + ` FROM execution_records` }
