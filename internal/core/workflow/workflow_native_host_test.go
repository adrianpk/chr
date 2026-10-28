package workflow

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type nativeHostFixture struct {
	f        workflowPreparationFixture
	provider *NativeReviewAdapter
	profile  NativePublisherProfile
	forge    nativeCapabilityFixture
}

func newNativeHostFixture(t *testing.T) nativeHostFixture {
	t.Helper()

	f, _ := workflowTicketConfiguredFixture(t, "rebase", "ticket", func(_ *Store, _ Record, request *StartRequest) {
		request.Policy.MaxWorkerAttempts, request.Policy.MaxIntegrationAttempts = 1, 1
	})
	writeWorkflowTestConfigValue(t, f.store.layout.ProjectDir(), "work", "integration.publication_profile", workspaceconfig.WorkflowPublicationGitAtomicDualRef)

	runner := &workflowPRCommandFixture{fakeRemoteReviewRunner: newFakeRemoteReviewRunner()}
	remote := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)
	provider := NewWorkflowNativeReviewAdapter(remote, f.store, f.run.ID, t.TempDir())
	profile := NativePublisherProfile{Repository: f.provider.pr.Repository, TargetRef: "refs/heads/dev", Transport: f.remote, CredentialRef: "fixture-publisher", APIIdentityFingerprint: contenthash.Sum([]byte(`{"id":1,"login":"fixture-publisher"}`)), ObservationContract: "forgejo_same_pr_git_auto_merge/v1", CertifiedMethods: []string{"rebase"}, CertificationRef: "fixture/native-rebase-certificate"}
	writeNativeHostProfile(t, provider.configDir, profile)

	forge := nativeCapabilityFixture{adapter: remote, runner: runner, request: NativeCapabilityRequest{PullRequest: f.provider.pr}}
	forge.prefix = nativePolicyCommand(t, forge, "")
	setNativePRFixture(t, forge, f.provider.pr)

	runner.outputs[forge.prefix] = []byte(`{"html_url":"` + profile.Repository + `","permissions":{"push":true},"allow_rebase":true}`)
	runner.outputs[forge.prefix+"/branches/dev"] = []byte(`{"protected":false,"commit":{"id":"` + f.base + `"}}`)
	runner.outputs[forge.prefix+"/branches/"+url.PathEscape(f.provider.pr.HeadBranch)] = []byte(`{"protected":false,"commit":{"id":"` + f.provider.pr.HeadCommit + `"}}`)

	resolvedReview, err := remote.resolve(f.provider.pr.URL)
	if err != nil {
		t.Fatal(err)
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	runner.outputs[remoteReviewCommandKey(config.CLI, workflowForgeEndpoint(profile.Repository, reference, config).Arguments(workflowforge.Request{Method: "GET", Path: "/user"})...)] = []byte(`{"id":1,"login":"fixture-publisher"}`)

	return nativeHostFixture{f, provider, profile, forge}
}

func writeNativeHostProfile(t *testing.T, configDir string, profile NativePublisherProfile) {
	t.Helper()

	data, err := json.Marshal(struct {
		Schema   string                   `json:"schema"`
		Profiles []NativePublisherProfile `json:"profiles"`
	}{"chr-native-publishers/v1", []NativePublisherProfile{profile}})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "native-publishers.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// The real remote adapter prepares and reviews combined content before the
// production atomic transport advances both owned refs. Worker work is reused.
func TestNativeHostPublication(t *testing.T) {
	for _, mode := range []string{"automatic", "explicit"} {
		t.Run(mode, func(t *testing.T) { nativeHostPublication(t, mode) })
	}
}

func nativeHostPublication(t *testing.T, mode string) {
	t.Helper()

	host := newNativeHostFixture(t)
	f, provider := host.f, host.provider
	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	entry = nativeRoleTransition(t, f.store, entry, "preparing")
	nativeRoleTransition(t, f.store, entry, "validating")

	candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, provider, f.git)
	if err != nil {
		t.Fatal("real provider failed native preparation", err)
	}

	if candidate.Validation.Validation[0].Stdout != "combined" {
		t.Fatal("host gate did not validate the combined result")
	}

	review, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, provider, nil)
	if err != nil || review.State != workflowstate.ReviewAuthorized {
		t.Fatal("trusted review did not authorize the combined result", review, err)
	}

	checkNativeHostRevocation(t, host)

	nativeContinuation, err := f.store.workflowNativeContinuationOwner(t.Context(), f.run.ID)
	entry, guard := nativeContinuation.Entry, nativeContinuation.Guard

	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.transitionWorkflowNativeContinuation(t.Context(), entry, guard, "ready_to_publish"); err != nil {
		t.Fatal(err)
	}

	nativeContinuation2, err := f.store.workflowNativeContinuationOwner(t.Context(), f.run.ID)
	guard = nativeContinuation2.Guard

	if err != nil {
		t.Fatal(err)
	}

	nativePublication, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, provider, provider)

	publication := nativePublication.Transport
	if err != nil || publication.State != workflowstate.GitPublicationApplied {
		t.Fatal("real provider failed native publication", publication.State, err)
	}

	checkNativeHostPublishedRefs(t, host, candidate.Input.Git.ResultCommit)
	checkNativeHostReceipt(t, host, candidate.Input.Git.ResultCommit, mode)
}

