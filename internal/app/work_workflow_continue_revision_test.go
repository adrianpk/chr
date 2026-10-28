package app

import (
	"context"
	"errors"
	"testing"

	"github.com/urfave/cli/v3"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

type unavailableWorkflowPreparation struct{}

func (unavailableWorkflowPreparation) PrepareWorkflowIntegration(context.Context, workflowcore.PreparationQuery) (workflowcore.IntegrationProof, error) {
	return workflowcore.IntegrationProof{}, errors.New("publisher credential unavailable")
}

// The CLI must bind explicit continuation to its observed revision. Otherwise
// recovering corrected credentials silently returns the same blocked state.
func TestWorkContinuePreparation(t *testing.T) {
	operations, id, original := serviceWorkflowPreparationReady(t)
	operations.workflowPullRequests = &serviceNativeReviewFixture{original, "policy-v1"}
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")
	serviceNativeReviewAdmit(t, operations, id)

	preparer := operations.workflowPreparationGit
	operations.workflowPreparationGit = unavailableWorkflowPreparation{}

	if _, err := operations.PrepareWorkflowSliceIntegration(t.Context(), id); err == nil {
		t.Fatal("preparation did not fail")
	}

	operations.workflowPreparationGit = preparer

	before, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	command := &cli.Command{Name: "continue", Flags: []cli.Flag{&cli.IntFlag{Name: "max-steps", Value: 1}}, Action: func(ctx context.Context, cmd *cli.Command) error {
		return performWorkflowControl(ctx, operations, id, "continue", cmd)
	}}
	if err := command.Run(t.Context(), []string{"continue"}); err != nil {
		t.Fatal(err)
	}

	after, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || after.Preparation == nil || after.Preparation.State != "validated" || after.PreparationCount != before.PreparationCount+1 {
		t.Fatal("explicit CLI continuation did not recover preparation", err)
	}

	if after.Cursor != before.Cursor || len(after.Attempts) != len(before.Attempts)+1 {
		t.Fatal("CLI continuation changed the worker extent or repeated work")
	}
}
