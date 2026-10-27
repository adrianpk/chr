package core

import (
	"path/filepath"
	"testing"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowReviewRevocation re-resolves project policy and the current target,
// preserving historical decisions while replacing a stale grant with a new wait.
func TestWorkflowReviewRevocation(t *testing.T) {
	store, _, run, provider := workflowReviewFixture(t, ReviewWorkflowModeTrustedAutoMerge)

	old, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil)
	if err != nil {
		t.Fatal(err)
	}

	workflowReviewConfig(t, store, ReviewWorkflowModeHumanReview)

	if _, err := store.CheckWorkflowReviewGrant(t.Context(), run.ID, provider); err == nil {
		t.Fatal("policy change retained old grant")
	}

	current, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil)
	if err != nil || current.State != "waiting_approval" || current.InputHash == old.InputHash || current.Round != old.Round+1 {
		t.Fatal("changed policy reused another candidate grant")
	}

	if _, err := store.ApproveWorkflowSlice(t.Context(), run.ID, old.InputHash, provider); err == nil {
		t.Fatal("old displayed hash approved the new policy")
	}

	if _, err := store.ApproveWorkflowSlice(t.Context(), run.ID, current.InputHash, provider); err != nil {
		t.Fatal(err)
	}

	if _, err := opsfile.Read(filepath.Join(store.layout.Root(), old.ResultRef)); err != nil {
		t.Fatal("revocation discarded immutable historical decision")
	}
}

// TestWorkflowReviewDecisionChecks rejects fabricated pass evidence, repair
// outside membership, subjective blocking notes and decisions for another input.
func TestWorkflowReviewDecisionChecks(t *testing.T) {
	store, _, run, provider := workflowReviewFixture(t, ReviewWorkflowModeHumanReview)
	if _, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil); err != nil {
		t.Fatal(err)
	}

	file, err := store.readWorkflowInvocation(run.ID)
	if err != nil {
		t.Fatal(err)
	}

	input, _, err := store.readWorkflowReview(file)
	if err != nil {
		t.Fatal(err)
	}

	valid := WorkflowReviewDecision{BindingHash: file.Review.InputHash, Result: AgenticReviewResult{Decision: IntegrationReviewDecisionPass, Summary: "Required evidence inspected."}}

	for _, test := range []struct {
		name   string
		change func(*WorkflowReviewDecision)
	}{
		{"another input", func(d *WorkflowReviewDecision) { d.BindingHash = "another" }},
		{"pass findings", func(d *WorkflowReviewDecision) {
			d.Result.BlockingFindings = []workflowstate.AgenticReviewFinding{{Category: "correctness", Title: "Broken", Detail: "Wrong result."}}
		}},
		{"cosmetic", func(d *WorkflowReviewDecision) {
			d.Result.Decision = IntegrationReviewDecisionNeedsChanges
			d.RepairTaskKey = input.Slice.Tasks[0].Key
			d.Result.BlockingFindings = []workflowstate.AgenticReviewFinding{{Category: "cosmetic", Title: "Style", Detail: "Optional formatting."}}
		}},
		{"unknown task", func(d *WorkflowReviewDecision) {
			d.Result.Decision = IntegrationReviewDecisionNeedsChanges
			d.RepairTaskKey = "another task"
			d.Result.BlockingFindings = []workflowstate.AgenticReviewFinding{{Category: "correctness", Title: "Broken", Detail: "Wrong result."}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			d := valid
			test.change(&d)

			if err := validateWorkflowReviewDecision(d, input); err == nil {
				t.Fatal("unsafe role evidence accepted")
			}
		})
	}
}
