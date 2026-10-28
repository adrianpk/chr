package workflow

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type nativePolicyFixture struct {
	policy workflowstate.NativePolicy
	calls  int
}

func (fixture *nativePolicyFixture) ReadWorkflowNativePolicy(_ context.Context, _ workflowstate.PullRequestSnapshot, _ string) (workflowstate.NativePolicy, error) {
	fixture.calls++

	return fixture.policy, nil
}

type nativeInspectorFixture struct {
	GitPublicationInspector
	calls int
}

func (fixture *nativeInspectorFixture) InspectWorkflowGitPublication(ctx context.Context, request GitPublication) (workflowstate.GitPublicationAdvertisement, error) {
	fixture.calls++

	return fixture.GitPublicationInspector.InspectWorkflowGitPublication(ctx, request)
}

type nativeCapabilityFixture struct {
	adapter RemoteReviewAdapter
	runner  *workflowPRCommandFixture
	request NativeCapabilityRequest
	policy  *nativePolicyFixture
	git     *nativeInspectorFixture
	remote  string
	prefix  string
}

func newNativeCapabilityFixture(t *testing.T, forge string) nativeCapabilityFixture {
	t.Helper()

	fixture := newWorkflowGitFixture(t, "rebase")

	repository, marker, contract := fixture.request.Repository, "/pulls/", "forgejo_same_pr_git_auto_merge/v1"
	if forge == "github" {
		repository, marker, contract = "https://github.com/team/project", "/pull/", "github_same_pr_git_merge/v1"
	}

	publication := fixture.request
	publication.Repository, publication.SourceRepository, publication.Git.Repository = repository, repository, repository
	sealWorkflowGitFixture(t, &publication)

	pr := workflowstate.PullRequestSnapshot{URL: repository + marker + "7", Repository: repository, HeadRepository: repository, HeadBranch: strings.TrimPrefix(publication.SourceRef, "refs/heads/"), BaseBranch: "dev", HeadCommit: publication.Git.SourceCommit, BaseCommit: publication.Git.BaseCommit, State: "open"}
	runner := &workflowPRCommandFixture{fakeRemoteReviewRunner: newFakeRemoteReviewRunner()}
	adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)
	policy := workflowstate.NativePolicy{Repository: repository, SourceRef: publication.SourceRef, TargetRef: publication.TargetRef, OwnershipRevision: publication.OwnershipRevision, SourceRewriteAuthorized: true, FastForwardTargetAuthorized: true, HostCandidateReview: true, CredentialIdentityFingerprint: contenthash.Sum([]byte("fixture-role")), ObservationContract: contract, CertifiedMethods: []string{"rebase"}}
	policy.ReviewMode = workflowstate.ReviewHuman
	result := nativeCapabilityFixture{adapter: adapter, runner: runner, request: NativeCapabilityRequest{Config: workspaceconfig.WorkIntegrationConfig{PublicationProfile: workspaceconfig.WorkflowPublicationGitAtomicDualRef}, PullRequest: pr, Publication: publication}, policy: &nativePolicyFixture{policy: policy}, git: &nativeInspectorFixture{GitPublicationInspector: NewWorkflowGitPublisher(repository, fixture.remote)}, remote: fixture.remote}
	result.prefix = nativePolicyCommand(t, result, "")

	resolvedReview, err := adapter.resolve(pr.URL)
	if err != nil {
		t.Fatal(err)
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	runner.outputs[remoteReviewCommandKey(config.CLI, workflowForgeEndpoint(repository, reference, config).Arguments(workflowforge.Request{Method: "GET", Path: "/user"})...)] = []byte(`{"id":1,"login":"fixture-publisher"}`)

	setNativePRFixture(t, result, pr)

	runner.outputs[result.prefix] = []byte(`{"html_url":"` + repository + `","permissions":{"push":true},"allow_rebase":true,"allow_rebase_merge":true}`)
	runner.outputs[result.prefix+"/branches/dev"] = []byte(`{"protected":false}`)

	runner.outputs[result.prefix+"/branches/chr%2Fworkflow%2Ffixture"] = []byte(`{"protected":false}`)
	if forge == "github" {
		runner.outputs[result.prefix+"/rules/branches/dev"] = []byte(`[]`)
		runner.outputs[result.prefix+"/rules/branches/chr%2Fworkflow%2Ffixture"] = []byte(`[]`)
	}

	return result
}

