package workflow

import (
	"slices"
	"testing"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeMergerDecisions prevents incompatible intent or unsupported file
// procedures from being represented as an implicitly resolved proposal.
func TestNativeMergerDecisions(t *testing.T) {
	for _, kind := range []string{"generated", "lockfile", "binary", "submodule", "lfs"} {
		t.Run(kind, func(t *testing.T) {
			_, input, now := nativeMergerInput(t)
			input.Conflicts[0].Kind = kind

			input, err := workflowmerge.SealInput(input, now)
			if err != nil {
				t.Fatal(err)
			}

			decisions, err := workflowmerge.RequiredDecisions(input)
			if err != nil || len(decisions) != 1 || workflowmerge.CheckResult(nativeMergerResolved(input), input) == nil {
				t.Fatal("unsupported procedure silently resolved", err)
			}

			input.Conflicts[0].RegenerationCommand = []string{"make", "regenerate"}
			input.Conflicts[0].ValidationCommand = slices.Clone(input.ValidationCommands[0])

			input, err = workflowmerge.SealInput(input, now)
			if err != nil {
				t.Fatal(err)
			}

			decisions, err = workflowmerge.RequiredDecisions(input)
			if err != nil || len(decisions) != 0 {
				t.Fatal("declared regeneration and frozen validation were ignored", err)
			}
		})
	}
}

// TestNativeMergerIntentDecision requires a concrete maintainer decision when
// frozen source and target requirements explicitly declare incompatible intent.
func TestNativeMergerIntentDecision(t *testing.T) {
	_, input, now := nativeMergerInput(t)
	input.Requirements[0].ConflictsWith = []string{input.Requirements[1].ID}

	input, err := workflowmerge.SealInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	decisions, err := workflowmerge.RequiredDecisions(input)
	if err != nil || len(decisions) != 1 || workflowmerge.CheckResult(nativeMergerResolved(input), input) == nil {
		t.Fatal("incompatible required intent was implicitly approved", err)
	}
}

// TestNativeMergerRegeneration requires host observations of both declared
// regeneration and frozen validation, including their exact input/output trees.
func TestNativeMergerRegeneration(t *testing.T) {
	f, query := nativeMergerRepairFixture(t, "squash")
	git := NewWorkflowPreparationGit(query.Repository, f.remote)

	state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
	if err != nil {
		t.Fatal(err)
	}

	repair := nativeMergerTreeRepair(t, state, state.Conflicts[0], "Regenerated source and target intent\n")
	repair.Input.Conflicts[0].Kind = "generated"
	repair.Input.Conflicts[0].RegenerationCommand = []string{"make", "regenerate"}
	repair.Input.Conflicts[0].ValidationCommand = slices.Clone(repair.Input.ValidationCommands[0])

	repair.Input, err = workflowmerge.SealInput(repair.Input, repair.Input.AdmittedAt)
	if err != nil {
		t.Fatal(err)
	}

	repair.Proposal.InputHash = repair.Input.InputHash
	repair.Proposal.ResolutionNotes[0].RegenerationCommand = []string{"make", "regenerate"}

	proof := workflowMergerRepairProof(state)
	if workflowmerge.ValidateWorkflowTreeRepair(proof, repair) == nil {
		t.Fatal("model procedure declaration substituted for host observations")
	}

	evidence := workflowstate.MergerRegenerationEvidence{ConflictID: repair.Input.Conflicts[0].ID, Command: workflowstate.CommandResult{Command: []string{"make", "regenerate"}}, Validation: workflowstate.CommandResult{Command: slices.Clone(repair.Input.ValidationCommands[0])}, InputTree: repair.BaselineTree, OutputTree: repair.ResolvedTree}

	repair.Regeneration = []workflowstate.MergerRegenerationEvidence{evidence}
	if workflowmerge.ValidateWorkflowTreeRepair(proof, repair) != nil {
		t.Fatal("complete host procedure observations were rejected")
	}

	repair.Regeneration[0].Validation.Command = []string{"make", "weaker-check"}
	if workflowmerge.ValidateWorkflowTreeRepair(proof, repair) == nil {
		t.Fatal("undeclared validation substituted for frozen validation")
	}
}
