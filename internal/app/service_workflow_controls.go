package app

import (
	"context"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkflowWorkPreview reports the exact scope and current publication policy.
type WorkflowWorkPreview struct {
	Preview core.WorkflowPreview       `json:"selection"`
	Target  core.WorkflowPreviewTarget `json:"target"`
	Policy  core.WorkflowReviewPolicy  `json:"publication_policy"`
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

	return result, err
}

func (s service) previewWorkTarget(config core.WorkspaceConfig, identity core.WorkflowPreviewTarget) (core.WorkflowPreviewTarget, error) {
	run := core.WorkflowInvocation{Selection: core.WorkflowPreviewRequest{Target: identity}}

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
func (s service) StartWorkExtent(ctx context.Context, request core.WorkflowPreviewRequest, steps int) (AgentWorkResult, error) {
	s.ctx = ctx

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

	return s.startNativeAgentWork(config, request, steps)
}

// ContinueWorkExtent retains the invocation's selector, target and budgets.
func (s service) ContinueWorkExtent(ctx context.Context, id string, steps int) (core.WorkflowContinuationProgress, error) {
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.WorkflowContinuationProgress{}, err
	}

	worker := slices.Clone(s.workflowWorker)
	if worker == nil {
		worker = core.CodexWorkflowWorker()
	}

	request := core.WorkflowContinuationRequest{Worker: worker, StepLimit: steps, PullRequest: core.WorkflowSlicePullRequestRequest{BaseBranch: config.Work.BranchPolicy.IntegrationBranch, Title: "Deliver ticket-native work", Body: "Deliver the host-validated ticket candidate."}}

	return s.ContinueWorkflow(ctx, id, request, worker)
}
