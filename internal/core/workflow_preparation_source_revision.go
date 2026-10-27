package core

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *WorkflowStore) checkWorkflowNativeSourceRevision(ctx context.Context, run *workflowInvocationFile, provider WorkflowPullRequestProvider) error {
	config, err := workspaceconfig.Load(s.layout.ProjectDir())
	if err != nil || !workflowNativeSourceRevisionRequired(config, *run) {
		return err
	}

	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil || pr.Merged {
		return err
	}

	if !workflowNativeSourceIdentityMatches(*run.PullRequest, pr) {
		return workflowstate.ErrConflict
	}

	if pr.HeadRepository != run.PullRequest.Request.Repository {
		return s.freezeWorkflowNativeSourceRevision(ctx, run, pr, "")
	}

	head, err := provider.ReadWorkflowBranch(ctx, pr.HeadRepository, pr.HeadBranch)
	if err != nil {
		return err
	}

	if pr.HeadCommit != run.PullRequest.SourceCommit || head != run.PullRequest.SourceCommit {
		return s.freezeWorkflowNativeSourceRevision(ctx, run, pr, head)
	}

	return nil
}

func workflowNativeSourceRevisionRequired(config workspaceconfig.Config, run workflowInvocationFile) bool {
	return config.Work.Integration.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef && run.Integration != nil && run.PullRequest != nil
}

func workflowNativeSourceIdentityMatches(binding WorkflowPullRequestBinding, pr WorkflowPullRequestSnapshot) bool {
	observed := pr
	observed.HeadRepository, observed.HeadCommit = binding.Request.Repository, binding.SourceCommit

	return workflowPullRequestMismatch(binding, []WorkflowPullRequestSnapshot{observed}) == ""
}

func (s *WorkflowStore) freezeWorkflowNativeSourceRevision(ctx context.Context, run *workflowInvocationFile, pr WorkflowPullRequestSnapshot, head string) error {
	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return err
	}

	entry, err = s.resumeWorkflowNativeSourceDecision(ctx, *run, entry)
	if err != nil {
		return err
	}

	request := workflowNativeReplacementTransition(entry, "blocked")

	request.Blocker = &WorkflowIntegrationBlocker{Reason: "source_revision_requires_reconciliation", RequiredAction: "verify_source_ownership_and_frozen_scope", ResumeState: "queued"}
	if err := s.checkWorkflowCandidateRepairTransition(ctx, *run, entry, request); err != nil {
		return err
	}

	// A new advertised source has no host task proof for the frozen scope.
	// Retain both observations for an explicit ownership/scope decision.
	observation := struct {
		PullRequest WorkflowPullRequestSnapshot `json:"pull_request"`
		SourceHead  string                      `json:"source_head"`
	}{pr, head}

	file, _, err := workflowNativePreparationEvidenceFile(run.ID, observation)
	if err != nil {
		return err
	}

	updateWorkflowIntegrationEntry(&entry, request)
	staleWorkflowNativePreparation(run)
	projectWorkflowIntegrationPhase(run, entry)

	run.Outcome = request.Blocker.Reason

	queue, err := workflowIntegrationEntryChange(entry, run.Integration.EntryRef.Hash)
	if err != nil {
		return err
	}

	updateWorkflowIntegrationBinding(run, WorkflowIntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: queue.Path, Hash: contenthash.Sum(queue.After)}})

	return errors.Join(workflowstate.ErrConflict, s.commitWorkflowPullRequestFiles(run, []workflowstate.FileChange{queue}, file))
}

func (s *WorkflowStore) resumeWorkflowNativeSourceDecision(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry) (WorkflowIntegrationEntry, error) {
	if entry.State != "blocked" || entry.Blocker.Reason != "native_not_applied" {
		return entry, nil
	}

	request := workflowNativeReplacementTransition(entry, "queued")
	if err := s.checkWorkflowCandidateRepairTransition(ctx, run, entry, request); err != nil {
		return entry, err
	}

	updateWorkflowIntegrationEntry(&entry, request)

	return entry, nil
}
