package app

import (
	"context"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

// WorkflowWorkPreview reports the exact scope and current publication policy.
type WorkflowWorkPreview struct {
	ExecutionPolicy core.WorkflowExecutionPolicy `json:"execution_policy"`
	Preview         core.WorkflowPreview         `json:"selection"`
	Target          core.WorkflowPreviewTarget   `json:"target"`
	Policy          core.WorkflowReviewPolicy    `json:"publication_policy"`
}

// PreviewWorkExtent observes explicit native work without admission or fetching.
func (s service) PreviewWorkExtent(ctx context.Context, request core.WorkflowPreviewRequest) (result WorkflowWorkPreview, err error) {
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

func (s service) previewWorkTarget(config core.WorkspaceConfig, identity core.WorkflowPreviewTarget) (core.WorkflowPreviewTarget, error) {
	run := core.WorkflowInvocation{WorkflowInvocationInput: core.WorkflowInvocationInput{Selection: core.WorkflowPreviewRequest{Target: identity}}}

	reader, ok := s.workflowIntegrationObserver(run).(core.WorkflowTargetReader)
	if !ok {
		return identity, fmt.Errorf("configured observer cannot read preview targets without fetching")
	}

	return reader.ReadWorkflowTarget(s.ctx, core.WorkflowTargetQuery{Repository: identity.Repository, Directory: s.projectContext.ProjectDir, Branch: config.Work.BranchPolicy.IntegrationBranch})
}

func (s service) workflowExtentConfig() (core.WorkspaceConfig, error) {
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err == nil && core.WorkAuthority(config.Work) != core.WorkStateAuthorityTickets {
		err = fmt.Errorf("explicit workflow extents require tickets authority; current authority is %s", core.WorkAuthority(config.Work))
	}

	return config, err
}

// StartWorkExtent delegates admission and bounded execution to the common owner.
func (s service) StartWorkExtent(ctx context.Context, request core.WorkflowPreviewRequest, steps int, restrictions ...core.ReviewWorkflowMode) (AgentWorkResult, error) {
	s.ctx = ctx

	restriction, err := workflowInvocationRestriction(restrictions)
	if err != nil {
		return AgentWorkResult{}, err
	}

	config, err := s.workflowExtentConfig()
	if err != nil {
		return AgentWorkResult{}, err
	}

	if steps < 0 || steps > core.WorkflowPageLimit || request.Offset != 0 {
		return AgentWorkResult{}, fmt.Errorf("execution requires offset zero and 0..128 checkpoint steps")
	}

	request.Limit = core.WorkflowPageLimit
	preflight := request

	preflight.SnapshotHash = ""
	if _, err := s.PreviewWorkflow(ctx, preflight); err != nil {
		return AgentWorkResult{}, err
	}

	return s.startNativeAgentWork(config, request, steps, restriction)
}

// ContinueWorkExtent retains the invocation's selector, target and budgets.
func (s service) ContinueWorkExtent(ctx context.Context, id string, steps int, expected ...uint64) (core.WorkflowContinuationProgress, error) {
	return s.continueWorkExtent(ctx, id, steps, nil, expected...)
}

func (s service) continueWorkExtent(ctx context.Context, id string, steps int, check core.WorkflowAdmissionCheck, expected ...uint64) (core.WorkflowContinuationProgress, error) {
	request, err := s.workflowContinuationInput(steps)
	if err != nil {
		return core.WorkflowContinuationProgress{}, err
	}

	request.AuthorizationCheck = check

	if len(expected) > 1 || len(expected) == 1 && expected[0] == 0 {
		return core.WorkflowContinuationProgress{}, workflowstate.ErrConflict
	}

	if len(expected) == 1 {
		request.ExpectedRevision = expected[0]
	}

	return s.ContinueWorkflow(ctx, id, request, request.Worker)
}

func workflowInvocationRestriction(restrictions []core.ReviewWorkflowMode) (core.ReviewWorkflowMode, error) {
	if len(restrictions) == 0 {
		return "", nil
	}

	if len(restrictions) != 1 || restrictions[0] != "" && restrictions[0] != core.ReviewWorkflowModeHumanReview {
		return "", fmt.Errorf("an invocation may restrict review to human_review only")
	}

	return restrictions[0], nil
}

func (s service) workflowContinuationInput(steps int) (core.WorkflowContinuationRequest, error) {
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.WorkflowContinuationRequest{}, err
	}

	worker := slices.Clone(s.workflowWorker)
	if worker == nil {
		worker = (codex.Runner{}).WorkflowWorker()
	}

	request := core.WorkflowContinuationRequest{Worker: worker, StepLimit: steps, PullRequest: core.WorkflowSlicePullRequestRequest{BaseBranch: config.Work.BranchPolicy.IntegrationBranch, Title: "Deliver ticket-native work", Body: "Deliver the host-validated ticket candidate."}}

	return request, nil
}
