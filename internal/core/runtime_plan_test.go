package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanRuntimeUsesDefaultConfig(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.ProjectDir != dir {
		t.Fatalf("ProjectDir = %q, want %q", plan.ProjectDir, dir)
	}

	if plan.WorkspaceDir != filepath.Join(dir, ".chc") {
		t.Fatalf("WorkspaceDir = %q, want .chc path", plan.WorkspaceDir)
	}

	if plan.DockerfilePath != filepath.Join(dir, ".chc", "user", "runtime", "agent", "Dockerfile") {
		t.Fatalf("DockerfilePath = %q, want default Dockerfile path", plan.DockerfilePath)
	}

	if plan.Image != DefaultRuntimeImage {
		t.Fatalf("Image = %q, want %q", plan.Image, DefaultRuntimeImage)
	}

	if plan.ContainerWorkdir != DefaultRuntimeWorkdir {
		t.Fatalf("ContainerWorkdir = %q, want %q", plan.ContainerWorkdir, DefaultRuntimeWorkdir)
	}

	if plan.ArtifactDir != filepath.Join(dir, ".chc", "user", "runtime", "runs") {
		t.Fatalf("ArtifactDir = %q, want default artifact dir", plan.ArtifactDir)
	}

	if plan.ResultPath != filepath.Join(dir, ".chc", "user", "runtime", "runs", "result.json") {
		t.Fatalf("ResultPath = %q, want default result path", plan.ResultPath)
	}
}

func TestPlanRuntimeUsesDefaultArtifactPaths(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.WorkerContextPath != filepath.Join(dir, ".chc", "user", "runtime", "runs", "context.md") {
		t.Fatalf("WorkerContextPath = %q, want default context path", plan.WorkerContextPath)
	}

	if plan.ValidationPath != filepath.Join(dir, ".chc", "user", "runtime", "runs", "validation.log") {
		t.Fatalf("ValidationPath = %q, want default validation path", plan.ValidationPath)
	}
}

func TestPlanRuntimeUsesCustomConfig(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		Config: RuntimeConfig{
			Image:             "example/runtime:dev",
			DockerfilePath:    "ops/runtime.Dockerfile",
			Workdir:           "/repo",
			ArtifactDir:       "var/chc/runs",
			ValidationCommand: "go test ./...",
		},
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.DockerfilePath != filepath.Join(dir, "ops", "runtime.Dockerfile") {
		t.Fatalf("DockerfilePath = %q, want custom Dockerfile path", plan.DockerfilePath)
	}

	if plan.Image != "example/runtime:dev" {
		t.Fatalf("Image = %q, want custom image", plan.Image)
	}

	if plan.ContainerWorkdir != "/repo" {
		t.Fatalf("ContainerWorkdir = %q, want /repo", plan.ContainerWorkdir)
	}

	if plan.Mounts[0].Target != "/repo" {
		t.Fatalf("Mount target = %q, want /repo", plan.Mounts[0].Target)
	}

	if plan.ArtifactDir != filepath.Join(dir, "var", "chc", "runs") {
		t.Fatalf("ArtifactDir = %q, want custom artifact dir", plan.ArtifactDir)
	}

	if plan.ValidationCommand != "go test ./..." {
		t.Fatalf("ValidationCommand = %q, want custom command", plan.ValidationCommand)
	}
}

func TestPlanRuntimeUsesRunArtifactPaths(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		RunID:      "run-123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	wantRunDir := filepath.Join(dir, ".chc", "user", "runtime", "runs", "run-123")
	if plan.RunArtifactDir != wantRunDir {
		t.Fatalf("RunArtifactDir = %q, want %q", plan.RunArtifactDir, wantRunDir)
	}

	if plan.ResultPath != filepath.Join(wantRunDir, "result.json") {
		t.Fatalf("ResultPath = %q, want run result path", plan.ResultPath)
	}

	if plan.LogPath != filepath.Join(wantRunDir, "runtime.log") {
		t.Fatalf("LogPath = %q, want run log path", plan.LogPath)
	}

	if plan.WorkerContextPath != filepath.Join(wantRunDir, "context.md") {
		t.Fatalf("WorkerContextPath = %q, want run context path", plan.WorkerContextPath)
	}

	if plan.ValidationPath != filepath.Join(wantRunDir, "validation.log") {
		t.Fatalf("ValidationPath = %q, want run validation path", plan.ValidationPath)
	}
}

func TestPlanRuntimeUsesTicketScopedRunArtifactPaths(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		TicketKey:  "CHC-1",
		RunID:      "exec-123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	wantRunDir := filepath.Join(dir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-123")
	if plan.RunArtifactDir != wantRunDir {
		t.Fatalf("RunArtifactDir = %q, want %q", plan.RunArtifactDir, wantRunDir)
	}

	if plan.ResultPath != filepath.Join(wantRunDir, "result.json") {
		t.Fatalf("ResultPath = %q, want ticket-scoped result path", plan.ResultPath)
	}

	if plan.WorkerContextPath != filepath.Join(wantRunDir, "context.md") {
		t.Fatalf("WorkerContextPath = %q, want ticket-scoped context path", plan.WorkerContextPath)
	}
}

func TestPlanRuntimeNormalizesRelativeContainerWorkdirMount(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		Config: RuntimeConfig{
			Workdir: "workspace",
		},
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.Mounts[0].Source != dir {
		t.Fatalf("Mount source = %q, want %q", plan.Mounts[0].Source, dir)
	}

	if plan.Mounts[0].Target != "/workspace" {
		t.Fatalf("Mount target = %q, want /workspace", plan.Mounts[0].Target)
	}

	if plan.Mounts[0].ReadOnly {
		t.Fatal("Mount is read-only, want writable project mount")
	}
}

func TestCodexCredentialMountsUseReadOnlyRootCodexPaths(t *testing.T) {
	homeDir := t.TempDir()
	writeRuntimePlanTestFile(t, filepath.Join(homeDir, ".codex", "auth.json"), "{}\n")
	writeRuntimePlanTestFile(t, filepath.Join(homeDir, ".codex", "config.toml"), "model = \"gpt-5.5\"\n")

	mounts := codexCredentialMounts(homeDir)
	if len(mounts) != 2 {
		t.Fatalf("len(mounts) = %d, want 2", len(mounts))
	}

	wantTargets := []string{
		"/root/.codex/auth.json",
		"/root/.codex/config.toml",
	}

	for index, wantTarget := range wantTargets {
		if mounts[index].Target != wantTarget {
			t.Fatalf("mount target %d = %q, want %q", index, mounts[index].Target, wantTarget)
		}

		if !mounts[index].ReadOnly {
			t.Fatalf("mount %d is writable, want read-only", index)
		}
	}
}

func writeRuntimePlanTestFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create runtime plan test directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write runtime plan test file: %v", err)
	}
}
