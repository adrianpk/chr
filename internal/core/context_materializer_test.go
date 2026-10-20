package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextPackageMaterializerWritesArtifactLayout(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testContextPackageRuntimePlan(dir, runDir)
	composition := testContextPackageComposition(t, plan)

	err := NewContextPackageMaterializer().MaterializeContextPackage(context.Background(), MaterializeContextPackageOptions{
		Composition:     composition,
		Plan:            plan,
		Agent:           "# Agent\n",
		Project:         "# Project\n",
		Protocols:       "# Protocols\n",
		Documents:       "# Documents\n",
		CodeContext:     "# Code Context\n",
		PreviousAttempt: "# Previous Attempt\n",
	})
	if err != nil {
		t.Fatalf("MaterializeContextPackage returned error: %v", err)
	}

	assertContextPackageFile(t, composition.Paths.MainContextPath, "Worker Context")
	assertContextPackageFile(t, composition.Paths.AgentPath, "# Agent")
	assertContextPackageFile(t, composition.Paths.ProjectPath, "# Project")
	assertContextPackageFile(t, composition.Paths.ProtocolsPath, "# Protocols")
	assertContextPackageFile(t, composition.Paths.DocumentsPath, "# Documents")
	assertContextPackageFile(t, composition.Paths.CodeContextPath, "# Code Context")
	assertContextPackageFile(t, composition.Paths.PreviousAttemptPath, "# Previous Attempt")
	assertContextPackageFile(t, composition.Paths.SourceInventoryPath, "agent-default")
}

func TestRenderMainContextPackageUsesDockerVisiblePaths(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testContextPackageRuntimePlan(dir, runDir)
	composition := testContextPackageComposition(t, plan)

	rendered := RenderMainContextPackage(MaterializeContextPackageOptions{
		Composition: composition,
		Plan:        plan,
	})

	for _, want := range []string{
		"/workspace/.chr/default/runtime/runs/CHR-1/exec-1/context/agent.md",
		"/workspace/.chr/default/runtime/runs/CHR-1/exec-1/context/code-context.md",
		"/workspace/.chr/default/runtime/runs/CHR-1/exec-1/result.json",
		"agent-default",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("main context missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderMainContextPackageUsesWorktreeDockerVisiblePaths(t *testing.T) {
	projectDir := t.TempDir()
	executionWorkspaceDir := filepath.Join(projectDir, ".chr", "default", "worktrees", "CHR-1")
	runDir := filepath.Join(executionWorkspaceDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testContextPackageRuntimePlan(projectDir, runDir)
	plan.ProjectCheckoutDir = projectDir
	plan.ExecutionWorkspaceDir = executionWorkspaceDir
	plan.WorkspaceMode = WorkWorkspaceModeWorktree
	composition := testContextPackageComposition(t, plan)

	rendered := RenderMainContextPackage(MaterializeContextPackageOptions{
		Composition: composition,
		Plan:        plan,
	})

	for _, want := range []string{
		"/workspace/.chr/default/runtime/runs/CHR-1/exec-1/context/agent.md",
		"/workspace/.chr/default/runtime/runs/CHR-1/exec-1/context/code-context.md",
		"/workspace/.chr/default/runtime/runs/CHR-1/exec-1/result.json",
		"host: `" + filepath.Join(executionWorkspaceDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1", "result.json") + "`",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("main context missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderContextPackageDryRunShowsIncludedAndSkippedSources(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testContextPackageRuntimePlan(dir, runDir)
	composition := testContextPackageComposition(t, plan)

	rendered := RenderContextPackageDryRun(MaterializeContextPackageOptions{
		Composition: composition,
		Plan:        plan,
		AdditionalSources: []ContextSourceInventoryEntry{
			{
				ID:       "missing-doc",
				Kind:     ContextSourceKindDocument,
				Class:    ContextSourceClassFresh,
				Path:     "docs/missing.md",
				Included: false,
				Reason:   "file was not found",
			},
		},
	})

	for _, want := range []string{"agent-default", "included", "missing-doc", "skipped", "file was not found"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("dry run missing %q:\n%s", want, rendered)
		}
	}
}

func testContextPackageRuntimePlan(dir, runDir string) RuntimePlan {
	return RuntimePlan{
		ProjectDir:        dir,
		ContainerWorkdir:  "/workspace",
		RunArtifactDir:    runDir,
		WorkerContextPath: filepath.Join(runDir, "context.md"),
		ResultPath:        filepath.Join(runDir, "result.json"),
		LogPath:           filepath.Join(runDir, "runtime.log"),
		ValidationPath:    filepath.Join(runDir, "validation.log"),
	}
}

func testContextPackageComposition(t *testing.T, plan RuntimePlan) ContextCompositionPlan {
	t.Helper()

	composition, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(plan.ProjectDir, ".chr", "default", "tickets", "CHR-1.md"),
		ExecutionID: "exec-1",
		Plan:        plan,
		Sources: []ContextSourceInventoryEntry{
			{
				ID:       "agent-default",
				Kind:     ContextSourceKindAgentIdentity,
				Class:    ContextSourceClassInvariant,
				Path:     "agent.md",
				Included: true,
				Required: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("PlanContextComposition returned error: %v", err)
	}

	return composition
}

func assertContextPackageFile(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if !strings.Contains(string(content), want) {
		t.Fatalf("%s missing %q:\n%s", path, want, string(content))
	}
}
