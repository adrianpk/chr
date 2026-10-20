package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui"
)

// CLI resolution refuses missing or invalid required login and never prints credentials.
func TestWebLoginConfig(t *testing.T) {
	root, project := t.TempDir(), newTestWorkspace(t)
	args := []string{"web", "serve", "--dev-root", root, "--dir", project, "--allow-lan", "--require-login", "--dry-run"}

	var output bytes.Buffer
	if err := run(t.Context(), args, &output, &bytes.Buffer{}); err == nil {
		t.Fatal("required login started without credentials")
	}

	credentials, err := webui.CreateLoginCredentials("operator", "example-only-password-2026")
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, "config", "chr", "web-login.json")
	if err := webui.SaveLoginCredentials(path, credentials, false); err != nil {
		t.Fatal(err)
	}

	output.Reset()

	if err := run(t.Context(), args, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	assertWebLANOutput(t, output.String(), nil, []string{"login_required: true", "token_required: false", "allow_lan: true"}, "")

	if strings.Contains(output.String(), "operator") || strings.Contains(output.String(), "example-only-password") {
		t.Fatal("dry-run exposed credentials")
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("unsafe configured login fell back to unauthenticated access")
	}
}

// Setup cannot read passwords from unmasked streams or overwrite an account implicitly.
func TestWebLoginSetupGuards(t *testing.T) {
	root := t.TempDir()
	args := []string{"web", "login", "setup", "--dev-root", root, "--username", "operator"}

	var output bytes.Buffer
	if err := run(t.Context(), args, &output, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatal("setup accepted an unmasked password stream", err)
	}

	path := filepath.Join(root, "config", "chr", "web-login.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("failed setup created credentials", err)
	}
}
