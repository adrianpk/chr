package workflow

import (
	"context"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (git OSWorkflowGitPublisher) validatePublicationProof(ctx context.Context, request GitPublication) error {
	if err := validateWorkflowPublicationOwnership(request); err != nil {
		return err
	}

	proof := request.Git
	if proof.Repository != request.Repository || request.TargetRef != "refs/heads/"+proof.TargetBranch {
		return workflowstate.ErrConflict
	}

	if err := workflowintegrate.CheckProof(proof.IntegrationQuery, proof, workflowIntegrationRepairs{}); err != nil {
		return err
	}

	if !validWorkflowPublicationProofObjects(proof) {
		return workflowstate.ErrConflict
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: request.Directory}}

	tree, err := reader.tree(ctx, proof.SourceCommit)
	if err != nil || tree != proof.SourceTree {
		return workflowstate.ErrConflict
	}

	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.SourceCommit); err != nil {
		return workflowstate.ErrConflict
	}

	return reader.verifyPublicationHistory(ctx, proof)
}

func validWorkflowPublicationProofObjects(proof workflowstate.IntegrationProof) bool {
	for _, id := range []string{proof.SourceTree, proof.OriginCommit, proof.PreparedTree} {
		if !workflowstate.ValidCommitID(id) || len(id) != len(proof.ResultCommit) {
			return false
		}
	}

	return true
}

func validateWorkflowPublicationOwnership(request GitPublication) error {
	source, err := workflowstate.CanonicalRepository(request.SourceRepository)
	if err != nil || source != request.Repository || !request.SourceRewriteAuthorized {
		return workflowstate.ErrConflict
	}

	if request.OwnershipRevision == "" || len(request.OwnershipRevision) > 256 || strings.ContainsAny(request.OwnershipRevision, "\x00\r\n") {
		return workflowstate.ErrConflict
	}

	revision, err := workflowforge.PublicationRevision(workflowPublicationIdentity(request))
	if err != nil || revision != request.CandidateRevision {
		return workflowstate.ErrConflict
	}

	return nil
}

func (reader workflowIntegrationReader) verifyPublicationHistory(ctx context.Context, proof workflowstate.IntegrationProof) error {
	actual := proof

	actual.History = nil
	if err := reader.verifyMethod(ctx, &actual); err != nil {
		return err
	}

	hash, err := workflowintegrate.HistoryHash(actual)
	if err != nil || hash != proof.HistoryHash || actual.BaseCommit != proof.BaseCommit || actual.VerifiedTree != proof.VerifiedTree {
		return workflowstate.ErrConflict
	}

	return nil
}