func checkNativeHostReceipt(t *testing.T, host nativeHostFixture, result, mode string) {
	t.Helper()

	if _, ok := any(host.provider).(GitPublicationInspector); !ok {
		t.Fatal("production publisher cannot reconcile its own publication")
	}

	pr := host.f.provider.pr
	pr.State, pr.Merged, pr.HeadCommit, pr.BaseCommit, pr.ResultCommit = "closed", true, result, result, result
	setNativePRFixture(t, host.forge, pr)

	host.forge.runner.outputs[host.forge.prefix+"/branches/"+url.PathEscape(pr.HeadBranch)] = []byte(`{"protected":false,"commit":{"id":"` + result + `"}}`)

	ports := PublicationPorts{Provider: host.provider, Preparer: host.f.git, Observer: NewWorkflowIntegrationGit(pr.Repository, host.f.remote), NativePublisher: host.provider, NativeObserver: host.provider}
	if err := confirmNativeHostReceipt(t, host, ports, mode); err != nil {
		t.Fatal("real provider could not confirm its publication receipt", err)
	}

	if _, err := host.f.store.AdvanceWorkflowSliceIntegration(t.Context(), host.f.run.ID, ports); err != nil {
		t.Fatal("real provider could not finalize its publication receipt", err)
	}

	if run := workflowTestRead(t, host.f.store, host.f.run.ID); run.DeliveryRef == "" {
		t.Fatal("publication did not produce a canonical delivery receipt")
	}
}

func confirmNativeHostReceipt(t *testing.T, host nativeHostFixture, ports PublicationPorts, mode string) error {
	t.Helper()

	if mode == "explicit" {
		_, err := host.f.store.ReconcileWorkflowNativeIntegration(t.Context(), host.f.run.ID, ports)

		return err
	}

	_, err := host.f.store.AdvanceWorkflowSliceIntegration(t.Context(), host.f.run.ID, ports)

	return err
}

func checkNativeHostRevocation(t *testing.T, host nativeHostFixture) {
	t.Helper()

	changed := host.profile
	changed.CredentialRef = "changed-publisher"
	writeNativeHostProfile(t, host.provider.configDir, changed)

	if _, err := host.f.store.CheckWorkflowPublicationGuard(t.Context(), host.f.run.ID, host.provider); err == nil {
		t.Fatal("changed host publisher reused previous review authorization")
	}

	writeNativeHostProfile(t, host.provider.configDir, host.profile)
}

func checkNativeHostPublishedRefs(t *testing.T, host nativeHostFixture, result string) {
	t.Helper()

	for _, ref := range []string{host.f.run.Lane.Branch, host.profile.TargetRef} {
		if actual := workflowIntegrationCommand(t, host.f.remote, "rev-parse", ref); actual != result {
			t.Fatal("publication did not advance the exact owned refs", ref, actual)
		}
	}

	run := workflowTestRead(t, host.f.store, host.f.run.ID)
	if run.Cursor != host.f.run.Cursor || run.CandidateHash != host.f.run.CandidateHash || run.DeliveryRef != "" {
		t.Fatal("publication repeated worker work or claimed delivery without canonical observation")
	}
}

// A certified target cannot authorize a different source, fork, method or API
// identity. All denials happen before sending either ref update.
func TestNativeHostDenials(t *testing.T) {
	host := newNativeHostFixture(t)
	for _, test := range []struct {
		name   string
		change func(*workflowstate.PullRequestSnapshot)
	}{
		{"source", func(pr *workflowstate.PullRequestSnapshot) { pr.HeadBranch = "human-owned" }},
		{"fork", func(pr *workflowstate.PullRequestSnapshot) { pr.HeadRepository = "https://forge.example/other/repo" }},
		{"target", func(pr *workflowstate.PullRequestSnapshot) { pr.BaseBranch = "main" }},
		{"identity", func(pr *workflowstate.PullRequestSnapshot) { pr.URL += "0" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			pr := host.f.provider.pr
			test.change(&pr)

			if _, err := host.provider.ReadWorkflowNativePolicy(t.Context(), pr, "rebase"); err == nil {
				t.Fatal("foreign ownership reused native publication authority")
			}
		})
	}

	if _, err := host.provider.ReadWorkflowNativePolicy(t.Context(), host.f.provider.pr, "squash"); err == nil {
		t.Fatal("uncertified method was authorized")
	}

	changed := host.profile
	changed.APIIdentityFingerprint = contenthash.Sum([]byte("different-api-identity"))
	writeNativeHostProfile(t, host.provider.configDir, changed)

	if _, err := host.provider.WorkflowIntegrationCapability(t.Context(), host.f.provider.pr, "rebase"); err == nil {
		t.Fatal("a different API identity reused the host publisher certificate")
	}
}
