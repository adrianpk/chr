package core

import (
	"path/filepath"
	"testing"
)

func TestPlanContextCompositionDerivesDockerVisibleArtifactPaths(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "user", "runtime", "runs", "CHR-1", "exec-1")

	plan, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "user", "tickets", "CHR-1-example.md"),
		ExecutionID: "exec-1",
		Plan: RuntimePlan{
			ProjectDir:        dir,
			RunArtifactDir:    runDir,
			WorkerContextPath: filepath.Join(runDir, "context.md"),
		},
		Sources: []ContextSourceInventoryEntry{
			{
				ID:       "agent-default",
				Kind:     ContextSourceKindAgentIdentity,
				Class:    ContextSourceClassInvariant,
				Required: true,
				Included: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("PlanContextComposition returned error: %v", err)
	}

	if plan.Paths.MainContextPath != filepath.Join(runDir, "context.md") {
		t.Fatalf("MainContextPath = %q, want run context path", plan.Paths.MainContextPath)
	}

	if plan.Paths.AgentPath != filepath.Join(runDir, "context", "agent.md") {
		t.Fatalf("AgentPath = %q, want agent context artifact", plan.Paths.AgentPath)
	}

	if plan.Paths.SourceInventoryPath != filepath.Join(runDir, "context", "source-inventory.md") {
		t.Fatalf("SourceInventoryPath = %q, want source inventory artifact", plan.Paths.SourceInventoryPath)
	}

	if plan.Readiness.Status != ContextReadinessReady {
		t.Fatalf("Readiness.Status = %q, want ready", plan.Readiness.Status)
	}
}

func TestPlanContextCompositionBlocksWhenRequiredSourcesAreMissing(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "user", "runtime", "runs", "CHR-1", "exec-1")

	plan, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "user", "tickets", "CHR-1-example.md"),
		ExecutionID: "exec-1",
		Plan: RuntimePlan{
			ProjectDir:        dir,
			RunArtifactDir:    runDir,
			WorkerContextPath: filepath.Join(runDir, "context.md"),
		},
		Sources: []ContextSourceInventoryEntry{
			{
				ID:       "protocols",
				Kind:     ContextSourceKindProtocol,
				Class:    ContextSourceClassInvariant,
				Required: true,
				Included: false,
				Reason:   "AGENTS.md was not found",
			},
		},
	})
	if err != nil {
		t.Fatalf("PlanContextComposition returned error: %v", err)
	}

	if plan.Readiness.Status != ContextReadinessBlocked {
		t.Fatalf("Readiness.Status = %q, want blocked", plan.Readiness.Status)
	}

	if len(plan.Readiness.MissingRequiredSources) != 1 {
		t.Fatalf("missing required sources = %d, want 1", len(plan.Readiness.MissingRequiredSources))
	}

	if plan.Readiness.MissingRequiredSources[0].ID != "protocols" {
		t.Fatalf("missing source id = %q, want protocols", plan.Readiness.MissingRequiredSources[0].ID)
	}
}

func TestPlanContextCompositionPreservesInvariantAndFreshSources(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "user", "runtime", "runs", "CHR-1", "exec-1")

	plan, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "user", "tickets", "CHR-1-example.md"),
		ExecutionID: "exec-1",
		Plan: RuntimePlan{
			ProjectDir:        dir,
			RunArtifactDir:    runDir,
			WorkerContextPath: filepath.Join(runDir, "context.md"),
		},
		Sources: []ContextSourceInventoryEntry{
			{
				ID:       "agent-default",
				Kind:     ContextSourceKindAgentIdentity,
				Class:    ContextSourceClassInvariant,
				Included: true,
			},
			{
				ID:       "current-tree",
				Kind:     ContextSourceKindFilesystemTree,
				Class:    ContextSourceClassFresh,
				Included: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("PlanContextComposition returned error: %v", err)
	}

	if plan.Sources[0].Class != ContextSourceClassInvariant {
		t.Fatalf("first source class = %q, want invariant", plan.Sources[0].Class)
	}

	if plan.Sources[1].Class != ContextSourceClassFresh {
		t.Fatalf("second source class = %q, want fresh", plan.Sources[1].Class)
	}
}

func TestPlanContextCompositionRejectsInvalidSourceKind(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chr", "user", "runtime", "runs", "CHR-1", "exec-1")

	_, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "user", "tickets", "CHR-1-example.md"),
		ExecutionID: "exec-1",
		Plan: RuntimePlan{
			ProjectDir:        dir,
			RunArtifactDir:    runDir,
			WorkerContextPath: filepath.Join(runDir, "context.md"),
		},
		Sources: []ContextSourceInventoryEntry{
			{
				ID:    "unknown",
				Kind:  ContextSourceKind("unknown"),
				Class: ContextSourceClassFresh,
			},
		},
	})
	if err == nil {
		t.Fatal("PlanContextComposition returned nil error")
	}
}
