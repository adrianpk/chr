package core

import (
	"context"
	"encoding/json"
	"strings"
)

// WorkflowGitPublicationRevision binds the original source, base, exact result,
// ownership grant, full refs and method/history proof. Directory is host-local IO.
func WorkflowGitPublicationRevision(request WorkflowGitPublication) (string, error) {
	data, err := json.Marshal(struct {
		Schema, Repository, SourceRepository, SourceRef, TargetRef, OwnershipRevision string
		SourceRewriteAuthorized                                                       bool
		Git                                                                           WorkflowIntegrationProof
	}{"chr-git-publication/v1", request.Repository, request.SourceRepository, request.SourceRef, request.TargetRef, request.OwnershipRevision, request.SourceRewriteAuthorized, request.Git})

	return workflowHash(data), err
}

func (git OSWorkflowGitPublisher) validatePublicationProof(ctx context.Context, request WorkflowGitPublication) error {
	if err := validateWorkflowPublicationOwnership(request); err != nil {
		return err
	}

	proof := request.Git
	if proof.Repository != request.Repository || request.TargetRef != "refs/heads/"+proof.TargetBranch {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof); err != nil {
		return err
	}

	if !validWorkflowPublicationProofObjects(proof) {
		return ErrWorkflowConflict
	}

	reader := workflowIntegrationReader{dir: request.Directory}

	tree, err := reader.tree(ctx, proof.SourceCommit)
	if err != nil || tree != proof.SourceTree {
		return ErrWorkflowConflict
	}

	if _, err := reader.output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.SourceCommit); err != nil {
		return ErrWorkflowConflict
	}

	return reader.verifyPublicationHistory(ctx, proof)
}

func validWorkflowPublicationProofObjects(proof WorkflowIntegrationProof) bool {
	for _, id := range []string{proof.SourceTree, proof.OriginCommit, proof.PreparedTree} {
		if !workflowCommitID(id) || len(id) != len(proof.ResultCommit) {
			return false
		}
	}

	return true
}

func validateWorkflowPublicationOwnership(request WorkflowGitPublication) error {
	source, err := workflowCanonicalRepository(request.SourceRepository)
	if err != nil || source != request.Repository || !request.SourceRewriteAuthorized {
		return ErrWorkflowConflict
	}

	if request.OwnershipRevision == "" || len(request.OwnershipRevision) > 256 || strings.ContainsAny(request.OwnershipRevision, "\x00\r\n") {
		return ErrWorkflowConflict
	}

	revision, err := WorkflowGitPublicationRevision(request)
	if err != nil || revision != request.CandidateRevision {
		return ErrWorkflowConflict
	}

	return nil
}

func (reader workflowIntegrationReader) verifyPublicationHistory(ctx context.Context, proof WorkflowIntegrationProof) error {
	actual := proof

	actual.History = nil
	if err := reader.verifyMethod(ctx, &actual); err != nil {
		return err
	}

	hash, err := workflowIntegrationHistoryHash(actual)
	if err != nil || hash != proof.HistoryHash || actual.BaseCommit != proof.BaseCommit || actual.VerifiedTree != proof.VerifiedTree {
		return ErrWorkflowConflict
	}

	return nil
}
