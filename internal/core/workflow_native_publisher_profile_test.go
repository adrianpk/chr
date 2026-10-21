package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Host certification is fail-closed: malformed, ambiguous, public or absent
// configuration cannot enable publication for a project or another target.
func TestNativePublisherProfiles(t *testing.T) {
	profile := WorkflowNativePublisherProfile{Repository: "https://forge.example/team/project", TargetRef: "refs/heads/dev", Transport: "/srv/git/example.git", CredentialRef: "publisher", APIIdentityFingerprint: strings.Repeat("a", 64), ObservationContract: "forgejo_same_pr_git_auto_merge/v1", CertifiedMethods: []string{"rebase"}, CertificationRef: "native-rebase-acceptance"}
	dir := t.TempDir()
	writeNativeHostProfile(t, dir, profile)

	path := filepath.Join(dir, "native-publishers.json")

	image, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	valid := string(image)

	duplicate, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name, image string
		valid       bool
	}{
		{"valid", valid, true},
		{"unknown field", strings.Replace(valid, `"schema":`, `"unexpected":true,"schema":`, 1), false},
		{"trailing image", valid + `{}`, false},
		{"missing certificate", strings.Replace(valid, "native-rebase-acceptance", "", 1), false},
		{"unknown method", strings.Replace(valid, `"rebase"`, `"automatic"`, 1), false},
		{"duplicate method", strings.Replace(valid, `["rebase"]`, `["rebase","rebase"]`, 1), false},
		{"duplicate profile", strings.Replace(valid, `"profiles":[`, `"profiles":[`+string(duplicate)+`,`, 1), false},
		{"unknown identity", strings.Replace(valid, strings.Repeat("a", 64), "unknown", 1), false},
		{"other target", strings.Replace(valid, "refs/heads/dev", "refs/heads/main", 1), false},
		{"oversized", valid + strings.Repeat(" ", 65537), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(test.image), 0o600); err != nil {
				t.Fatal(err)
			}

			_, err := LoadWorkflowNativePublisher(dir, profile.Repository, profile.TargetRef)
			if (err == nil) != test.valid {
				t.Fatal("publisher configuration acceptance mismatch", err)
			}
		})
	}

	checkNativePublisherPrivacy(t, dir, profile)
}

func checkNativePublisherPrivacy(t *testing.T, dir string, profile WorkflowNativePublisherProfile) {
	t.Helper()

	writeNativeHostProfile(t, dir, profile)

	path := filepath.Join(dir, "native-publishers.json")
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadWorkflowNativePublisher(dir, profile.Repository, profile.TargetRef); err == nil {
		t.Fatal("public host certificate authorized publication")
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadWorkflowNativePublisher(dir, profile.Repository, profile.TargetRef); err == nil {
		t.Fatal("missing host certificate authorized publication")
	}
}
