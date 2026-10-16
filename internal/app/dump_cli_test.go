package app

import (
	"os"
	"path/filepath"
	"testing"
)

// configureDumpCLICuration pins the external provider response for CLI persistence assertions.
func configureDumpCLICuration(t *testing.T) {
	t.Helper()
	program := filepath.Join(t.TempDir(), "dump-curation-provider")
	script := `#!/bin/sh
printf '%s\n' '{"polished_body":"Dump body from CLI.","highlights":["Dump body from CLI."],"warnings":[]}'
`
	if err := os.WriteFile(program, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHR_DUMP_GENERATION_BACKEND", "codex")
	t.Setenv("CHR_CODEX_COMMAND", program)
}
