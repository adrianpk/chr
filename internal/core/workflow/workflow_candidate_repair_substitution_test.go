package workflow

import (
	"context"
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeCandidateRegenerateGit struct{ OSWorkflowPreparationGit }

func (git nativeCandidateRegenerateGit) PrepareWorkflowRepairedIntegration(ctx context.Context, query PreparationQuery, _ MergerRepairState) (IntegrationProof, error) {
	return git.PrepareWorkflowIntegration(ctx, query)
}

// TestNativeCandidateSubstitution refuses a valid deterministic preparation
// substituted for accepted semantic repair, even when the unchanged gate passes.
func TestNativeCandidateSubstitution(t *testing.T) {
	f, request, harness := nativeCandidateRepairFixture(t, "rebase", "actionable_review_finding")
	store, id := f.f.store, f.f.run.ID

	if _, err := store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, f.f.git); err != nil {
		t.Fatal(err)
	}

	git := nativeCandidateRegenerateGit{f.f.git}
	if _, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, f.provider, git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("host gate substituted another P for accepted repair", err)
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider); err == nil {
		t.Fatal("substituted result produced publication authority")
	}

	assertWorkflowPreparationSource(t, f.f)
}
