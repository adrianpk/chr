package core

import (
	"os"
	"testing"
)

// TestNativeHTTPSMethods repeats all method and contention proofs on each real
// forge's smart HTTPS endpoint using its existing host credential helper.
func TestNativeHTTPSMethods(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live certification requires explicit fixture opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		t.Run(profile.Forge, func(t *testing.T) {
			nativeMethodScenarios(t, state, profile, []string{"https_publication", "https_source_race", "https_target_race"})
		})
	}
}

func nativeCertificationPublisher(t *testing.T, profile nativeLiveProfile, request WorkflowGitPublication, transport string) OSWorkflowGitPublisher {
	t.Helper()

	if transport == "ssh" {
		return NewWorkflowGitPublisher(profile.URL, nativeLiveTransport(t, profile))
	}

	helper := "!tea logins helper"
	if profile.Forge == "github" {
		helper = "!gh auth git-credential"
	}

	workflowIntegrationCommand(t, request.Directory, "config", "credential.helper", "")
	workflowIntegrationCommand(t, request.Directory, "config", "--add", "credential.helper", helper)

	return NewWorkflowGitPublisher(profile.URL, profile.URL+".git")
}
