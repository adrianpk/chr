package workflow

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativePolicyDenials proves anonymous permission denial and incompatible
// enforced native review on real owned fixtures without bypassing protections.
func TestNativePolicyDenials(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live certification requires explicit fixture opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		t.Run(profile.Forge, func(t *testing.T) {
			t.Run("permission", func(t *testing.T) { nativePermissionDenial(t, state, profile) })
			t.Run("native_review", func(t *testing.T) { nativeReviewDenial(t, state, profile) })
		})
	}
}

type nativeHTTPProbe struct {
	status int
	posts  int
}

func (probe *nativeHTTPProbe) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method == http.MethodPost {
		probe.posts++
	}

	response, err := http.DefaultTransport.RoundTrip(request)
	if response != nil {
		probe.status = response.StatusCode
	}

	return response, err
}

func nativePermissionDenial(t *testing.T, state string, profile nativeLiveProfile) {
	t.Helper()

	f := newNativeFaultCase(t, state, profile, "permission_denied")
	f.evidence.Role = "anonymous HTTPS publisher without repository credentials"
	workflowIntegrationCommand(t, f.request.Directory, "config", "--unset-all", "credential.helper")
	workflowIntegrationCommand(t, f.request.Directory, "config", "credential.helper", "")

	probe := &nativeHTTPProbe{}
	result, err := f.publisher.WithHTTPClient(&http.Client{Transport: probe}).PublishWorkflowGit(t.Context(), f.request)

	f.evidence.Native.Binding.Publication, f.evidence.HTTPStatus, f.evidence.Posts = result, probe.status, probe.posts
	if err == nil || result.CommandsSent || result.State != "not_sent" || probe.status != http.StatusUnauthorized || probe.posts != 0 {
		t.Fatalf("anonymous publication was not denied before commands: %s, status %d, %v", result.State, probe.status, err)
	}

	f.checkRefs(t, f.request.Git.SourceCommit, f.request.Git.BaseCommit)
}

func nativeReviewDenial(t *testing.T, state string, profile nativeLiveProfile) {
	t.Helper()

	f := newNativeFaultCase(t, state, profile, "native_review_denied")
	branch := strings.TrimPrefix(f.request.TargetRef, "refs/heads/")
	endpoint := "/repos/" + profile.Repository + "/branch_protections"
	method := http.MethodPost

	payload := map[string]any{"rule_name": branch, "enable_push": false, "required_approvals": 1}
	if profile.Forge == "github" {
		endpoint = "/repos/" + profile.Repository + "/branches/" + url.PathEscape(branch) + "/protection"
		method = http.MethodPut
		payload = map[string]any{"required_status_checks": nil, "enforce_admins": true, "required_pull_request_reviews": map[string]int{"required_approving_review_count": 1}, "restrictions": nil}
	}

	f.evidence.Policy = nativeFixtureAPI(t, t.Context(), profile, method, endpoint, payload)
	nativeFixtureNote(t, t.Context(), "Created an exact owned native-integration QA target policy requiring native review; retained protection for "+profile.URL+" branch "+branch+".")

	query := NativeCapabilityRequest{Config: workspaceconfig.WorkIntegrationConfig{PublicationProfile: workspaceconfig.WorkflowPublicationGitAtomicDualRef}, PullRequest: f.evidence.Native.Binding.PullRequest, Publication: f.request}
	policy := nativeMethodProbePolicy{nativeLivePolicyReader: nativeLivePolicyReader{state: state, profile: profile, publication: f.request}, https: true}
	capability, err := f.adapter.WorkflowNativeIntegrationCapability(t.Context(), query, policy, f.publisher)

	f.evidence.Native.Binding.Capability = capability
	if err != nil || capability.AtomicSourceBase || !nativeProtectionDenied(capability) {
		t.Fatalf("enforced native review was not rejected: %+v, %v", capability.Unsupported, err)
	}

	f.checkRefs(t, f.request.Git.SourceCommit, f.request.Git.BaseCommit)

	read := "/repos/" + profile.Repository + "/branches/" + url.PathEscape(branch)
	data := nativeFixtureAPI(t, t.Context(), profile, http.MethodGet, read, nil)

	var current struct {
		Protected *bool `json:"protected"`
	}
	if json.Unmarshal(data, &current) != nil || current.Protected == nil || !*current.Protected {
		t.Fatal("native review protection was not retained after denial")
	}
}

func nativeProtectionDenied(capability workflowstate.NativeCapability) bool {
	for _, reason := range capability.Unsupported {
		if reason.Condition == "native_protections_required" || reason.Condition == "native_rules_required" {
			return true
		}
	}

	return false
}
