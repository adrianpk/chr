package app

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// TestCommandFactoryBinding proves actions use their bound assets even when
// unrelated root metadata contains an invalid legacy dependency value.
func TestCommandFactoryBinding(t *testing.T) {
	for _, test := range []struct {
		name    string
		assets  fs.FS
		missing bool
	}{
		{"bound assets", testAssetsFS(), false},
		{"missing assets", fstest.MapFS{}, true},
	} {
		t.Run(test.name, func(t *testing.T) { assertCommandFactoryBinding(t, test.assets, test.missing) })
	}
}

func assertCommandFactoryBinding(t *testing.T, assets fs.FS, missing bool) {
	t.Helper()

	var output bytes.Buffer

	factory := commandFactory{io: IO{Stdin: strings.NewReader(""), Stdout: &output, Stderr: &output}, deps: appDependencies{assetsFS: assets}}
	command := factory.rootCommand()
	command.Metadata = map[string]any{"app.dependencies": struct{}{}}

	project := newTestWorkspace(t)
	if err := os.MkdirAll(filepath.Join(project, ".chr", "default", "tickets"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := command.Run(t.Context(), []string{"chr", "index", "rebuild", "--dir", project, "--dev-root", t.TempDir()})
	if missing {
		if err == nil || !strings.Contains(err.Error(), "migration") {
			t.Fatalf("missing bound assets: %v", err)
		}

		return
	}

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "indexed:") {
		t.Fatalf("bound command output missing: %s", output.String())
	}
}
