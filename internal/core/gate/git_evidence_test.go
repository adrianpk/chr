package gate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGateGitEvidenceServiceCollectsRangeEvidence(t *testing.T) {
	projectDir := newGateGitEvidenceRepo(t)
	writeGateGitEvidenceFile(t, projectDir, "README.md", "base\n")
	runGateGitEvidenceGit(t, projectDir, "add", ".")
	runGateGitEvidenceGit(t, projectDir, "commit", "-m", "base commit")

	baseSHA := gateGitEvidenceOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateGitEvidenceFile(t, projectDir, "README.md", "changed\n")
	writeGateGitEvidenceFile(t, projectDir, "internal/core/gate.go", "package core\n")
	runGateGitEvidenceGit(t, projectDir, "add", ".")
	runGateGitEvidenceGit(t, projectDir, "commit", "-m", "gate evidence")

	headSHA := gateGitEvidenceOutput(t, projectDir, "rev-parse", "HEAD")

	evidence, err := NewGitEvidenceService().Collect(t.Context(), GitEvidenceRequest{
		ProjectDir: projectDir,
		BaseRef:    baseSHA,
		HeadRef:    "HEAD",
	})
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if evidence.Base.SHA != baseSHA {
		t.Fatalf("Base.SHA = %q, want %q", evidence.Base.SHA, baseSHA)
	}

	if evidence.Head.SHA != headSHA {
		t.Fatalf("Head.SHA = %q, want %q", evidence.Head.SHA, headSHA)
	}

	if evidence.MergeBase != baseSHA {
		t.Fatalf("MergeBase = %q, want %q", evidence.MergeBase, baseSHA)
	}

	assertGateGitCommits(t, evidence.Commits, []string{"gate evidence"})
	assertGateGitFiles(t, evidence.Files, []GitChangedFile{
		{Status: "M", Path: "README.md"},
		{Status: "A", Path: "internal/core/gate.go"},
	})

	if evidence.DiffStat == "" {
		t.Fatal("DiffStat is empty")
	}
}

func TestGateGitEvidenceServiceRejectsInvalidRefs(t *testing.T) {
	projectDir := newGateGitEvidenceRepo(t)
	writeGateGitEvidenceFile(t, projectDir, "README.md", "base\n")
	runGateGitEvidenceGit(t, projectDir, "add", ".")
	runGateGitEvidenceGit(t, projectDir, "commit", "-m", "base commit")

	_, err := NewGitEvidenceService().Collect(t.Context(), GitEvidenceRequest{
		ProjectDir: projectDir,
		BaseRef:    "missing-ref",
		HeadRef:    "HEAD",
	})
	if err == nil {
		t.Fatal("Collect returned nil error")
	}

	if !strings.Contains(err.Error(), "resolve base ref") {
		t.Fatalf("error = %q, want base ref context", err.Error())
	}
}

func TestParseGateGitChangedFilesHandlesRenames(t *testing.T) {
	files := parseGitChangedFiles("R100\told/name.go\tnew/name.go\n")

	assertGateGitFiles(t, files, []GitChangedFile{
		{Status: "R100", Old: "old/name.go", Path: "new/name.go"},
	})
}

func TestGateGitEvidenceServiceRejectsMissingProjectDir(t *testing.T) {
	_, err := NewGitEvidenceService().Collect(context.Background(), GitEvidenceRequest{
		BaseRef: "main",
		HeadRef: "dev",
	})
	if err == nil {
		t.Fatal("Collect returned nil error")
	}
}

func newGateGitEvidenceRepo(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()
	runGateGitEvidenceGit(t, projectDir, "init")

	return projectDir
}

func runGateGitEvidenceGit(t *testing.T, projectDir string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{
		"-C", projectDir,
		"-c", "user.name=chr Test",
		"-c", "user.email=chr@example.test",
	}, args...)
	cmd := exec.Command("git", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func gateGitEvidenceOutput(t *testing.T, projectDir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", projectDir}, args...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}

	return strings.TrimSpace(string(output))
}

func writeGateGitEvidenceFile(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func assertGateGitCommits(t *testing.T, commits []GitCommit, wantSubjects []string) {
	t.Helper()

	if len(commits) != len(wantSubjects) {
		t.Fatalf("len(commits) = %d, want %d: %#v", len(commits), len(wantSubjects), commits)
	}

	for index := range wantSubjects {
		if commits[index].Subject != wantSubjects[index] {
			t.Fatalf("commits[%d].Subject = %q, want %q", index, commits[index].Subject, wantSubjects[index])
		}
	}
}

func assertGateGitFiles(t *testing.T, files, want []GitChangedFile) {
	t.Helper()

	if len(files) != len(want) {
		t.Fatalf("len(files) = %d, want %d: %#v", len(files), len(want), files)
	}

	for index := range want {
		if files[index] != want[index] {
			t.Fatalf("files[%d] = %#v, want %#v", index, files[index], want[index])
		}
	}
}
