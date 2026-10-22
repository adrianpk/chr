package core

import "testing"

// TestGuidedUseCaseAdmissionProvider checks only local readiness and never probes an invented remote PR.
func TestGuidedUseCaseAdmissionProvider(t *testing.T) {
	for _, name := range []string{"ready", "disabled", "missing CLI", "missing token"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("CHR_GUIDED_PROVIDER_TEST_TOKEN", "fixture")

			config := ReviewProviderConfig{Forgejo: ReviewProviderTargetConfig{Enabled: true, CLI: "sh", TokenEnv: "CHR_GUIDED_PROVIDER_TEST_TOKEN"}}

			switch name {
			case "disabled":
				config.Forgejo.Enabled = false
			case "missing CLI":
				config.Forgejo.CLI = "unavailable-guided-provider-test"
			case "missing token":
				t.Setenv("CHR_GUIDED_PROVIDER_TEST_TOKEN", "")
			}

			err := NewRemoteReviewAdapter(config, nil).CheckWorkflowRepositoryProvider(t.Context(), "https://forge.example/team/project")
			if name == "ready" && err != nil || name != "ready" && err == nil {
				t.Fatal("incorrect local provider readiness", err)
			}
		})
	}
}
