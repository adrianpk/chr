package workflow

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativePRChanges changes the original real PR after advertisement and
// proves reconciliation while preserving the irreversible exact Git result.
func TestNativePRChanges(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live certification requires explicit fixture opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		t.Run(profile.Forge, func(t *testing.T) {
			for _, scenario := range []string{"close", "retarget"} {
				t.Run(scenario, func(t *testing.T) { nativePRChange(t, state, profile, scenario) })
			}
		})
	}
}

func nativePRChange(t *testing.T, state string, profile nativeLiveProfile, scenario string) {
	t.Helper()

	f := newNativeFaultCase(t, state, profile, "external_"+scenario)
	f.admit(t)

	f.request.BeforeSend = func(ctx context.Context, _ workflowstate.GitPublicationAdvertisement) error {
		nativeChangeOriginalPR(t, ctx, f, scenario)

		return nil
	}
	result, err := f.publisher.PublishWorkflowGit(t.Context(), f.request)

	f.evidence.Native.Binding.Publication = result
	if err != nil || result.State != "applied" {
		t.Fatalf("metadata edit unexpectedly prevented exact Git application: %s, %v", result.State, err)
	}

	f.checkRefs(t, f.request.Git.ResultCommit, f.request.Git.ResultCommit)

	git := NewWorkflowIntegrationGit(profile.URL, profile.GitRemote)
	observed, err := f.adapter.ObserveWorkflowNativePublication(t.Context(), f.evidence.Native.Binding, git)
	f.evidence.Native.Observations = append(f.evidence.Native.Observations, observed)

	expected := "canonical_pr_closed_unmerged"
	if scenario == "retarget" {
		expected = "canonical_pr_retargeted"
	}

	if err != nil || observed.State != "reconciliation_required" || observed.Condition != expected || observed.PullRequest.Merged {
		t.Fatalf("changed original PR did not require reconciliation: %s/%s, %v", observed.State, observed.Condition, err)
	}

	f.checkRefs(t, f.request.Git.ResultCommit, f.request.Git.ResultCommit)
}

func nativeChangeOriginalPR(t *testing.T, ctx context.Context, f *nativeFaultCase, scenario string) {
	t.Helper()

	pr := f.evidence.Native.Binding.PullRequest

	resolvedReview, err := f.adapter.resolve(pr.URL)
	if err != nil {
		t.Fatal(err)
	}

	reference := resolvedReview.reference

	payload := map[string]string{"state": "closed"}

	if scenario == "retarget" {
		ref := f.request.TargetRef + "-retarget"
		if _, err := outputWorktreeGit(ctx, f.request.Directory, "push", f.profile.GitRemote, f.request.Git.BaseCommit+":"+ref); err != nil {
			t.Fatal(err)
		}

		payload = map[string]string{"base": strings.TrimPrefix(ref, "refs/heads/")}
	}

	endpoint := "/repos/" + f.profile.Repository + "/pulls/" + strconv.Itoa(reference.Number)
	nativeFixtureAPI(t, ctx, f.profile, http.MethodPatch, endpoint, payload)

	if scenario == "close" {
		nativeLiveActivityContext(t, ctx, f.adapter, f.profile, pr, "closed")
	} else {
		nativeFixtureNote(t, ctx, "Retargeted the owned native-integration QA original PR to another isolated fixture target: "+pr.URL)
	}
}
