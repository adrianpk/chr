package core

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type nativeFaultCase struct {
	state     string
	profile   nativeLiveProfile
	request   WorkflowGitPublication
	adapter   RemoteReviewAdapter
	publisher OSWorkflowGitPublisher
	evidence  nativeCertificationEvidence
}

func newNativeFaultCase(t *testing.T, state string, profile nativeLiveProfile, scenario string) *nativeFaultCase {
	t.Helper()

	request, key := buildNativeLiveMethod(t, profile, "rebase")
	f := &nativeFaultCase{
		state: state, profile: profile, request: request,
		adapter: NewRemoteReviewAdapter(DefaultReviewProviderConfig(), OSReviewRemoteCommandRunner{}),
		evidence: nativeCertificationEvidence{
			Schema: "chr-native-certification/v1", Method: "rebase", Scenario: scenario,
			Transport: "https", Server: readNativeCertificationServer(t, profile),
			Role: "designated repository owner with API and HTTPS Git push access", StartedAt: time.Now().UTC(),
			Native: nativeLiveEvidence{Profile: profile.Forge},
		},
	}
	f.evidence.Native.Binding = WorkflowNativePublicationBinding{Git: request.Git, SourceDirectory: request.Directory, ObservationDeadline: time.Now().Add(2 * time.Minute)}

	t.Cleanup(func() { retainNativeCertification(t, state, key, &f.evidence) })
	pushNativeLiveInitialRefs(t, profile, request)

	f.evidence.Native.Binding.PullRequest = createNativeLivePR(t, f.adapter, profile, request)
	f.publisher = nativeCertificationPublisher(t, profile, request, "https")

	return f
}

func (f *nativeFaultCase) admit(t *testing.T) {
	t.Helper()

	f.evidence.Native.Binding.Capability = admitNativeMethodProbe(t, f.state, f.adapter, f.profile, f.request, f.evidence.Native.Binding.PullRequest, f.publisher)
}

func (f *nativeFaultCase) checkRefs(t *testing.T, source, target string) {
	t.Helper()

	f.evidence.ObservedSource = readNativeCertificationRef(t, f.profile, f.request.Directory, f.request.SourceRef)

	f.evidence.ObservedTarget = readNativeCertificationRef(t, f.profile, f.request.Directory, f.request.TargetRef)
	if f.evidence.ObservedSource != source || f.evidence.ObservedTarget != target {
		t.Fatal("fault case changed an unexpected canonical ref")
	}
}

func nativeFixtureAPI(t *testing.T, ctx context.Context, profile nativeLiveProfile, method, endpoint string, payload any) []byte {
	t.Helper()

	adapter := NewRemoteReviewAdapter(DefaultReviewProviderConfig(), OSReviewRemoteCommandRunner{})

	resolvedReview, err := adapter.resolveWorkflowRepository(profile.URL)

	reference, config := resolvedReview.reference, resolvedReview.config

	if err != nil || !validNativeLiveProfile(profile) || !strings.HasPrefix(endpoint, workflowRemoteRepositoryPath(reference)+"/") {
		t.Fatal("fixture API mutation escaped its designated repository")
	}

	args := workflowRemoteAPIArgs(profile.URL, reference, method, endpoint, payload != nil)
	if payload == nil {
		data, readErr := adapter.workflowProviderRunner().Output(ctx, config.CLI, args...)
		if readErr != nil {
			t.Fatal(readErr)
		}

		return data
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	response, err := adapter.workflowProviderRunner().Input(ctx, config.CLI, data, args...)
	if err != nil {
		t.Fatalf("designated fixture API %s failed: %v", method, err)
	}

	return response
}

func nativeFixtureNote(t *testing.T, ctx context.Context, summary string) {
	t.Helper()

	if _, err := exec.CommandContext(ctx, "acta", "record", "note", "--source", "agent", "--harness", "codex", summary).Output(); err != nil {
		t.Fatal("required fixture activity recording failed")
	}
}
