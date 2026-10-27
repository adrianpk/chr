package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestExecutionArtifactDriftServiceReportsPresentArtifacts(t *testing.T) {
	repository := newTestExecutionRepository(t)
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	record := testExecutionArtifactRecord(t, now)
	writeExecutionArtifacts(t, record)

	err := repository.CreateExecution(context.Background(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	diagnostics, err := executioncore.NewArtifactDriftService(repository).ListDiagnostics(
		context.Background(),
		record.ProjectID,
	)
	if err != nil {
		t.Fatalf("ListDiagnostics returned error: %v", err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("len(diagnostics) = %d, want 1", len(diagnostics))
	}

	if diagnostics[0].Status != executioncore.ArtifactDriftStatusOK {
		t.Fatalf("Status = %q, want ok", diagnostics[0].Status)
	}

	for _, artifact := range diagnostics[0].Artifacts {
		if !artifact.Exists {
			t.Fatalf("artifact %q exists = false, want true", artifact.Name)
		}
	}
}

func TestExecutionArtifactDriftServiceReportsMissingArtifacts(t *testing.T) {
	repository := newTestExecutionRepository(t)
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	record := testExecutionArtifactRecord(t, now)

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	err = repository.CreateExecution(context.Background(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	diagnostics, err := executioncore.NewArtifactDriftService(repository).ListDiagnostics(
		context.Background(),
		record.ProjectID,
	)
	if err != nil {
		t.Fatalf("ListDiagnostics returned error: %v", err)
	}

	if diagnostics[0].Status != executioncore.ArtifactDriftStatusMissing {
		t.Fatalf("Status = %q, want missing artifacts", diagnostics[0].Status)
	}

	assertArtifactExists(t, diagnostics[0], "run_dir", true)
	assertArtifactExists(t, diagnostics[0], "result", false)
	assertArtifactExists(t, diagnostics[0], "log", false)
	assertArtifactExists(t, diagnostics[0], "validation", false)
	assertArtifactExists(t, diagnostics[0], "report", false)
}

func testExecutionArtifactRecord(t *testing.T, now time.Time) executioncore.Record {
	t.Helper()

	runDir := filepath.Join(t.TempDir(), ".chr", "default", "runtime", "runs", "exec-1")

	record := testExecutionRecord(now)
	record.Artifacts = executioncore.Artifacts{
		RunDir:         runDir,
		ResultPath:     filepath.Join(runDir, "result.json"),
		LogPath:        filepath.Join(runDir, "runtime.log"),
		ValidationPath: filepath.Join(runDir, "validation.log"),
	}

	return record
}

func writeExecutionArtifacts(t *testing.T, record executioncore.Record) {
	t.Helper()

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	for _, path := range []string{
		record.Artifacts.ResultPath,
		record.Artifacts.LogPath,
		record.Artifacts.ValidationPath,
		executioncore.ReportPath(record),
	} {
		err = os.WriteFile(path, []byte("artifact\n"), 0o644)
		if err != nil {
			t.Fatalf("write artifact %q: %v", path, err)
		}
	}
}

func assertArtifactExists(
	t *testing.T,
	diagnostic executioncore.ArtifactDriftDiagnostic,
	name string,
	want bool,
) {
	t.Helper()

	for _, artifact := range diagnostic.Artifacts {
		if artifact.Name == name {
			if artifact.Exists != want {
				t.Fatalf("artifact %q exists = %t, want %t", name, artifact.Exists, want)
			}

			return
		}
	}

	t.Fatalf("artifact %q not found", name)
}