func nativePolicyCommand(t *testing.T, fixture nativeCapabilityFixture, suffix string) string {
	t.Helper()

	resolvedReview, err := fixture.adapter.resolve(fixture.request.PullRequest.URL)
	if err != nil {
		t.Fatal(err)
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	return remoteReviewCommandKey(config.CLI, workflowForgeEndpoint(fixture.request.PullRequest.Repository, reference, config).Arguments(workflowforge.Request{Method: "GET", Path: workflowForgeEndpoint(fixture.request.PullRequest.Repository, reference, config).RepositoryPath() + suffix})...)
}

func setNativePRFixture(t *testing.T, fixture nativeCapabilityFixture, pr workflowstate.PullRequestSnapshot) {
	t.Helper()

	remote := workflowPRWireFixture{URL: pr.URL, State: pr.State, Merged: pr.Merged, ResultCommit: pr.ResultCommit}
	remote.Head.Ref, remote.Head.SHA, remote.Head.Repo.URL = pr.HeadBranch, pr.HeadCommit, pr.HeadRepository
	remote.Base.Ref, remote.Base.SHA, remote.Base.Repo.URL = pr.BaseBranch, pr.BaseCommit, pr.Repository

	data, err := json.Marshal(remote)
	if err != nil {
		t.Fatal(err)
	}

	fixture.runner.outputs[nativePolicyCommand(t, fixture, "/pulls/7")] = data
}

// TestNativeCapability checks real read-only receive-pack admission for both
// existing adapters, while retaining S/B/P and invalidating current role drift.
func TestNativeCapability(t *testing.T) {
	for _, forge := range []string{"forgejo", "github"} {
		t.Run(forge, func(t *testing.T) {
			fixture := newNativeCapabilityFixture(t, forge)
			capability, err := fixture.adapter.WorkflowNativeIntegrationCapability(t.Context(), fixture.request, fixture.policy, fixture.git)
			checkNativeCapability(t, fixture, capability, err)

			if err := fixture.adapter.RecheckWorkflowNativeCapability(t.Context(), fixture.request, capability, fixture.policy, fixture.git); err != nil {
				t.Fatal(err)
			}

			preview := fixture.request.PullRequest
			preview.ResultCommit = strings.Repeat("e", 40)
			setNativePRFixture(t, fixture, preview)

			if err := fixture.adapter.RecheckWorkflowNativeCapability(t.Context(), fixture.request, capability, fixture.policy, fixture.git); err != nil {
				t.Fatal("unmerged speculative result invalidated unchanged S/B/P", err)
			}

			fixture.policy.policy.CredentialIdentityFingerprint = contenthash.Sum([]byte("changed-role"))
			if err := fixture.adapter.RecheckWorkflowNativeCapability(t.Context(), fixture.request, capability, fixture.policy, fixture.git); err == nil {
				t.Fatal("changed credential identity reused old authorization")
			}

			checkWorkflowGitRefs(t, workflowGitFixture{request: fixture.request.Publication, remote: fixture.remote}, fixture.request.Publication.Git.SourceCommit, fixture.request.Publication.Git.BaseCommit)
		})
	}
}

func checkNativeCapability(t *testing.T, fixture nativeCapabilityFixture, capability workflowstate.NativeCapability, err error) {
	t.Helper()

	if err != nil || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 || capability.OriginalSourceCommit != fixture.request.Publication.Git.SourceCommit || capability.PreparedResultCommit != fixture.request.Publication.Git.ResultCommit {
		t.Fatalf("native capability: %+v, %v", capability, err)
	}
}

// TestNativeCapabilityDenials blocks policy combinations before opening a
// receive-pack exchange; no unsupported path invokes an unguarded merge API.
func TestNativeCapabilityDenials(t *testing.T) {
	cases := []struct {
		name   string
		change func(*nativeCapabilityFixture)
	}{
		{"disabled", func(f *nativeCapabilityFixture) {
			f.request.Config.PublicationProfile = workspaceconfig.WorkflowPublicationDisabled
		}},
		{"human_source", func(f *nativeCapabilityFixture) { f.policy.policy.SourceRewriteAuthorized = false }},
		{"ownership_changed", func(f *nativeCapabilityFixture) { f.policy.policy.OwnershipRevision = "changed" }},
		{"fork_source", func(f *nativeCapabilityFixture) {
			f.request.Publication.SourceRepository = "https://github.com/fork/project"
			f.request.PullRequest.HeadRepository = f.request.Publication.SourceRepository
		}},
		{"native_approval", func(f *nativeCapabilityFixture) { f.policy.policy.HostCandidateReview = false }},
		{"no_target_grant", func(f *nativeCapabilityFixture) { f.policy.policy.FastForwardTargetAuthorized = false }},
		{"unknown_identity", func(f *nativeCapabilityFixture) { f.policy.policy.CredentialIdentityFingerprint = "" }},
		{"uncertified_method", func(f *nativeCapabilityFixture) { f.policy.policy.CertifiedMethods = nil }},
		{"missing_observer", func(f *nativeCapabilityFixture) { f.policy.policy.ObservationContract = "" }},
		{"protected_target", func(f *nativeCapabilityFixture) {
			f.runner.outputs[f.prefix+"/branches/dev"] = []byte(`{"protected":true}`)
		}},
		{"native_rules", func(f *nativeCapabilityFixture) {
			f.runner.outputs[f.prefix+"/rules/branches/dev"] = []byte(`[{"type":"pull_request"}]`)
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newNativeCapabilityFixture(t, "github")
			test.change(&fixture)
			setNativePRFixture(t, fixture, fixture.request.PullRequest)

			capability, err := fixture.adapter.WorkflowNativeIntegrationCapability(t.Context(), fixture.request, fixture.policy, fixture.git)
			if err != nil || len(capability.Unsupported) == 0 || capability.AtomicSourceBase || capability.ExactResult || fixture.git.calls != 0 {
				t.Fatalf("unsafe admission: %+v, %v, transport calls %d", capability, err, fixture.git.calls)
			}
		})
	}
}
