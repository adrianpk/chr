package chc

import (
	"path/filepath"
	"strings"
	"testing"
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
	for _, want := range []string{"Context Readiness Handoff", "/workspace/.chc/runtime/runs/CHC-1/exec-1/context.md", "PAT-backed"} {
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
	sources := []ContextSourceInventoryEntry{
		{
			ID:       "gograph",
			Kind:     ContextSourceKindGoGraph,
			Class:    ContextSourceClassFresh,
			Included: true,
		},
	}
	composition := testReadinessComposition(t, plan, sources)

	handoff := EvaluateContextReadinessHandoff(ContextReadinessHandoffOptions{
		Composition:   composition,
		Plan:          plan,
		RequiredKinds: []ContextSourceKind{ContextSourceKindGoGraph},
	})

	if handoff.Status != ContextReadinessReady {
		t.Fatalf("Status = %q, want ready", handoff.Status)
	}
}

func testReadinessRuntimePlan(dir string) RuntimePlan {
	runDir := filepath.Join(dir, ".chc", "runtime", "runs", "CHC-1", "exec-1")

	return RuntimePlan{
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
	plan RuntimePlan,
	sources []ContextSourceInventoryEntry,
) ContextCompositionPlan {
	t.Helper()

	composition, err := PlanContextComposition(ContextComposerOptions{
		TicketKey:   "CHC-1",
		TicketPath:  filepath.Join(plan.ProjectDir, ".chc", "tickets", "CHC-1.md"),
		ExecutionID: "exec-1",
		Plan:        plan,
		Sources:     sources,
	})
	if err != nil {
		t.Fatalf("PlanContextComposition returned error: %v", err)
	}

	return composition
}

func completeReadinessSources() []ContextSourceInventoryEntry {
	return []ContextSourceInventoryEntry{
		{
			ID:       "agent-default",
			Kind:     ContextSourceKindAgentIdentity,
			Class:    ContextSourceClassInvariant,
			Included: true,
			Required: true,
		},
		{
			ID:       "agents:AGENTS.md",
			Kind:     ContextSourceKindProtocol,
			Class:    ContextSourceClassInvariant,
			Included: true,
			Required: true,
		},
		{
			ID:       "project:demo",
			Kind:     ContextSourceKindProject,
			Class:    ContextSourceClassInvariant,
			Included: true,
			Required: true,
		},
		{
			ID:       "ticket:CHC-1.md",
			Kind:     ContextSourceKindDocument,
			Class:    ContextSourceClassFresh,
			Included: true,
			Required: true,
		},
	}
}
