package workflow

import (
	"context"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadWorkflowNativeProviderPolicy resolves current enabled host configuration before policy IO.
func (adapter RemoteReviewAdapter) ReadWorkflowNativeProviderPolicy(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowforge.NativePolicyObservation, error) {
	resolved, err := adapter.resolve(pr.URL)
	if err != nil {
		return workflowforge.NativePolicyObservation{}, err
	}

	client := adapter.workflowForgeClient(pr.Repository, resolved)

	return client.NativePolicy(ctx, pr, method)
}

// WorkflowNativeIntegrationCapability binds host observations to immutable forge policy.
func (adapter RemoteReviewAdapter) WorkflowNativeIntegrationCapability(ctx context.Context, request NativeCapabilityRequest, policies workflowforge.NativePolicyReader, git GitPublicationInspector) (workflowstate.NativeCapability, error) {
	return workflowforge.ObserveNativeCapability(ctx, workflowNativeCapabilityInput(request), adapter.nativeCapabilityReaders(request, policies, git))
}

// RecheckWorkflowNativeCapability re-observes the same current source and authority inputs.
func (adapter RemoteReviewAdapter) RecheckWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, prior workflowstate.NativeCapability, policies workflowforge.NativePolicyReader, git GitPublicationInspector) error {
	return workflowforge.RecheckNativeCapability(ctx, workflowNativeCapabilityInput(request), prior, adapter.nativeCapabilityReaders(request, policies, git))
}

func (adapter RemoteReviewAdapter) nativeCapabilityReaders(request NativeCapabilityRequest, policies workflowforge.NativePolicyReader, git GitPublicationInspector) workflowforge.NativeCapabilityReaders {
	readers := workflowforge.NativeCapabilityReaders{Provider: adapter, Policy: policies}
	if git != nil {
		readers.Transport = workflowNativeCapabilityTransport{Observer: git, Request: request.Publication}
	}

	return readers
}

type workflowNativeCapabilityTransport struct {
	Observer GitPublicationInspector
	Request  GitPublication
}

func (transport workflowNativeCapabilityTransport) InspectNativePublication(ctx context.Context) (workflowstate.GitPublicationAdvertisement, error) {
	return transport.Observer.InspectWorkflowGitPublication(ctx, transport.Request)
}
