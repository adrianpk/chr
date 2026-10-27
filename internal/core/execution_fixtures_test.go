package core

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func newTestExecutionRepository(t *testing.T) executioncore.Repository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := executioncore.NewRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("close sqlite after repository error: %v", closeErr)
		}

		t.Fatalf("NewExecutionRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	return repository
}

func testExecutionRecord(now time.Time) executioncore.Record {
	return executioncore.Record{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHR-1",
		TicketPath: "/tmp/project/.chr/default/tickets/CHR-1.md",
		Status:     executioncore.StatusPending,
		Runtime: executioncore.RuntimeRef{
			Image:             "chr-agent-runtime:local",
			DockerfilePath:    "/tmp/project/.chr/default/runtime/agent/Dockerfile",
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Workspace: executioncore.WorkspaceRef{
			Mode:                  workspacecore.ExecutionWorktree,
			ProjectCheckoutDir:    "/tmp/project",
			ExecutionWorkspaceDir: "/tmp/project/.chr/default/worktrees/CHR-1",
			BaseRef:               "origin/dev",
			BaseSHA:               "abc123",
			TicketBranch:          "chr/CHR-1",
		},
		Artifacts: executioncore.Artifacts{
			ArtifactID: "00001-parser-validation",
			RunDir:     "/tmp/project/.chr/default/runtime/runs/exec-1",
			ResultPath: "/tmp/project/.chr/default/runtime/runs/exec-1/result.json",
			LogPath:    "/tmp/project/.chr/default/runtime/runs/exec-1/runtime.log",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func optionalTimeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}

	return *value
}

func testExecutionResultEvidenceSummary() executioncore.ResultEvidenceSummary {
	return executioncore.ResultEvidenceSummary{
		Strategy:   "planned",
		Produced:   "evidence produced",
		Validation: "validation passed",
	}
}
