package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// nativeMethodProbePolicy admits only the designated synthetic case under its
// current host authorization. Passing the live case supplies certification;
// this reader does not change the host's previously certified method list.
type nativeMethodProbePolicy struct {
	nativeLivePolicyReader
	https bool
}

func (reader nativeMethodProbePolicy) ReadWorkflowNativePolicy(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowNativePolicy, error) {
	if err := ctx.Err(); err != nil {
		return WorkflowNativePolicy{}, err
	}

	profiles, err := loadNativeLiveProfiles(reader.state)
	if err != nil {
		return WorkflowNativePolicy{}, err
	}

	for _, profile := range profiles {
		if profile.Forge != reader.profile.Forge || profile.URL != pr.Repository || profile.GitRemote != reader.profile.GitRemote {
			continue
		}

		return reader.methodPolicy(profile, method)
	}

	return WorkflowNativePolicy{}, fmt.Errorf("designated method probe profile changed")
}

func (reader nativeMethodProbePolicy) methodPolicy(profile nativeLiveProfile, method string) (WorkflowNativePolicy, error) {
	if !validNativeLiveProfile(profile) || !slices.Contains(profile.Policy.Methods, method) {
		return WorkflowNativePolicy{}, fmt.Errorf("selected method lacks designated fixture authorization")
	}

	profile.Certification.Methods = []string{method}
	if profile.Forge == "github" {
		profile.Certification.Contract = workflowNativeSourceResultContract
	}

	if reader.https {
		if profile.Role.HTTPReference == "" {
			return WorkflowNativePolicy{}, fmt.Errorf("designated HTTPS publisher role is unavailable")
		}

		profile.Role.GitReference = profile.Role.HTTPReference
	}

	return reader.policy(profile, method)
}

func admitNativeMethodProbe(t *testing.T, state string, adapter RemoteReviewAdapter, profile nativeLiveProfile, request WorkflowGitPublication, pr WorkflowPullRequestSnapshot, publisher OSWorkflowGitPublisher) WorkflowNativeCapability {
	t.Helper()

	policy := nativeMethodProbePolicy{nativeLivePolicyReader: nativeLivePolicyReader{state: state, profile: profile, publication: request}, https: strings.HasPrefix(publisher.transport, "https://")}
	query := WorkflowNativeCapabilityRequest{Config: workspaceconfig.WorkIntegrationConfig{PublicationProfile: workspaceconfig.WorkflowPublicationGitAtomicDualRef}, PullRequest: pr, Publication: request}

	capability, err := adapter.WorkflowNativeIntegrationCapability(t.Context(), query, policy, publisher)
	if err != nil || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 {
		t.Fatalf("live method probe capability: %+v, %v", capability, err)
	}

	if err := adapter.RecheckWorkflowNativeCapability(t.Context(), query, capability, policy, publisher); err != nil {
		t.Fatal(err)
	}

	return capability
}

func readNativeCertificationServer(t *testing.T, profile nativeLiveProfile) string {
	t.Helper()

	if profile.Forge == "github" {
		return "github.com hosted REST API through the existing gh adapter; server version is not exposed"
	}

	data, err := exec.CommandContext(t.Context(), "tea", "api", "--login", "forge", "/version").Output()
	if err != nil {
		t.Fatal("designated Forgejo server version is unavailable")
	}

	var response struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(data, &response) != nil || response.Version == "" {
		t.Fatal("designated Forgejo server version is invalid")
	}

	return "Forgejo " + response.Version + "; REST api/v1 through the existing tea adapter"
}
