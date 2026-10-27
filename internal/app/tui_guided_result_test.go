package app

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// TestGuidedUseCaseTUIResultGuard refuses a completed label when the native
// result is missing, partial or invalid, even if a terminal counter says complete.
func TestGuidedUseCaseTUIResultGuard(t *testing.T) {
	for _, name := range []string{"missing", "partial", "invalid", "verified"} {
		t.Run(name, func(t *testing.T) {
			session := &authoringcore.AuthoringSession{Guided: &authoringcore.GuidedUseCaseState{Status: "completed", Stage: "results"}}

			observed := GuidedUseCaseInspection{Session: session}
			if name != "missing" {
				observed.Result = &core.WorkflowResultInspection{Complete: name == "verified"}
			}

			if name == "invalid" {
				observed.ResultProblem = "Retained evidence needs recovery."
			}

			want := "blocked"
			if name == "verified" {
				want = "completed"
			}

			if view := tuiGuidedView(observed); view.Status != want || guidedResultSnapshot(observed).Status != want {
				t.Fatal("completion escaped the verified result boundary", view.Status)
			}
		})
	}
}
