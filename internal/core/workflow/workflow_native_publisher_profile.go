package workflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativePublisherProfile is host configuration for one canonical target.
// Credentials remain in the configured Git and API credential stores.
type NativePublisherProfile struct {
	Repository             string   `json:"repository"`
	TargetRef              string   `json:"target_ref"`
	Transport              string   `json:"transport"`
	CredentialRef          string   `json:"credential_ref"`
	APIIdentityFingerprint string   `json:"api_identity_fingerprint"`
	ObservationContract    string   `json:"observation_contract"`
	CertifiedMethods       []string `json:"certified_methods"`
	CertificationRef       string   `json:"certification_ref"`
}

// LoadWorkflowNativePublisher requires a private host profile outside the
// project. A project publication setting alone cannot invent certification.
func LoadWorkflowNativePublisher(configDir, repository, target string) (NativePublisherProfile, error) {
	path := filepath.Join(configDir, "native-publishers.json")

	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > 65536 {
		return NativePublisherProfile{}, fmt.Errorf("%w: configure a private native-publishers.json host profile", workflowstate.ErrWorkflowAtomicGuard)
	}

	file, err := os.Open(path)
	if err != nil {
		return NativePublisherProfile{}, fmt.Errorf("native publisher configuration is unavailable")
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // Read-only; the bounded reader reports IO failures.

	data, err := io.ReadAll(io.LimitReader(file, 65537))
	if err != nil || len(data) > 65536 {
		return NativePublisherProfile{}, fmt.Errorf("native publisher configuration is unavailable")
	}

	return parseWorkflowNativePublisher(data, repository, target)
}

func parseWorkflowNativePublisher(data []byte, repository, target string) (NativePublisherProfile, error) {
	var file struct {
		Schema   string                   `json:"schema"`
		Profiles []NativePublisherProfile `json:"profiles"`
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&file); err != nil {
		return NativePublisherProfile{}, fmt.Errorf("invalid native publisher configuration")
	}

	if file.Schema != "chr-native-publishers/v1" || len(file.Profiles) > 64 || !errors.Is(decoder.Decode(new(any)), io.EOF) {
		return NativePublisherProfile{}, fmt.Errorf("invalid native publisher configuration")
	}

	return selectWorkflowNativePublisher(file.Profiles, repository, target)
}

func selectWorkflowNativePublisher(profiles []NativePublisherProfile, repository, target string) (NativePublisherProfile, error) {
	seen := map[string]bool{}

	var selected NativePublisherProfile

	for _, profile := range profiles {
		key := profile.Repository + "\x00" + profile.TargetRef
		if err := validateWorkflowNativePublisher(profile); err != nil || seen[key] {
			return selected, fmt.Errorf("invalid or duplicate native publisher profile")
		}

		seen[key] = true

		if profile.Repository == repository && profile.TargetRef == target {
			selected = profile
		}
	}

	if selected.Repository == "" {
		return selected, fmt.Errorf("%w: no certified publisher for the exact canonical target", workflowstate.ErrWorkflowAtomicGuard)
	}

	return selected, nil
}

func validateWorkflowNativePublisher(profile NativePublisherProfile) error {
	canonical, err := workflowstate.CanonicalRepository(profile.Repository)
	if err != nil || canonical != profile.Repository || !workflowstate.ValidFingerprint(profile.APIIdentityFingerprint) {
		return workflowstate.ErrConflict
	}

	if !strings.HasPrefix(profile.TargetRef, "refs/heads/") || !workflowstate.ValidBranch(strings.TrimPrefix(profile.TargetRef, "refs/heads/")) {
		return workflowstate.ErrConflict
	}

	for _, value := range []string{profile.Transport, profile.CredentialRef, profile.CertificationRef} {
		if value == "" || len(value) > 1024 || strings.ContainsAny(value, "\x00\r\n") {
			return workflowstate.ErrConflict
		}
	}

	return validateWorkflowNativePublisherMethods(profile)
}

func validateWorkflowNativePublisherMethods(profile NativePublisherProfile) error {
	if !slices.Contains([]string{"forgejo_same_pr_git_auto_merge/v1", "github_same_pr_git_merge/v1", workflowNativeSourceResultContract}, profile.ObservationContract) || len(profile.CertifiedMethods) == 0 || len(profile.CertifiedMethods) > 3 {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, method := range profile.CertifiedMethods {
		if !slices.Contains([]string{"merge", "rebase", "squash"}, method) || seen[method] {
			return workflowstate.ErrConflict
		}

		seen[method] = true
	}

	return nil
}
