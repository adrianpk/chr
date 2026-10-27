package app

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

func nativeAgentWorkSelection(opts AgentWorkSelectionOptions) workflowstate.WorkflowPreviewRequest {
	return workflowstate.WorkflowPreviewRequest{Selector: workflowstate.WorkflowSelector{Extent: "loose"}, Statuses: agentWorkStatuses(opts.Statuses), Filters: opts.Filters, Order: agentWorkOrder(opts.Order), Limit: workflowstate.PageLimit}
}

func legacyAgentWorkResult(result AgentWorkResult) AgentWorkResult {
	result.Scope = "unadopted_legacy"
	result.Authority = core.WorkStateAuthorityTrackerCompat
	result.Message = fmt.Sprintf("Authority tracker_compat; scope unadopted legacy tickets; %d adopted tickets excluded. %s", result.Excluded["ticket_native"], result.Message)

	return result
}

func (s service) checkLegacyAgentWorkKey(key string) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	return store.CheckLegacyWorkKey(s.ctx, key)
}

func (s service) nativeAgentWorkTarget(config core.WorkspaceConfig) (workflowstate.WorkflowPreviewTarget, error) {
	if config.Work.BranchPolicy.Mode != core.WorkBranchPolicyModeIntegrationBranch {
		return workflowstate.WorkflowPreviewTarget{}, fmt.Errorf("ticket authority requires an explicit integration branch")
	}

	repository, err := core.WorkflowEntryRepository(s.ctx, s.projectContext.ProjectDir, config.Work.BranchPolicy.CanonicalRemote)

	return workflowstate.WorkflowPreviewTarget{Repository: repository}, err
}

func (s service) observeNativeAgentWorkTarget(config core.WorkspaceConfig, identity workflowstate.WorkflowPreviewTarget) (workflowstate.WorkflowPreviewTarget, error) {
	run := core.WorkflowInvocation{WorkflowInvocationInput: core.WorkflowInvocationInput{Selection: workflowstate.WorkflowPreviewRequest{Target: identity}}}

	observer, ok := s.workflowIntegrationObserver(run).(core.WorkflowTargetGit)
	if !ok {
		return identity, fmt.Errorf("configured observer cannot hydrate entrypoint targets")
	}

	return observer.ObserveWorkflowTarget(s.ctx, core.WorkflowTargetQuery{Repository: identity.Repository, Directory: s.projectContext.ProjectDir, Branch: config.Work.BranchPolicy.IntegrationBranch})
}

func (s service) nativeAgentWorkStart(config core.WorkspaceConfig, selection workflowstate.WorkflowPreviewRequest, restriction workflowstate.ReviewWorkflowMode, selected ...*workflowstate.WorkflowSelectionGuard) (core.WorkflowStartResult, error) {
	target, err := s.nativeAgentWorkTarget(config)
	if err != nil {
		return core.WorkflowStartResult{}, err
	}

	selection.Target = target
	request := core.WorkflowStartRequest{RequestKey: "work-entry-" + strconv.FormatInt(time.Now().UnixNano(), 10), Selection: selection, ReviewRestriction: restriction}

	if len(selected) == 1 {
		request.Selected = selected[0]

		request.Selection.Target, err = s.observeNativeAgentWorkTarget(config, target)
		if err != nil {
			return core.WorkflowStartResult{}, err
		}

		return s.StartWorkflow(s.ctx, request)
	}

	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowStartResult{}, err
	}

	previous, found, err := store.FindWorkflowEntry(s.ctx, request)
	if err != nil {
		return core.WorkflowStartResult{}, err
	}

	if found {
		request, err = retainedNativeAgentWorkRequest(request, previous)
	} else {
		request.Selection.Target, err = s.observeNativeAgentWorkTarget(config, target)
	}

	if err != nil {
		return core.WorkflowStartResult{}, err
	}

	return s.StartWorkflow(s.ctx, request)
}

func retainedNativeAgentWorkRequest(request core.WorkflowStartRequest, previous core.WorkflowInvocation) (core.WorkflowStartRequest, error) {
	if request.ReviewRestriction != "" && request.ReviewRestriction != previous.ReviewRestriction {
		return request, fmt.Errorf("%w: retained invocation has a different review restriction", workflowstate.ErrConflict)
	}

	request.RequestKey, request.Selection = previous.RequestKey, previous.Selection
	request.Admission = previous.Admission
	request.HarnessBinding = previous.HarnessBinding
	request.Policy, request.ReviewRestriction = previous.Policy, previous.ReviewRestriction

	return request, nil
}

func (s service) startNativeAgentWork(config core.WorkspaceConfig, selection workflowstate.WorkflowPreviewRequest, limit int, restriction workflowstate.ReviewWorkflowMode, selected ...*workflowstate.WorkflowSelectionGuard) (AgentWorkResult, error) {
	if limit < 0 || limit > workflowstate.PageLimit {
		return AgentWorkResult{}, fmt.Errorf("ticket authority max-iterations must contain 0..128 checkpoint steps")
	}

	started, err := s.nativeAgentWorkStart(config, selection, restriction, selected...)
	if err != nil {
		return AgentWorkResult{}, err
	}

	worker := slices.Clone(s.workflowWorker)
	if worker == nil {
		worker = (codex.Runner{}).WorkflowWorker()
	}

	request := core.WorkflowContinuationRequest{Worker: worker, StepLimit: limit, PullRequest: core.WorkflowSlicePullRequestRequest{BaseBranch: config.Work.BranchPolicy.IntegrationBranch, Title: "Deliver ticket-native work", Body: "Deliver the host-validated ticket candidate."}}

	progress, err := s.ContinueWorkflow(s.ctx, started.Invocation.ID, request, worker)
	if progress.Invocation.ID == "" {
		progress.Invocation = started.Invocation
	}

	result, displayErr := s.nativeAgentWorkResult(progress.Invocation)
	if err != nil {
		return result, fmt.Errorf("workflow %s: %w", started.Invocation.ID, err)
	}

	result.Message += fmt.Sprintf(" Accepted steps: %d; checkpoint: %s.", progress.Steps, progress.Reason)

	return result, displayErr
}

func (s service) nativeAgentWorkResult(run core.WorkflowInvocation) (AgentWorkResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return AgentWorkResult{}, err
	}

	tickets, err := store.WorkflowTickets(s.ctx)

	return AgentWorkResult{Scope: run.Selection.Selector.Extent, Authority: core.WorkStateAuthorityTickets, Workflow: &run, Excluded: run.Excluded, Tickets: tickets, EligibleCount: len(run.Members), Message: fmt.Sprintf("Authority tickets; workflow %s is %s; confirmed %d of %d admitted tickets; %d slice-owned tickets excluded.", run.ID, run.Status, run.ConfirmedCursor, len(run.Members), run.Excluded["slice_owned"])}, err
}
