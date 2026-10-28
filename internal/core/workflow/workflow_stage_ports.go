package workflow

import (
	"context"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowIntegrationValidation struct {
	store    *Store
	claim    *Claim
	prepared workflowPreparedTask
	provider PublicationGuardProvider
}

func (operation workflowIntegrationValidation) Validate(ctx context.Context, result *workflowstate.TaskResult) error {
	return operation.store.validateWorkflowTask(ctx, operation.claim, operation.prepared, result)
}

func (operation workflowIntegrationValidation) Accept(ctx context.Context, completion workflowintegrate.Completion) error {
	return operation.store.finishWorkflowIntegrationValidation(ctx, operation, completion)
}

type workflowReviewDecisions struct {
	store    *Store
	claim    *Claim
	prepared workflowPreparedTask
	provider PullRequestProvider
}

func (operation workflowReviewDecisions) Accept(ctx context.Context, completion workflowreview.Completion) (workflowstate.ReviewBinding, error) {
	return operation.store.finishWorkflowReview(ctx, operation, completion)
}

type workflowSubmissionResponses struct {
	store        *Store
	invocationID string
	submissionID string
}

func (operation workflowSubmissionResponses) Retain(ctx context.Context, response workflowforge.Response) (workflowstate.Submission, error) {
	return operation.store.retainWorkflowSubmissionResponse(ctx, operation, response)
}

type workflowClosingWorkspace struct {
	store *Store
	claim *workflowClosingClaim
}

func (operation workflowClosingWorkspace) Prepare(ctx context.Context) (string, error) {
	return operation.store.prepareWorkflowClosingCheckout(ctx, operation.claim)
}

func (operation workflowClosingWorkspace) Check(ctx context.Context, directory string) error {
	return checkWorkflowClosingCheckout(ctx, directory, operation.claim.input)
}

func (operation workflowClosingWorkspace) Run(ctx context.Context, directory string, command []string) (workflowstate.CommandResult, error) {
	return operation.store.runWorkflowClosingCommand(ctx, operation.claim, directory, command)
}

var (
	_ workflowintegrate.Validation = workflowIntegrationValidation{}
	_ workflowreview.Decisions     = workflowReviewDecisions{}
	_ workflowforge.Responses      = workflowSubmissionResponses{}
	_ workflowcloseout.Workspace   = workflowClosingWorkspace{}
	_ workflowcloseout.Commands    = workflowClosingWorkspace{}
)

// workflowIntegrationRepairs composes pure repair checks without sibling imports.
type workflowIntegrationRepairs struct{}

func (workflowIntegrationRepairs) Check(proof workflowstate.IntegrationProof) error {
	if err := workflowmerge.CheckRepairs(proof); err != nil {
		return err
	}

	return workflowmerge.CheckHistory(proof)
}

var _ workflowintegrate.RepairProof = workflowIntegrationRepairs{}
