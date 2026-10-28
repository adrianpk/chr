package workflow

import (
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeDecisionReplay carries a typed replay owner or explicit actionable
// fix into the existing repair planner; an unrelated commit cannot become owner.
func TestNativeDecisionReplay(t *testing.T) {
	_, state := nativeMergerCleanReplay(t)
	_, original, now := nativeMergerInput(t)
	original.Repository, original.Source, original.Base, original.Method = state.Query.Repository, state.Query.SourceCommit, state.Query.BaseCommit, "rebase"
	original.AllowedPaths, original.Conflicts[0].Path = []string{"source.txt", "next.txt"}, "source.txt"
	conflict := original.Conflicts[0]
	conflict.ID, conflict.Path = "conflict-2", "next.txt"
	original.Conflicts = append(original.Conflicts, conflict)

	cases := []struct {
		name    string
		answer  workflowstate.IntegrationAnswer
		outcome string
		denied  bool
	}{
		{"owner", workflowstate.IntegrationAnswer{DecisionID: "replay-owner", Action: "resolve", OriginalCommit: state.Query.SourceCommit}, "replay_step", false},
		{"foreign owner", workflowstate.IntegrationAnswer{DecisionID: "replay-owner", Action: "resolve", OriginalCommit: state.Query.BaseCommit}, "", true},
		{"fix", workflowstate.IntegrationAnswer{DecisionID: "replay-owner", Action: "resolve", IntegrationFix: &workflowstate.ReviewFinding{Category: workflowstate.AgenticReviewFindingCorrectness, Title: "Retain compatible output", Detail: "Combined validation requires an explicit compatible integration repair"}}, "integration_fix", false},
		{"defer", workflowstate.IntegrationAnswer{DecisionID: "replay-owner", Action: "defer"}, "needs_decision", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := original
			input.OperatorDecisionRef = &input.SourceManifestRef
			input.OperatorAnswers = []workflowstate.IntegrationAnswer{tc.answer}

			input, err := SealWorkflowMergerInput(input, now)
			if err != nil {
				t.Fatal(err)
			}

			plan, err := PlanWorkflowMergerRepair(t.Context(), state.Query.Lane.Worktree, input, *state.Result, nil)
			if tc.denied {
				if !errors.Is(err, workflowstate.ErrConflict) {
					t.Fatal("unowned source step was accepted", err)
				}

				return
			}

			if err != nil || plan.Outcome != tc.outcome {
				t.Fatal("typed answer did not reach repair planner", err)
			}
		})
	}
}
