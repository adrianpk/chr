package forge

import (
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PublicationIdentity binds immutable source, target, ownership and prepared proof.
type PublicationIdentity struct {
	Repository              string
	SourceRepository        string
	SourceRef               string
	TargetRef               string
	OwnershipRevision       string
	SourceRewriteAuthorized bool
	Git                     workflowstate.IntegrationProof
}

// NativeCapabilityInput supplies a selected profile and exact immutable source.
type NativeCapabilityInput struct {
	Profile     string
	PullRequest workflowstate.PullRequestSnapshot
	Publication PublicationIdentity
}

// PublicationRevision binds the original source, base, exact result,
// ownership grant, full refs and method/history proof. Directory is host-local IO.
func PublicationRevision(request PublicationIdentity) (string, error) {
	data, err := json.Marshal(struct {
		Schema, Repository, SourceRepository, SourceRef, TargetRef, OwnershipRevision string
		SourceRewriteAuthorized                                                       bool
		Git                                                                           workflowstate.IntegrationProof
	}{"chr-git-publication/v1", request.Repository, request.SourceRepository, request.SourceRef, request.TargetRef, request.OwnershipRevision, request.SourceRewriteAuthorized, request.Git})

	return contenthash.Sum(data), err
}
