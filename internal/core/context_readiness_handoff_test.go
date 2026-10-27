package core

import (
	"path/filepath"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
)

func TestEvaluateContextReadinessHandoffReady(t *testing.T) {
	dir := t.TempDir()
	plan := testReadinessRuntimePlan(dir)
	composition := testReadinessComposition(t, plan, completeReadinessSources())

	handoff := EvaluateContextReadinessHandoff(ContextReadinessHandoffOptions{
		Composition: composition,
		Plan:        plan,
	})

	if handoff.Status != ContextReadinessReady {
		t.Fatalf("Status = %q, want ready", handoff.Status)
	}

	if handoff.Completeness.Passed != handoff.Completeness.Total {
		t.Fatalf("Completeness = %#v, want complete", handoff.Completeness)
	}

	rendered := RenderContextReadinessHandoff(handoff)
	for _, want := range []string{"Context Readiness Handoff", "/workspace/.chr/default/runtime/runs/CHR-1/exec-1/context.md", "PAT-backed"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("handoff missing %q:\n%s", want, rendered)
		}
	}
}

func TestEvaluateContextReadinessHandoffBlocksMissingProtocolAndValidation(t *testing.T) {
	dir := t.TempDir()
	plan := testReadinessRuntimePlan(dir)
	plan.ValidationCommand = ""

	sources := completeReadinessSources()
	sources[1].Included = false
	sources[1].Reason = "AGENTS.md was not found"
	composition := testReadinessComposition(t, plan, sources)

	handoff := EvaluateContextReadinessHandoff(ContextReadinessHandoffOptions{
		Composition: composition,
		Plan:        plan,
	})

	if handoff.Status != ContextReadinessBlocked {
		t.Fatalf("Status = %q, want blocked", handoff.Status)
	}

	rendered := RenderContextReadinessHandoff(handoff)
	for _, want := range []string{"source:protocol", "validation_command", "context composition readiness is blocked"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("blocked handoff missing %q:\n%s", want, rendered)
		}
	}
}

func TestEvaluateContextReadinessHandoffSupportsCustomRequiredKinds(t *testing.T) {
	dir := t.TempDir()
	plan := testReadinessRuntimePlan(dir)
	sources := []authoringcore.ContextSourceInventoryEntry{
		{
			ID:       "gograph",
			Kind:     authoringcore.ContextSourceKindGoGraph,
			Class:    authoringcore.ContextSourceClassFresh,
			Included: true,
		},
	}
	composition := testReadinessComposition(t, plan, sources)

	handoff := EvaluateContextReadinessHandoff(ContextReadinessHandoffOptions{
		Composition:   composition,
		Plan:          plan,
		RequiredKinds: []authoringcore.ContextSourceKind{authoringcore.ContextSourceKindGoGraph},
	})

	if handoff.Status != ContextReadinessReady {
		t.Fatalf("Status = %q, want ready", handoff.Status)
	}
}

func testReadinessRuntimePlan(dir string) harnessruntime.RuntimePlan {
	runDir := filepath.Join(dir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")

	return harnessruntime.RuntimePlan{
		ProjectDir:        dir,
		ContainerWorkdir:  "/workspace",
		RunArtifactDir:    runDir,
		WorkerContextPath: filepath.Join(runDir, "context.md"),
		ResultPath:        filepath.Join(runDir, "result.json"),
		LogPath:           filepath.Join(runDir, "runtime.log"),
		ValidationPath:    filepath.Join(runDir, "validation.log"),
		ValidationCommand: "make verify",
	}
}

func testReadinessComposition(
	t *testing.T,
	plan harnessruntime.RuntimePlan,
	sources []authoringcore.ContextSourceInventoryEntry,
) ContextCompositionPlan {
	t.Helper()

	composition, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(plan.ProjectDir, ".chr", "default", "tickets", "CHR-1.md"),
		ExecutionID: "exec-1",
		Plan:        plan,
		Sources:     sources,
	})
	if err != nil {
		t.Fatalf("PlanContextComposition returned error: %v", err)
	}

	return composition
}

func completeReadinessSources() []authoringcore.ContextSourceInventoryEntry {
	return []authoringcore.ContextSourceInventoryEntry{
		{
			ID:       "agent-default",
			Kind:     authoringcore.ContextSourceKindAgentIdentity,
			Class:    authoringcore.ContextSourceClassInvariant,
			Included: true,
			Required: true,
		},
		{
			ID:       "agents:AGENTS.md",
			Kind:     authoringcore.ContextSourceKindProtocol,
			Class:    authoringcore.ContextSourceClassInvariant,
			Included: true,
			Required: true,
		},
		{
			ID:       "project:demo",
			Kind:     authoringcore.ContextSourceKindProject,
			Class:    authoringcore.ContextSourceClassInvariant,
			Included: true,
			Required: true,
		},
		{
			ID:       "ticket:CHR-1.md",
			Kind:     authoringcore.ContextSourceKindDocument,
			Class:    authoringcore.ContextSourceClassFresh,
			Included: true,
			Required: true,
		},
	}
}
