package workflow

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativeForgejoMethods certifies exact merge, rebase and squash publication
// and both advertised-ref races against the designated real Forgejo fixture.
func TestNativeForgejoMethods(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live certification requires explicit fixture opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		if profile.Forge == "forgejo" {
			nativeMethodCases(t, state, profile)
		}
	}
}

func nativeMethodCases(t *testing.T, state string, profile nativeLiveProfile) {
	t.Helper()

	nativeMethodScenarios(t, state, profile, []string{"publication", "source_race", "target_race"})
}

func nativeMethodScenarios(t *testing.T, state string, profile nativeLiveProfile, scenarios []string) {
	t.Helper()

	server := readNativeCertificationServer(t, profile)
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			for _, scenario := range scenarios {
				t.Run(scenario, func(t *testing.T) {
					nativeMethodCase(t, state, profile, server, method, scenario)
				})
			}
		})
	}
}

func nativeMethodCase(t *testing.T, state string, profile nativeLiveProfile, server, method, scenario string) {
	t.Helper()

	transport := "ssh"
	if strings.HasPrefix(scenario, "https_") {
		transport, scenario = "https", strings.TrimPrefix(scenario, "https_")
	}

	evidence := nativeCertificationEvidence{
		Schema: "chr-native-certification/v1", Method: method, Scenario: scenario,
		Transport: transport,
		Server:    server, Role: "designated repository owner with API and Git push access",
		StartedAt: time.Now().UTC(), Native: nativeLiveEvidence{Profile: profile.Forge},
	}

	request, key := buildNativeLiveMethod(t, profile, method)
	defer retainNativeCertification(t, state, key, &evidence)

	pushNativeLiveInitialRefs(t, profile, request)

	adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), OSReviewRemoteCommandRunner{})
	pr := createNativeLivePR(t, adapter, profile, request)

	evidence.Native.Binding = workflowstate.NativePublicationBinding{
		PullRequest: pr, Git: request.Git, SourceDirectory: request.Directory,
		ObservationDeadline: time.Now().Add(2 * time.Minute),
	}
	publisher := nativeCertificationPublisher(t, profile, request, transport)
	evidence.Native.Binding.Capability = admitNativeMethodProbe(t, state, adapter, profile, request, pr, publisher)

	advanced := installNativeContention(t, profile, &request, scenario)
	result, err := publisher.PublishWorkflowGit(t.Context(), request)

	evidence.Native.Binding.Publication = result
	if scenario != "publication" {
		checkNativeContention(t, adapter, profile, request, advanced, result, err, &evidence)

		return
	}

	if err != nil || result.State != "applied" || !result.CommandsSent {
		t.Fatalf("live publication: %s, %v", result.State, err)
	}

	observeNativeLiveRebase(t, adapter, profile, &evidence.Native)
	checkNativeReceiptCompatibility(t, &evidence)
}

func installNativeContention(t *testing.T, profile nativeLiveProfile, request *GitPublication, scenario string) string {
	t.Helper()

	if scenario == "publication" {
		return ""
	}

	ref, old := request.SourceRef, request.Git.SourceCommit
	if scenario == "target_race" {
		ref, old = request.TargetRef, request.Git.BaseCommit
	}

	workflowIntegrationCommand(t, request.Directory, "checkout", "--detach", old)

	advanced := workflowIntegrationCommit(t, request.Directory, "contention.txt", "External fixture writer")
	request.BeforeSend = func(ctx context.Context, _ workflowstate.GitPublicationAdvertisement) error {
		_, err := outputWorktreeGit(ctx, request.Directory, "push", "--force-with-lease="+ref+":"+old, profile.GitRemote, advanced+":"+ref)

		return err
	}

	return advanced
}

func checkNativeContention(t *testing.T, adapter RemoteReviewAdapter, profile nativeLiveProfile, request GitPublication, advanced string, result workflowstate.GitPublicationResult, err error, evidence *nativeCertificationEvidence) {
	t.Helper()

	if !errors.Is(err, workflowstate.ErrConflict) || !result.CommandsSent || result.State != "rejected" {
		t.Fatalf("atomic contention outcome: %s, %v", result.State, err)
	}

	source, target := request.Git.SourceCommit, request.Git.BaseCommit
	if evidence.Scenario == "source_race" {
		source = advanced
	} else {
		target = advanced
	}

	evidence.ObservedSource = readNativeCertificationRef(t, profile, request.Directory, request.SourceRef)

	evidence.ObservedTarget = readNativeCertificationRef(t, profile, request.Directory, request.TargetRef)
	if evidence.ObservedSource != source || evidence.ObservedTarget != target {
		t.Fatal("rejected atomic publication changed a canonical ref")
	}

	pr, readErr := adapter.ReadWorkflowPullRequest(t.Context(), evidence.Native.Binding.PullRequest.URL)
	if readErr != nil || pr.Merged || pr.State != "open" {
		t.Fatalf("contended original PR: %+v, %v", pr, readErr)
	}

	evidence.ContendedPR = &pr
}
