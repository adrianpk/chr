package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Caller intent, rather than transient model output, identifies an operation.
// A retry after another generation retains its original revision and history.
func TestPlanPublicationRetry(t *testing.T) {
	project := t.TempDir()
	plan := orderedPlanCandidate()
	plan.CreateOnly = true

	plan.Operation = &generationcore.PlanPublicationOperation{ID: "first-operation", Scope: plan.Scope}
	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal(err)
	}

	first := readTestPlanBinding(t, project, plan.Scope)
	next := orderedPlanCandidate()

	next.Body = "Another explicitly requested generation."
	if _, err := WriteGeneratedPlan(t.Context(), project, next); err != nil {
		t.Fatal(err)
	}

	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal("committed create retry failed", err)
	}

	current := readTestPlanBinding(t, project, plan.Scope)

	result, found, err := LookupPlanPublication(t.Context(), project, *plan.Operation)
	if err != nil || !found || result.Binding != first || current.Revision != 2 {
		t.Fatal(result, found, err, current)
	}

	changed := *plan.Operation

	changed.Replace = true
	if _, _, err := LookupPlanPublication(t.Context(), project, changed); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("operation ID accepted changed intent", err)
	}
}

// Checksums and root-bound intent prevent a copied or corrupted receipt from
// authorizing replay in another project. Missing retained images also fail.
func TestPlanPublicationReceiptGuards(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, string, PlanPublicationResult) string
	}{
		{"another root", copyPlanReceiptProject},
		{"corrupt receipt", corruptPlanReceipt},
		{"missing retained body", removePlanReceiptBody},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			project := t.TempDir()
			plan := orderedPlanCandidate()
			plan.CreateOnly = true

			plan.Operation = &generationcore.PlanPublicationOperation{ID: "guard-operation", Scope: plan.Scope}
			if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
				t.Fatal(err)
			}

			result, found, err := LookupPlanPublication(t.Context(), project, *plan.Operation)
			if err != nil || !found {
				t.Fatal(err)
			}

			target := item.mutate(t, project, result)
			if _, _, err := LookupPlanPublication(t.Context(), target, *plan.Operation); err == nil {
				t.Fatal("invalid receipt replayed")
			}
		})
	}
}

func copyPlanReceiptProject(t *testing.T, project string, result PlanPublicationResult) string {
	t.Helper()

	target := t.TempDir()

	data, err := os.ReadFile(filepath.Join(project, authoringReceiptRef(result.OperationID)))
	if err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, filepath.Join(target, authoringReceiptRef(result.OperationID)), string(data))

	return target
}

func corruptPlanReceipt(t *testing.T, project string, result PlanPublicationResult) string {
	t.Helper()
	writeTestFile(t, filepath.Join(project, authoringReceiptRef(result.OperationID)), "{}")

	return project
}

func removePlanReceiptBody(t *testing.T, project string, result PlanPublicationResult) string {
	t.Helper()

	if err := os.Remove(filepath.Join(project, "ops", "default", result.Binding.BodyRef)); err != nil {
		t.Fatal(err)
	}

	return project
}

// Concurrent retries share the existing process-safe owner. Exactly one
// revision is published and all successful callers observe that same receipt.
func TestPlanPublicationConcurrentRetry(t *testing.T) {
	project := t.TempDir()
	plan := orderedPlanCandidate()
	plan.CreateOnly = true

	revision, err := PlanTargetRevision(t.Context(), project, plan.Scope)
	if err != nil {
		t.Fatal(err)
	}

	plan.ExpectedRevision = revision
	plan.Operation = &generationcore.PlanPublicationOperation{ID: "concurrent-operation", Scope: plan.Scope, ExpectedRevision: revision}
	results := make(chan error, 2)

	for range 2 {
		go func() { _, err := WriteGeneratedPlan(t.Context(), project, plan); results <- err }()
	}

	for range 2 {
		if err := <-results; err != nil {
			t.Fatal("concurrent retry failed", err)
		}
	}

	if binding := readTestPlanBinding(t, project, plan.Scope); binding.Revision != 1 {
		t.Fatal("retry allocated another revision", binding)
	}
}
