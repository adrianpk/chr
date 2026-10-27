package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Invalid generated identities cannot create a plan, map or ticket. Stable IDs
// and explicit order come from structured output, never Markdown inference.
func TestPlanOrderRejectsInvalid(t *testing.T) {
	cases := []struct {
		name   string
		change func(*generationcore.GeneratedPlan)
	}{
		{"missing map", func(plan *generationcore.GeneratedPlan) { plan.Order = nil }},
		{"allocated key", func(plan *generationcore.GeneratedPlan) { plan.Order[0].Tasks[0].TicketKey = "CHR-1" }},
		{"allocated reference", func(plan *generationcore.GeneratedPlan) {
			plan.Order[0].Tasks[0].TicketRef = "ticket/00001-task/artifact.md"
		}},
		{"duplicate slice", func(plan *generationcore.GeneratedPlan) { plan.Order = append(plan.Order, plan.Order[0]) }},
		{"duplicate task", func(plan *generationcore.GeneratedPlan) {
			plan.Order[0].Tasks = append(plan.Order[0].Tasks, plan.Order[0].Tasks[0])
		}},
		{"invalid ID", func(plan *generationcore.GeneratedPlan) { plan.Order[0].ID = "Slice 1" }},
		{"missing order", func(plan *generationcore.GeneratedPlan) { plan.Order[0].Tasks[0].Order = 0 }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			project := t.TempDir()
			plan := orderedPlanCandidate()
			item.change(&plan)

			if _, err := WriteGeneratedPlan(t.Context(), project, plan); err == nil {
				t.Fatal("invalid generated order accepted")
			}

			paths, err := GeneratedPlanLayout(project, plan.Scope)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := os.Stat(filepath.Dir(paths.Path)); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("rejected model output wrote artifacts", err)
			}
		})
	}
}

// Replacement needs the exact observed hash after a user edits the current
// map. Conflicting attempts leave the retained prior plan and pair unchanged.
func TestPlanOrderEditedOutput(t *testing.T) {
	project := t.TempDir()

	plan := orderedPlanCandidate()
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal(err)
	}

	first := readTestPlanBinding(t, project, plan.Scope)
	reference := filepath.Join(project, "ops", "default", filepath.FromSlash(first.CurrentTrackerRef))

	edited := []byte("An observed manual edit.\n")
	if err := os.WriteFile(reference, edited, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := WriteGeneratedPlan(t.Context(), project, plan); !errors.Is(err, workflowstate.ErrProjectionStale) {
		t.Fatal("overwrote edited output", err)
	}

	plan.TrackerReplaceHash = contenthash.Sum([]byte("another observation"))
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); !errors.Is(err, workflowstate.ErrProjectionStale) {
		t.Fatal("accepted wrong observation", err)
	}

	plan.TrackerReplaceHash = contenthash.Sum(edited)
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal(err)
	}

	if second := readTestPlanBinding(t, project, plan.Scope); second.Revision != 2 {
		t.Fatal("failed attempts advanced revision", second)
	}

	if contenthash.Sum(readTestPlanImage(t, project, first.TrackerRef)) != first.TrackerHash {
		t.Fatal("replacement altered retained map")
	}
}

// A required pair cannot be downgraded to a body-only legacy publication.
func TestPlanOrderCannotDropBinding(t *testing.T) {
	project := t.TempDir()

	plan := orderedPlanCandidate()
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal(err)
	}

	plan.Order, plan.RequireOrderMap = nil, false
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err == nil {
		t.Fatal("required companion silently removed")
	}
}

// Inspection never treats a missing required image or an altered current map
// as a complete, current pair. GET-style observation does not repair the files.
func TestPlanOrderMissingCompanion(t *testing.T) {
	for _, name := range []string{"order.json", "tracker.md", "current tracker"} {
		t.Run(name, func(t *testing.T) {
			checkMissingPlanCompanion(t, name)
		})
	}
}

// Extending the v2 authoring journal permits only the bounded retained revision
// files and existing tracker envelopes. It does not widen the v1 path contract.
func TestPlanOrderJournalPaths(t *testing.T) {
	root := "ops/default/plan/00000-chosen-plan/files/plan-revisions/1-" + contenthash.Sum([]byte("body")) + "/"

	cases := []struct {
		name, path string
		valid      bool
	}{
		{"retained binding", root + "binding.json", true},
		{"retained body", root + "plan.md", true},
		{"tracker envelope", "ops/default/tracker/00000-chosen-tracking/metadata.json", true},
		{"other retained file", root + "state.json", false},
		{"zero revision", "ops/default/plan/00000-chosen-plan/files/plan-revisions/0-" + contenthash.Sum([]byte("body")) + "/plan.md", false},
		{"unbounded revision", "ops/default/plan/00000-chosen-plan/files/plan-revisions/4097-" + contenthash.Sum([]byte("body")) + "/plan.md", false},
		{"invalid hash", "ops/default/plan/00000-chosen-plan/files/plan-revisions/1-invalid/plan.md", false},
		{"nested tracker", "ops/default/tracker/00000-chosen-tracking/files/run.json", false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if validAuthoringJournalVersionPath(projectDumpJournalSchema, item.path) != item.valid {
				t.Fatal("incorrect v2 path boundary", item.path)
			}

			if validAuthoringJournalVersionPath("chr-authoring-journal/v1", item.path) {
				t.Fatal("v1 path contract changed", item.path)
			}
		})
	}
}

func checkMissingPlanCompanion(t *testing.T, name string) {
	t.Helper()

	project := t.TempDir()

	plan := orderedPlanCandidate()
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal(err)
	}

	binding := readTestPlanBinding(t, project, plan.Scope)

	reference := filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), name))
	if name == "current tracker" {
		reference = binding.CurrentTrackerRef
	}

	path := filepath.Join(project, "ops", "default", filepath.FromSlash(reference))
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if _, err := ObservePlanRevision(t.Context(), project, plan.Scope); err == nil {
		t.Fatal("incomplete pair reported current")
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("inspection repaired missing image", err)
	}
}
