package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRuntimeArtifactCleanupPreviewReportsUnreferencedLeafDirs(t *testing.T) {
	repository := newTestExecutionRepository(t)
	projectDir := t.TempDir()
	artifactDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs")
	referencedDir := filepath.Join(artifactDir, "CHC-1", "exec-1")
	unreferencedDir := filepath.Join(artifactDir, "CHC-2", "exec-2")

	for _, dir := range []string{referencedDir, unreferencedDir} {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			t.Fatalf("create runtime artifact dir %q: %v", dir, err)
		}
	}

	record := testExecutionRecord(time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC))
	record.Artifacts.RunDir = referencedDir

	err := repository.CreateExecution(context.Background(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	preview, err := NewRuntimeArtifactCleanupPreviewService(repository).Preview(
		context.Background(),
		projectDir,
		record.ProjectID,
	)
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}

	if preview.ArtifactDir != artifactDir {
		t.Fatalf("ArtifactDir = %q, want %q", preview.ArtifactDir, artifactDir)
	}

	if len(preview.ReferencedRunDirs) != 1 {
		t.Fatalf("len(ReferencedRunDirs) = %d, want 1", len(preview.ReferencedRunDirs))
	}

	if preview.ReferencedRunDirs[0] != referencedDir {
		t.Fatalf("ReferencedRunDirs[0] = %q, want %q", preview.ReferencedRunDirs[0], referencedDir)
	}

	if len(preview.Candidates) != 1 {
		t.Fatalf("len(Candidates) = %d, want 1", len(preview.Candidates))
	}

	if preview.Candidates[0].Path != unreferencedDir {
		t.Fatalf("Candidates[0].Path = %q, want %q", preview.Candidates[0].Path, unreferencedDir)
	}
}

func TestRuntimeArtifactCleanupPreviewMissingArtifactDirHasNoCandidates(t *testing.T) {
	repository := newTestExecutionRepository(t)
	projectDir := t.TempDir()

	preview, err := NewRuntimeArtifactCleanupPreviewService(repository).Preview(
		context.Background(),
		projectDir,
		"project-1",
	)
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}

	if len(preview.Candidates) != 0 {
		t.Fatalf("len(Candidates) = %d, want 0", len(preview.Candidates))
	}
}
