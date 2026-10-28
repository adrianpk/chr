package workflow

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *Store) checkWorkflowNativeSourceRevision(ctx context.Context, run *workflowInvocationFile, provider PullRequestProvider) error {
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

func workflowNativeSourceIdentityMatches(binding workflowstate.PullRequestBinding, pr workflowstate.PullRequestSnapshot) bool {
	observed := pr
	observed.HeadRepository, observed.HeadCommit = binding.Request.Repository, binding.SourceCommit

	return workflowforge.PullRequestMismatch(binding, []workflowstate.PullRequestSnapshot{observed}) == ""
}

func (s *Store) freezeWorkflowNativeSourceRevision(ctx context.Context, run *workflowInvocationFile, pr workflowstate.PullRequestSnapshot, head string) error {
	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return err
	}

	entry, err = s.resumeWorkflowNativeSourceDecision(ctx, *run, entry)
	if err != nil {
		return err
	}

	request := workflowNativeReplacementTransition(entry, "blocked")

	request.Blocker = &IntegrationBlocker{Reason: "source_revision_requires_reconciliation", RequiredAction: "verify_source_ownership_and_frozen_scope", ResumeState: "queued"}
	if err := s.checkWorkflowCandidateRepairTransition(ctx, *run, entry, request); err != nil {
		return err
	}

	// A new advertised source has no host task proof for the frozen scope.
	// Retain both observations for an explicit ownership/scope decision.
	observation := struct {
		PullRequest workflowstate.PullRequestSnapshot `json:"pull_request"`
		SourceHead  string                            `json:"source_head"`
	}{pr, head}

	preparationEvidence, err := workflowNativePreparationEvidenceFile(run.ID, observation)

	file, _ := preparationEvidence.Change, preparationEvidence.Ref

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

	updateWorkflowIntegrationBinding(run, workflowstate.IntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: queue.Path, Hash: contenthash.Sum(queue.After)}})

	return errors.Join(workflowstate.ErrConflict, s.commitWorkflowPullRequestFiles(run, []workflowstate.FileChange{queue}, file))
}

func (s *Store) resumeWorkflowNativeSourceDecision(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry) (IntegrationEntry, error) {
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
