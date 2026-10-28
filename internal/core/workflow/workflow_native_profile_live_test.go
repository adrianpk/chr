package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type nativeLiveEvidence struct {
	Profile      string                                       `json:"profile"`
	Binding      workflowstate.NativePublicationBinding       `json:"binding"`
	Observations []workflowstate.NativePublicationObservation `json:"observations"`
	Failure      string                                       `json:"failure,omitempty"`
}

// TestWorkflowNativeRebase exercises the real publisher and existing adapters
// on the two explicitly designated private fixtures. Opt-in creates retained
// synthetic refs/PRs; neither a native merge API nor a provider queue is used.
func TestWorkflowNativeRebase(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live native integration fixtures require explicit opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		t.Run(profile.Forge, func(t *testing.T) { nativeLiveRebase(t, state, profile) })
	}
}

func nativeLiveRebase(t *testing.T, state string, profile nativeLiveProfile) {
	t.Helper()

	request, key := prepareNativeLiveRebase(t, profile)
	adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), OSReviewRemoteCommandRunner{})
	pr := createNativeLivePR(t, adapter, profile, request)
	publisher := NewWorkflowGitPublisher(profile.URL, nativeLiveTransport(t, profile))
	capability := admitNativeLiveProfile(t, state, adapter, profile, request, pr, publisher)
	result, publicationErr := publisher.PublishWorkflowGit(t.Context(), request)
	binding := workflowstate.NativePublicationBinding{Capability: capability, PullRequest: pr, Git: request.Git, Publication: result, SourceDirectory: request.Directory, ObservationDeadline: time.Now().Add(2 * time.Minute)}

	evidence := nativeLiveEvidence{Profile: profile.Forge, Binding: binding}
	if publicationErr != nil {
		evidence.Failure = publicationErr.Error()
	}

	defer writeNativeLiveEvidence(t, state, key, &evidence)

	if !result.CommandsSent {
		t.Fatalf("publication not sent: %v", publicationErr)
	}

	observeNativeLiveRebase(t, adapter, profile, &evidence)

	if publicationErr != nil || result.State != "applied" {
		t.Fatalf("transport requires reconciliation despite observed delivery: %s, %v", result.State, publicationErr)
	}
}

func createNativeLivePR(t *testing.T, adapter RemoteReviewAdapter, profile nativeLiveProfile, request GitPublication) workflowstate.PullRequestSnapshot {
	t.Helper()

	pr, err := adapter.CreateWorkflowPullRequest(t.Context(), workflowstate.PullRequestRequest{
		PullRequestQuery: workflowstate.PullRequestQuery{Repository: profile.URL, HeadBranch: strings.TrimPrefix(request.SourceRef, "refs/heads/"), BaseBranch: request.Git.TargetBranch},
		Title:            "test(integration): certify native atomic " + request.Git.Method, Body: "Synthetic owned-source fixture for exact-result publication and same-original-PR observation.",
	})
	if err != nil || pr.HeadCommit != request.Git.SourceCommit || pr.BaseCommit != request.Git.BaseCommit || pr.State != "open" {
		t.Fatalf("canonical fixture PR creation: %+v, %v", pr, err)
	}

	nativeLiveActivity(t, adapter, profile, pr, "opened")
	t.Logf("canonical PR %s, original source %s, base %s, prepared result %s", pr.URL, request.Git.SourceCommit, request.Git.BaseCommit, request.Git.ResultCommit)

	return pr
}

func admitNativeLiveProfile(t *testing.T, state string, adapter RemoteReviewAdapter, profile nativeLiveProfile, request GitPublication, pr workflowstate.PullRequestSnapshot, publisher OSWorkflowGitPublisher) workflowstate.NativeCapability {
	t.Helper()

	policy := nativeLivePolicyReader{state: state, profile: profile, publication: request}
	query := NativeCapabilityRequest{Config: workspaceconfig.WorkIntegrationConfig{PublicationProfile: workspaceconfig.WorkflowPublicationGitAtomicDualRef}, PullRequest: pr, Publication: request}

	capability, err := adapter.WorkflowNativeIntegrationCapability(t.Context(), query, policy, publisher)
	if err != nil || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 {
		t.Fatalf("live native capability: %+v, %v", capability, err)
	}

	if err := adapter.RecheckWorkflowNativeCapability(t.Context(), query, capability, policy, publisher); err != nil {
		t.Fatal(err)
	}

	return capability
}

