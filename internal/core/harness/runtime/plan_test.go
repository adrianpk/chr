package runtime

import (
	"path/filepath"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestPlanRuntimeUsesDefaultConfig(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.WorkspaceMode != workspacecore.WorkWorkspaceModeProject {
		t.Fatalf("WorkspaceMode = %q, want %q", plan.WorkspaceMode, workspacecore.WorkWorkspaceModeProject)
	}

	wantPaths := map[string]string{
		"ProjectDir":            dir,
		"ProjectCheckoutDir":    dir,
		"ExecutionWorkspaceDir": dir,
		"WorkspaceDir":          filepath.Join(dir, ".chr"),
		"DockerfilePath":        filepath.Join(dir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		"ArtifactDir":           filepath.Join(dir, "ops", "default", "run"),
		"ResultPath":            filepath.Join(dir, "ops", "default", "run", "result.json"),
	}

	gotPaths := map[string]string{
		"ProjectDir":            plan.ProjectDir,
		"ProjectCheckoutDir":    plan.ProjectCheckoutDir,
		"ExecutionWorkspaceDir": plan.ExecutionWorkspaceDir,
		"WorkspaceDir":          plan.WorkspaceDir,
		"DockerfilePath":        plan.DockerfilePath,
		"ArtifactDir":           plan.ArtifactDir,
		"ResultPath":            plan.ResultPath,
	}
	for name, want := range wantPaths {
		if gotPaths[name] != want {
			t.Fatalf("%s = %q, want %q", name, gotPaths[name], want)
		}
	}

	if plan.Image != DefaultRuntimeImage {
		t.Fatalf("Image = %q, want %q", plan.Image, DefaultRuntimeImage)
	}

	if plan.ContainerWorkdir != DefaultRuntimeWorkdir {
		t.Fatalf("ContainerWorkdir = %q, want %q", plan.ContainerWorkdir, DefaultRuntimeWorkdir)
	}
}

func TestPlanRuntimeUsesDefaultArtifactPaths(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.WorkerContextPath != filepath.Join(dir, "ops", "default", "run", "context.md") {
		t.Fatalf("WorkerContextPath = %q, want default context path", plan.WorkerContextPath)
	}

	if plan.ValidationPath != filepath.Join(dir, "ops", "default", "run", "validation.log") {
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
			ArtifactDir:       "var/chr/runs",
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

	if plan.ArtifactDir != filepath.Join(dir, "var", "chr", "runs") {
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

	wantRunDir := filepath.Join(dir, "ops", "default", "run", "run-123")
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

func TestPlanRuntimeUsesCanonicalOpsArtifactPaths(t *testing.T) {
	dir := t.TempDir()

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		OpsLayout:  &layout,
		ArtifactID: "00001-ticket-run",
		RunID:      "exec-1",
		TicketKey:  "CHR-1",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	wantRunDir := filepath.Join(dir, "ops", "default", "run", "00001-ticket-run")
	if plan.ArtifactID != "00001-ticket-run" || plan.RunArtifactDir != wantRunDir {
		t.Fatalf("canonical run = (%q, %q)", plan.ArtifactID, plan.RunArtifactDir)
	}

	if plan.WorkerContextPath != filepath.Join(wantRunDir, "files", "context.md") {
		t.Fatalf("WorkerContextPath = %q", plan.WorkerContextPath)
	}

	if plan.ResultPath != filepath.Join(wantRunDir, "evidence", "result.json") {
		t.Fatalf("ResultPath = %q", plan.ResultPath)
	}
}

func TestPlanRuntimeUsesTicketScopedRunArtifactPaths(t *testing.T) {
	dir := t.TempDir()

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		TicketKey:  "CHR-1",
		RunID:      "exec-123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	wantRunDir := filepath.Join(dir, "ops", "default", "run", "CHR-1", "exec-123")
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

func TestPlanRuntimeRecordsConfiguredExecutionWorkspace(t *testing.T) {
	dir := t.TempDir()
	executionWorkspaceDir := filepath.Join(dir, ".chr", "default", "runtime", "worktrees", "CHR-1")

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		Workspace: workspacecore.WorkWorkspaceConfig{
			Mode:            workspacecore.WorkWorkspaceModeWorktree,
			WorktreeRoot:    workspacecore.DefaultWorktreeRoot,
			WorktreeBase:    workspacecore.WorktreeBasePolicyRemoteIntegrationBranch,
			WorktreeCleanup: workspacecore.WorktreeCleanupPolicyAfterMerge,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
		TicketKey:             "CHR-1",
		TicketBranch:          "chr/CHR-1",
		WorkspaceBaseRef:      "origin/dev",
		WorkspaceBaseSHA:      "abc123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	wantArtifactDir := filepath.Join(executionWorkspaceDir, "ops", "default", "run")
	wantFields := map[string]string{
		"ProjectCheckoutDir":    dir,
		"ExecutionWorkspaceDir": executionWorkspaceDir,
		"WorkspaceDir":          filepath.Join(executionWorkspaceDir, ".chr"),
		"ArtifactDir":           wantArtifactDir,
		"ResultPath":            filepath.Join(wantArtifactDir, "result.json"),
		"WorkspaceMode":         string(workspacecore.WorkWorkspaceModeWorktree),
		"TicketKey":             "CHR-1",
		"TicketBranch":          "chr/CHR-1",
		"WorkspaceBaseRef":      "origin/dev",
		"WorkspaceBaseSHA":      "abc123",
		"MountSource":           executionWorkspaceDir,
		"MountDescription":      "execution workspace",
	}
	gotFields := map[string]string{
		"ProjectCheckoutDir":    plan.ProjectCheckoutDir,
		"ExecutionWorkspaceDir": plan.ExecutionWorkspaceDir,
		"WorkspaceDir":          plan.WorkspaceDir,
		"ArtifactDir":           plan.ArtifactDir,
		"ResultPath":            plan.ResultPath,
		"WorkspaceMode":         string(plan.WorkspaceMode),
		"TicketKey":             plan.TicketKey,
		"TicketBranch":          plan.TicketBranch,
		"WorkspaceBaseRef":      plan.WorkspaceBaseRef,
		"WorkspaceBaseSHA":      plan.WorkspaceBaseSHA,
		"MountSource":           plan.Mounts[0].Source,
		"MountDescription":      plan.Mounts[0].Description,
	}

	for name, want := range wantFields {
		if gotFields[name] != want {
			t.Fatalf("%s = %q, want %q", name, gotFields[name], want)
		}
	}
}

func TestPlanRuntimePreservesProjectModeWorkspaceMount(t *testing.T) {
	dir := t.TempDir()
	executionWorkspaceDir := filepath.Join(dir, ".chr", "default", "runtime", "worktrees", "CHR-1")

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: dir,
		Workspace: workspacecore.WorkWorkspaceConfig{
			Mode: workspacecore.WorkWorkspaceModeProject,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	if plan.ExecutionWorkspaceDir != dir {
		t.Fatalf("ExecutionWorkspaceDir = %q, want project dir %q", plan.ExecutionWorkspaceDir, dir)
	}

	if plan.Mounts[0].Source != dir {
		t.Fatalf("Mount source = %q, want project dir %q", plan.Mounts[0].Source, dir)
	}

	if plan.Mounts[0].Target != "/workspace" {
		t.Fatalf("Mount target = %q, want /workspace", plan.Mounts[0].Target)
	}

	if plan.Mounts[0].Description != "project workspace" {
		t.Fatalf("Mount description = %q, want project workspace", plan.Mounts[0].Description)
	}
}
