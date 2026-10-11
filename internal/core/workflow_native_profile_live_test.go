package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type nativeLiveEvidence struct {
	Profile      string                                 `json:"profile"`
	Binding      WorkflowNativePublicationBinding       `json:"binding"`
	Observations []WorkflowNativePublicationObservation `json:"observations"`
	Failure      string                                 `json:"failure,omitempty"`
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
	adapter := NewRemoteReviewAdapter(DefaultReviewProviderConfig(), OSReviewRemoteCommandRunner{})
	pr := createNativeLivePR(t, adapter, profile, request)
	publisher := NewWorkflowGitPublisher(profile.URL, nativeLiveTransport(t, profile))
	capability := admitNativeLiveProfile(t, state, adapter, profile, request, pr, publisher)
	result, publicationErr := publisher.PublishWorkflowGit(t.Context(), request)
	binding := WorkflowNativePublicationBinding{Capability: capability, PullRequest: pr, Git: request.Git, Publication: result, SourceDirectory: request.Directory, ObservationDeadline: time.Now().Add(2 * time.Minute)}

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

func createNativeLivePR(t *testing.T, adapter RemoteReviewAdapter, profile nativeLiveProfile, request WorkflowGitPublication) WorkflowPullRequestSnapshot {
	t.Helper()

	pr, err := adapter.CreateWorkflowPullRequest(t.Context(), WorkflowPullRequestRequest{
		WorkflowPullRequestQuery: WorkflowPullRequestQuery{Repository: profile.URL, HeadBranch: strings.TrimPrefix(request.SourceRef, "refs/heads/"), BaseBranch: request.Git.TargetBranch},
		Title:                    "Native atomic rebase certification", Body: "Synthetic owned-source rebase fixture for exact-result publication and same-original-PR observation.",
	})
	if err != nil || pr.HeadCommit != request.Git.SourceCommit || pr.BaseCommit != request.Git.BaseCommit || pr.State != "open" {
		t.Fatalf("canonical fixture PR creation: %+v, %v", pr, err)
	}

	nativeLiveActivity(t, adapter, profile, pr, "opened")
	t.Logf("canonical PR %s, original source %s, base %s, prepared result %s", pr.URL, request.Git.SourceCommit, request.Git.BaseCommit, request.Git.ResultCommit)

	return pr
}

func admitNativeLiveProfile(t *testing.T, state string, adapter RemoteReviewAdapter, profile nativeLiveProfile, request WorkflowGitPublication, pr WorkflowPullRequestSnapshot, publisher OSWorkflowGitPublisher) WorkflowNativeCapability {
	t.Helper()

	policy := nativeLivePolicyReader{state: state, profile: profile, publication: request}
	query := WorkflowNativeCapabilityRequest{Config: WorkIntegrationConfig{PublicationProfile: WorkflowPublicationGitAtomicDualRef}, PullRequest: pr, Publication: request}

	capability, err := adapter.WorkflowNativeIntegrationCapability(t.Context(), query, policy, publisher)
	if err != nil || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 {
		t.Fatalf("live native capability: %+v, %v", capability, err)
	}

	if err := adapter.RecheckWorkflowNativeCapability(t.Context(), query, capability, policy, publisher); err != nil {
		t.Fatal(err)
	}

	return capability
}

func prepareNativeLiveRebase(t *testing.T, profile nativeLiveProfile) (WorkflowGitPublication, string) {
	t.Helper()

	fixture := newWorkflowGitFixture(t, "rebase")
	request := fixture.request
	key := "native-rebase-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	request.Repository, request.SourceRepository = profile.URL, profile.URL
	request.SourceRef, request.TargetRef = profile.Source+key, profile.Target+key
	request.OwnershipRevision = "qa-owned-ref/" + key
	request.Git.Repository, request.Git.TargetBranch = profile.URL, strings.TrimPrefix(request.TargetRef, "refs/heads/")
	sealWorkflowGitFixture(t, &request)
	workflowIntegrationCommand(t, request.Directory, "push", profile.GitRemote, request.Git.SourceCommit+":"+request.SourceRef, request.Git.BaseCommit+":"+request.TargetRef)

	return request, key
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

func stopNativeLiveObservation(binding WorkflowNativePublicationBinding, observed WorkflowNativePublicationObservation) bool {
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

func nativeLiveActivity(t *testing.T, adapter RemoteReviewAdapter, profile nativeLiveProfile, pr WorkflowPullRequestSnapshot, action string) {
	t.Helper()

	reference, _, err := adapter.resolve(pr.URL)
	if err != nil {
		t.Fatal(err)
	}

	forge := strings.TrimSuffix(profile.URL, "/"+profile.Repository)

	args := []string{"record", "pull-request", "--forge", forge, "--repository", profile.Repository, "--number", strconv.Itoa(reference.Number), "--action", action, "--source", "agent", "--harness", "codex", fmt.Sprintf("Native rebase certification fixture PR %s; real atomic publication and same-original-PR observation.", action)}
	if _, err := exec.CommandContext(t.Context(), "acta", args...).Output(); err != nil {
		t.Fatal("required fixture activity recording failed")
	}
}
