package core

import (
	"os"
	"testing"
)

// TestNativeGitHubMethods certifies all three methods and both real ref races
// through the existing GitHub adapter without a provider merge queue.
func TestNativeGitHubMethods(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live certification requires explicit fixture opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		if profile.Forge == "github" {
			nativeMethodCases(t, state, profile)
		}
	}
}
