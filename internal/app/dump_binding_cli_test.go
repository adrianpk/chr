package app

import (
	"bytes"
	"strings"
	"testing"
)

// CLI callers can retain a document revision and operation identity across a
// process restart; append retry returns its receipt instead of duplicating text.
func TestDumpCLIWriteBinding(t *testing.T) {
	configureDumpCLICuration(t)

	project := newTestWorkspace(t)
	if _, err := runDumpBindingCLI(t, project, "new", "--title", "General", "--content", "Initial material."); err != nil {
		t.Fatal(err)
	}

	ref := "ops/default/dump/00000-general/artifact.md"

	revision, err := runDumpBindingCLI(t, project, "show", "--revision", ref)
	if err != nil || len(strings.TrimSpace(revision)) != 64 {
		t.Fatal("CLI did not expose a document revision", revision, err)
	}

	args := []string{"--operation-id", "cli-append", "--revision", strings.TrimSpace(revision), "--content", "Retry note.", ref}
	for range 2 {
		if _, err := runDumpBindingCLI(t, project, "append", args...); err != nil {
			t.Fatal("CLI retry lost its retained binding", err)
		}
	}

	content, err := runDumpBindingCLI(t, project, "show", ref)
	if err != nil || strings.Count(content, "Retry note.") != 1 {
		t.Fatal("CLI retry duplicated append", content, err)
	}

	if _, err := runDumpBindingCLI(t, project, "save", "--operation-id", "unbound", "--content", "Replacement.", ref); err == nil {
		t.Fatal("retained edit ID accepted without its original revision")
	}
}

func runDumpBindingCLI(t *testing.T, project, action string, args ...string) (string, error) {
	t.Helper()

	command := append([]string{"dump", action, "--dir", project}, args...)

	var output bytes.Buffer

	err := run(t.Context(), command, &output, &bytes.Buffer{})

	return output.String(), err
}
