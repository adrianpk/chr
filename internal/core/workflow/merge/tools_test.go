package merge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeMergerTools checks actual filesystem aliases, not just path strings.
// A permitted pathname must never grant a write to protected or external data.
func TestNativeMergerTools(t *testing.T) {
	input := workflowstate.MergerInput{MergerScope: workflowstate.MergerScope{AllowedPaths: []string{"src", "docs"}, ProtectedPaths: []string{"src/validation.go", "Makefile"}}}

	tools, protected, external := nativeMergerToolFiles(t, input)

	nativeMergerDeniedTools(t, tools)

	if _, err := tools.apply(t.Context(), toolRequest{Operation: "write", Path: "src/service.go", Content: "compatible result"}); err != nil {
		t.Fatal(err)
	}

	response, err := tools.apply(t.Context(), toolRequest{Operation: "read", Path: "src/service.go"})
	if err != nil || response.Content != "compatible result" {
		t.Fatal("owned edit not readable", err)
	}

	for path, expected := range map[string]string{protected: "required validation", external: "external content"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != expected {
			t.Fatal("protected data changed", err)
		}
	}
}

func nativeMergerDeniedTools(t *testing.T, tools *Tools) {
	t.Helper()

	for _, path := range []string{"../external.go", "/etc/passwd", "src/../Makefile", ".git/config", ".chr/receipt.json", "ops/default/ticket/ready/scope.md", "src/validation.go", "src/alias.go", "src/hard.go"} {
		t.Run(path, func(t *testing.T) {
			for _, operation := range []string{"read", "write", "delete"} {
				if _, err := tools.apply(t.Context(), toolRequest{Operation: operation, Path: path}); !errors.Is(err, workflowstate.ErrConflict) {
					t.Fatal("escaped model tool permission", operation, err)
				}
			}
		})
	}
}

func nativeMergerToolFiles(t *testing.T, input workflowstate.MergerInput) (*Tools, string, string) {
	t.Helper()

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}

	protected := filepath.Join(dir, "src", "validation.go")
	if err := os.WriteFile(protected, []byte("required validation"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink("validation.go", filepath.Join(dir, "src", "alias.go")); err != nil {
		t.Fatal(err)
	}

	external := filepath.Join(t.TempDir(), "external.go")
	if err := os.WriteFile(external, []byte("external content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Link(external, filepath.Join(dir, "src", "hard.go")); err != nil {
		t.Fatal(err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = root.Close() }) //nolint:errcheck // Owned fixture directory handle; tool results are asserted separately.

	tools := &Tools{input: input, root: root, guard: func(ctx context.Context) error { return ctx.Err() }}

	return tools, protected, external
}
