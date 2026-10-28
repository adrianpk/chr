package workflow

import (
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
)

func workflowPublicationIdentity(request GitPublication) workflowforge.PublicationIdentity {
	return workflowforge.PublicationIdentity{Repository: request.Repository, SourceRepository: request.SourceRepository, SourceRef: request.SourceRef, TargetRef: request.TargetRef, OwnershipRevision: request.OwnershipRevision, SourceRewriteAuthorized: request.SourceRewriteAuthorized, Git: request.Git}
}

func workflowNativeCapabilityInput(request NativeCapabilityRequest) workflowforge.NativeCapabilityInput {
	return workflowforge.NativeCapabilityInput{Profile: request.Config.PublicationProfile, PullRequest: request.PullRequest, Publication: workflowPublicationIdentity(request.Publication)}
}