func prepareNativeLiveRebase(t *testing.T, profile nativeLiveProfile) (GitPublication, string) {
	t.Helper()

	return prepareNativeLiveMethod(t, profile, "rebase")
}

func prepareNativeLiveMethod(t *testing.T, profile nativeLiveProfile, method string) (GitPublication, string) {
	t.Helper()

	request, key := buildNativeLiveMethod(t, profile, method)
	pushNativeLiveInitialRefs(t, profile, request)

	return request, key
}

func buildNativeLiveMethod(t *testing.T, profile nativeLiveProfile, method string) (GitPublication, string) {
	t.Helper()

	fixture := newWorkflowGitFixture(t, method)
	request := fixture.request
	key := "native-" + method + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	request.Repository, request.SourceRepository = profile.URL, profile.URL
	request.SourceRef, request.TargetRef = profile.Source+key, profile.Target+key
	request.OwnershipRevision = "qa-owned-ref/" + key
	request.Git.Repository, request.Git.TargetBranch = profile.URL, strings.TrimPrefix(request.TargetRef, "refs/heads/")
	sealWorkflowGitFixture(t, &request)

	return request, key
}

func pushNativeLiveInitialRefs(t *testing.T, profile nativeLiveProfile, request GitPublication) {
	t.Helper()

	workflowIntegrationCommand(t, request.Directory, "push", profile.GitRemote, request.Git.SourceCommit+":"+request.SourceRef, request.Git.BaseCommit+":"+request.TargetRef)
}

func observeNativeLiveRebase(t *testing.T, adapter RemoteReviewAdapter, profile nativeLiveProfile, evidence *nativeLiveEvidence) {
	t.Helper()

	binding := evidence.Binding

	refs := strings.Fields(workflowIntegrationCommand(t, binding.SourceDirectory, "ls-remote", profile.GitRemote, binding.Capability.SourceRef, binding.Capability.TargetRef))
	if len(refs) != 4 || refs[0] != binding.Git.ResultCommit || refs[2] != binding.Git.ResultCommit {
		t.Fatal("actual remote refs differ from the exact prepared result")
	}

	git := NewWorkflowIntegrationGit(profile.URL, profile.GitRemote)
	for attempt := 0; attempt < 40; attempt++ {
		observed, err := adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
		if err != nil {
			evidence.Failure = err.Error()
			t.Fatal(err)
		}

		evidence.Observations = append(evidence.Observations, observed)
		if observed.State == "merged" {
			nativeLiveActivity(t, adapter, profile, observed.PullRequest, "merged")

			return
		}

		if stopNativeLiveObservation(binding, observed) {
			break
		}

		select {
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		case <-time.After(3 * time.Second):
		}
	}

	evidence.Failure = "Git applied; original canonical PR did not converge to compatible merged proof"
	t.Fatal(evidence.Failure)
}

func stopNativeLiveObservation(binding workflowstate.NativePublicationBinding, observed workflowstate.NativePublicationObservation) bool {
	return !time.Now().Before(binding.ObservationDeadline) || (observed.Condition != "" && observed.Condition != "canonical_pr_closed_unmerged")
}

func writeNativeLiveEvidence(t *testing.T, state, key string, evidence *nativeLiveEvidence) {
	t.Helper()

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(state, "chr", "qa", "native-integration-slice-2")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, evidence.Profile+"-"+key+".json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func nativeLiveActivity(t *testing.T, adapter RemoteReviewAdapter, profile nativeLiveProfile, pr workflowstate.PullRequestSnapshot, action string) {
	t.Helper()

	nativeLiveActivityContext(t, t.Context(), adapter, profile, pr, action)
}

func nativeLiveActivityContext(t *testing.T, ctx context.Context, adapter RemoteReviewAdapter, profile nativeLiveProfile, pr workflowstate.PullRequestSnapshot, action string) {
	t.Helper()

	resolvedReview, err := adapter.resolve(pr.URL)
	if err != nil {
		t.Fatal(err)
	}

	reference := resolvedReview.reference

	forge := strings.TrimSuffix(profile.URL, "/"+profile.Repository)

	args := []string{"record", "pull-request", "--forge", forge, "--repository", profile.Repository, "--number", strconv.Itoa(reference.Number), "--action", action, "--source", "agent", "--harness", "codex", fmt.Sprintf("Native integration certification fixture PR %s; real atomic publication and same-original-PR observation.", action)}
	if _, err := exec.CommandContext(ctx, "acta", args...).Output(); err != nil {
		t.Fatal("required fixture activity recording failed")
	}
}
