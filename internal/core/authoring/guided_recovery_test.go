package authoring

import (
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseRecoveryInventory pages all namespaces without adopting
// manual sessions or foreign projects, and rejects unbounded inventory requests.
func TestGuidedUseCaseRecoveryInventory(t *testing.T) {
	root, original := guidedRecoveryInventoryFixture(t)
	seen := readGuidedRecoveryInventory(t, root, original.ProjectID)

	if len(seen) != 3 || seen["manual"] || !seen["research"] || !seen["capture"] {
		t.Fatal("inventory lost namespaces or adopted manual authoring", seen)
	}

	foreign, err := RecoverGuidedHostPage(t.Context(), root, "foreign", 0, workflowstate.PageLimit)
	if err != nil || len(foreign.Runs) != 0 {
		t.Fatal("foreign project was recovered", err)
	}

	for _, bounds := range [][2]int{{-1, 1}, {workflowstate.RecordLimit + 1, 1}, {0, 0}, {0, workflowstate.PageLimit + 1}} {
		if _, err := RecoverGuidedHostPage(t.Context(), root, original.ProjectID, bounds[0], bounds[1]); !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal("inventory accepted invalid bounds", bounds, err)
		}
	}
}

func guidedRecoveryInventoryFixture(t *testing.T) (string, Session) {
	t.Helper()

	root, original := guidedCaptureFixture(t)
	for _, namespace := range []string{"research", "other"} {
		scope := GuidedUseCaseScope{RunID: namespace, ProjectID: original.ProjectID, Namespace: namespace}
		if _, err := StartGuidedUseCase(t.Context(), root, GuidedUseCaseStart{GuidedUseCaseScope: scope, Actor: "operator", ProgressionIntent: true}); err != nil {
			t.Fatal(err)
		}
	}

	manual := original

	manual.ID, manual.Guided = "manual", nil
	if _, err := NewSessionRepository(root).CreateSession(t.Context(), manual); err != nil {
		t.Fatal(err)
	}

	return root, original
}

func readGuidedRecoveryInventory(t *testing.T, root, project string) map[string]bool {
	t.Helper()

	seen := map[string]bool{}

	offset := 0
	for range 4 {
		page, err := RecoverGuidedHostPage(t.Context(), root, project, offset, 2)
		if err != nil {
			t.Fatal(err)
		}

		for _, session := range page.Runs {
			seen[session.ID] = true
		}

		if page.NextOffset < 0 {
			break
		}

		offset = page.NextOffset
	}

	return seen
}
