package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Actual CLI status/select/clear adapters share the service transaction, preserve
// request replay and restore legacy generation without clearing the worker role.
func TestHarnessSelectionCLI(t *testing.T) {
	project, root := harnessSelectionCLIFixture(t)
	status := harnessSelectionCLIStatus(t, project, root)
	args := []string{"harness", "select", "--dir", project, "--dev-root", root, "--config-revision", status.Snapshot.Revision, "--request-id", "cli-select", "--worker", "writing", "--generation", "writing", "--json"}

	var first, second bytes.Buffer
	if err := run(t.Context(), args, &first, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), args, &second, &bytes.Buffer{}); err != nil || first.String() != second.String() {
		t.Fatal("CLI changed its retried response", err)
	}

	status = harnessSelectionCLIStatus(t, project, root)
	if status.Worker.Profile != "writing" || status.Generation.Profile != "writing" {
		t.Fatal("CLI did not publish both roles", status)
	}

	checkHarnessSelectionCLIClear(t, project, root, status.Snapshot.Revision)
}

func checkHarnessSelectionCLIClear(t *testing.T, project, root, revision string) {
	t.Helper()

	args := []string{"harness", "clear", "--dir", project, "--dev-root", root, "--config-revision", revision, "--request-id", "cli-clear", "--generation"}

	var cleared bytes.Buffer
	if err := run(t.Context(), args, &cleared, &bytes.Buffer{}); err != nil || !strings.Contains(cleared.String(), "Applies to new operations.") {
		t.Fatal("clear failed", cleared.String(), err)
	}

	status := harnessSelectionCLIStatus(t, project, root)
	if status.Worker.Profile != "writing" || status.Generation.Profile != "" || status.Generation.Source != "legacy" || status.Generation.Backend != "openai" {
		t.Fatal("clear did not restore the action default", status)
	}
}

// Malformed role changes and missing local actor authority cannot publish through
// the CLI, even when the supplied configuration revision is current.
func TestHarnessSelectionCLIRejects(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		actor string
	}{
		{"no role", nil, "operator"},
		{"empty profile", []string{"--worker", ""}, "operator"},
		{"executable path", []string{"--worker", "/private/driver"}, "operator"},
		{"missing actor", []string{"--worker", "writing"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project, root := harnessSelectionCLIFixture(t)
			t.Setenv(core.ActiveOperatorIDEnv, tc.actor)

			initial := harnessSelectionCLIStatus(t, project, root)
			args := []string{"harness", "select", "--dir", project, "--dev-root", root, "--config-revision", initial.Snapshot.Revision, "--request-id", "cli-reject"}

			args = append(args, tc.args...)
			if err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("invalid CLI request accepted")
			}

			after := harnessSelectionCLIStatus(t, project, root)
			if after.Snapshot != initial.Snapshot {
				t.Fatal("CLI rejection changed selections")
			}
		})
	}
}

func harnessSelectionCLIFixture(t *testing.T) (string, string) {
	t.Helper()

	_, project, configDir, _ := harnessSelectionFixture(t)
	root := t.TempDir()
	t.Setenv(core.ActiveOperatorIDEnv, "operator")
	t.Setenv("CHR_SPEC_GENERATION_BACKEND", "openai")

	profiles, err := os.ReadFile(filepath.Join(configDir, "harnesses.yml"))
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(root, "config", "chr", "harnesses.yml"), string(profiles))
	harnessSelectionWrite(t, core.UserPath(project, "config.yml"), "unknown_extension:\n  setting: keep\n")

	return project, root
}

func harnessSelectionCLIStatus(t *testing.T, project, root string) HarnessSelectionStatus {
	t.Helper()

	var output bytes.Buffer

	args := []string{"harness", "status", "--dir", project, "--dev-root", root, "--json"}
	if err := run(t.Context(), args, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var status HarnessSelectionStatus
	if err := json.Unmarshal(output.Bytes(), &status); err != nil {
		t.Fatal(err)
	}

	return status
}
