package app

import (
	"context"
	"fmt"
	"slices"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

// WorkflowWorkPreview reports the exact scope and current publication policy.
type WorkflowWorkPreview struct {
	ExecutionPolicy workflowstate.ExecutionPolicy `json:"execution_policy"`
	Preview         workflowcore.Preview          `json:"selection"`
	Target          workflowstate.PreviewTarget   `json:"target"`
	Policy          workflowstate.ReviewPolicy    `json:"publication_policy"`
}

// PreviewWorkExtent observes explicit native work without admission or fetching.
func (s service) PreviewWorkExtent(ctx context.Context, request workflowstate.PreviewRequest) (result WorkflowWorkPreview, err error) {
	s.ctx = ctx

	config, err := s.workflowExtentConfig()
	if err != nil {
		return result, err
	}

	// Validate selectors, filters and pagination before external target IO.
	preflight := request

	preflight.SnapshotHash = ""
	if _, err := s.PreviewWorkflow(ctx, preflight); err != nil {
		return result, err
	}

	identity, err := s.nativeAgentWorkTarget(config)
	if err != nil {
		return result, err
	}

	request.Target, err = s.previewWorkTarget(config, identity)
	if err != nil {
		return result, err
	}

	result.Target = request.Target

	result.Preview, err = s.PreviewWorkflow(ctx, request)
	if err != nil {
		return result, err
	}

	store, err := s.workflowStore()
	if err == nil {
		result.Policy, err = store.InspectWorkflowPolicy(ctx)
	}

	if err == nil {
		result.ExecutionPolicy, err = store.InspectWorkflowExecutionPolicy(ctx)
		if err == nil && result.Policy.Revision != result.ExecutionPolicy.ConfigRevision {
			err = workflowstate.ErrConflict
		}
	}

	return result, err
}

func (s service) previewWorkTarget(config workspaceconfig.Config, identity workflowstate.PreviewTarget) (workflowstate.PreviewTarget, error) {
	run := workflowcore.Invocation{InvocationInput: workflowcore.InvocationInput{Selection: workflowstate.PreviewRequest{Target: identity}}}

	reader, ok := s.workflowIntegrationObserver(run).(workflowcore.TargetReader)
	if !ok {
		return identity, fmt.Errorf("configured observer cannot read preview targets without fetching")
	}

	return reader.ReadWorkflowTarget(s.ctx, workflowcore.TargetQuery{Repository: identity.Repository, Directory: s.projectContext.ProjectDir, Branch: config.Work.BranchPolicy.IntegrationBranch})
}

func (s service) workflowExtentConfig() (workspaceconfig.Config, error) {
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err == nil && workspaceconfig.WorkAuthority(config.Work) != workspaceconfig.WorkStateAuthorityTickets {
		err = fmt.Errorf("explicit workflow extents require tickets authority; current authority is %s", workspaceconfig.WorkAuthority(config.Work))
	}

	return config, err
}

// StartWorkExtent delegates admission and bounded execution to the common owner.
func (s service) StartWorkExtent(ctx context.Context, request workflowstate.PreviewRequest, steps int, restrictions ...workflowstate.ReviewMode) (AgentWorkResult, error) {
	s.ctx = ctx

	restriction, err := workflowInvocationRestriction(restrictions)
	if err != nil {
		return AgentWorkResult{}, err
	}

	config, err := s.workflowExtentConfig()
	if err != nil {
		return AgentWorkResult{}, err
	}

	if steps < 0 || steps > workflowstate.PageLimit || request.Offset != 0 {
		return AgentWorkResult{}, fmt.Errorf("execution requires offset zero and 0..128 checkpoint steps")
	}

	request.Limit = workflowstate.PageLimit
	preflight := request

	preflight.SnapshotHash = ""
	if _, err := s.PreviewWorkflow(ctx, preflight); err != nil {
		return AgentWorkResult{}, err
	}

	return s.startNativeAgentWork(config, request, steps, restriction)
}

// ContinueWorkExtent retains the invocation's selector, target and budgets.
func (s service) ContinueWorkExtent(ctx context.Context, id string, steps int, expected ...uint64) (workflowcore.ContinuationProgress, error) {
	return s.continueWorkExtent(ctx, id, steps, nil, expected...)
}

func (s service) continueWorkExtent(ctx context.Context, id string, steps int, check workflowcore.AdmissionCheck, expected ...uint64) (workflowcore.ContinuationProgress, error) {
	request, err := s.workflowContinuationInput(steps)
	if err != nil {
		return workflowcore.ContinuationProgress{}, err
	}

	request.AuthorizationCheck = check

	if len(expected) > 1 || len(expected) == 1 && expected[0] == 0 {
		return workflowcore.ContinuationProgress{}, workflowstate.ErrConflict
	}

	if len(expected) == 1 {
		request.ExpectedRevision = expected[0]
	}

	return s.ContinueWorkflow(ctx, id, request, request.Worker)
}

func workflowInvocationRestriction(restrictions []workflowstate.ReviewMode) (workflowstate.ReviewMode, error) {
	if len(restrictions) == 0 {
		return "", nil
	}

	if len(restrictions) != 1 || restrictions[0] != "" && restrictions[0] != workflowstate.ReviewHuman {
		return "", fmt.Errorf("an invocation may restrict review to human_review only")
	}

	return restrictions[0], nil
}

func (s service) workflowContinuationInput(steps int) (workflowcore.ContinuationRequest, error) {
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return workflowcore.ContinuationRequest{}, err
	}

	worker := slices.Clone(s.workflowWorker)
	if worker == nil {
		worker = (codex.Runner{}).WorkflowWorker()
	}

	request := workflowcore.ContinuationRequest{Worker: worker, StepLimit: steps, PullRequest: workflowstate.PullRequestContent{BaseBranch: config.Work.BranchPolicy.IntegrationBranch, Title: "Deliver ticket-native work", Body: "Deliver the host-validated ticket candidate."}}

	return request, nil
}
