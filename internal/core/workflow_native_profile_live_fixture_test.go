package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

type nativeLiveProfile struct {
	Forge      string `json:"forge"`
	Repository string `json:"repository"`
	URL        string `json:"url"`
	Private    bool   `json:"private"`
	GitRemote  string `json:"git_remote"`
	Source     string `json:"source_ref_prefix"`
	Target     string `json:"isolated_target_ref_prefix"`
	Role       struct {
		Identity      string `json:"identity"`
		APIReference  string `json:"api_credential_reference"`
		GitReference  string `json:"git_credential_reference"`
		HTTPReference string `json:"https_git_credential_reference"`
		Push          bool   `json:"push"`
	} `json:"role"`
	Policy struct {
		Review            string   `json:"review"`
		TargetPublication string   `json:"target_publication"`
		Rewrite           bool     `json:"owned_source_rewrite_authorized"`
		RewriteCondition  string   `json:"rewrite_condition"`
		NativeProtection  bool     `json:"native_branch_protection"`
		NativeApproval    bool     `json:"native_required_approvals"`
		Signing           bool     `json:"required_commit_signing"`
		Methods           []string `json:"allowed_methods"`
	} `json:"policy"`
	Certification struct {
		Contract string   `json:"contract"`
		Methods  []string `json:"methods"`
		Enabled  bool     `json:"git_merge_detection_enabled"`
	} `json:"observation_certification"`
}

func nativeLiveState(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("live forge certification runs in the nightly suite")
	}

	state := os.Getenv("XDG_STATE_HOME")
	if state != "" {
		return state
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Join(home, ".local", "state")
}

func readNativeLiveProfiles(t *testing.T, state string) []nativeLiveProfile {
	t.Helper()

	profiles, err := loadNativeLiveProfiles(state)
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	for _, profile := range profiles {
		if !validNativeLiveProfile(profile) || seen[profile.Forge] {
			t.Fatal("designated fixture identity, privacy or owned ref prefixes changed")
		}

		seen[profile.Forge] = true
	}

	if len(seen) != 2 {
		t.Fatal("both designated native forge fixtures are required")
	}

	return profiles
}

func validNativeLiveProfile(profile nativeLiveProfile) bool {
	expected := "https://forge.adrianpk.com/adrianpk/chr-integration-qa"
	if profile.Forge == "github" {
		expected = "https://github.com/adrianpk/chr-integration-qa"
	}

	return (profile.Forge == "forgejo" || profile.Forge == "github") && profile.Repository == "adrianpk/chr-integration-qa" && profile.URL == expected && profile.Private && profile.Source == "refs/heads/chr/work/" && profile.Target == "refs/heads/chr/qa/"
}

func loadNativeLiveProfiles(state string) ([]nativeLiveProfile, error) {
	data, err := os.ReadFile(filepath.Join(state, "chr", "qa", "native-integration-fixtures.json"))
	if err != nil {
		return nil, err
	}

	var descriptor struct {
		DeliverySet string              `json:"delivery_set"`
		Profiles    []nativeLiveProfile `json:"profiles"`
	}
	if err := json.Unmarshal(data, &descriptor); err != nil {
		return nil, err
	}

	if descriptor.DeliverySet != "native-integration-queue" {
		return nil, fmt.Errorf("invalid designated fixture descriptor")
	}

	return descriptor.Profiles, nil
}

func nativeLiveTransport(t *testing.T, profile nativeLiveProfile) string {
	t.Helper()

	if strings.HasPrefix(profile.GitRemote, "ssh://") {
		return profile.GitRemote
	}

	location, repository, found := strings.Cut(profile.GitRemote, ":")
	if !found || strings.Contains(profile.GitRemote, "://") {
		t.Fatal("fixture transport is not an explicit SSH location")
	}

	return "ssh://" + location + "/" + repository
}

type nativeLivePolicyReader struct {
	state       string
	profile     nativeLiveProfile
	publication WorkflowGitPublication
}

func (reader nativeLivePolicyReader) ReadWorkflowNativePolicy(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowNativePolicy, error) {
	if err := ctx.Err(); err != nil {
		return WorkflowNativePolicy{}, err
	}

	profiles, err := loadNativeLiveProfiles(reader.state)
	if err != nil {
		return WorkflowNativePolicy{}, err
	}

	for _, profile := range profiles {
		if profile.Forge == reader.profile.Forge && profile.URL == pr.Repository && profile.GitRemote == reader.profile.GitRemote {
			return reader.policy(profile, method)
		}
	}

	return WorkflowNativePolicy{}, fmt.Errorf("current designated fixture profile changed")
}

func (reader nativeLivePolicyReader) policy(profile nativeLiveProfile, method string) (WorkflowNativePolicy, error) {
	request := reader.publication

	if !slices.Contains(profile.Certification.Methods, method) || profile.Role.Identity == "" || !profile.Role.Push || !profile.Certification.Enabled {
		return WorkflowNativePolicy{}, fmt.Errorf("current native fixture role or observation certification changed")
	}

	identity := profile.Role.Identity + "\x00" + profile.Role.APIReference + "\x00" + profile.Role.GitReference

	return WorkflowNativePolicy{
		Repository: profile.URL, SourceRef: request.SourceRef, TargetRef: request.TargetRef, OwnershipRevision: request.OwnershipRevision,
		SourceRewriteAuthorized:       profile.Policy.Rewrite && profile.Policy.RewriteCondition == "Only the exact test-owned source ref at its expected prior SHA",
		FastForwardTargetAuthorized:   profile.Policy.TargetPublication == "Explicitly authorized fast-forward Git publication",
		HostCandidateReview:           profile.Policy.Review == "CHR-managed validation and review of the prepared candidate" && !profile.Policy.NativeApproval && !profile.Policy.NativeProtection && !profile.Policy.Signing,
		ReviewMode:                    ReviewWorkflowModeHumanReview,
		CredentialIdentityFingerprint: contenthash.Sum([]byte(identity)), ObservationContract: profile.Certification.Contract, CertifiedMethods: profile.Certification.Methods,
	}, nil
}
