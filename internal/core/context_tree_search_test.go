package core

import (
	"path/filepath"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestLoadTreeSearchContextCapturesBoundedTree(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "internal", "todo", "store.go"), "package todo\n")
	writeTestFile(t, filepath.Join(dir, "internal", "todo", "store_test.go"), "package todo\n")
	writeTestFile(t, filepath.Join(dir, "docs", "reference", "todo.md"), "# Todo\n")
	writeTestFile(t, filepath.Join(dir, ".chr", "default", "config.yml"), "version: 1\n")
	writeTestFile(t, filepath.Join(dir, ".chr", "default", "runtime", "runs", "exec-1", "runtime.log"), "skip\n")
	writeTestFile(t, filepath.Join(dir, ".git", "config"), "skip\n")

	ctx, err := LoadTreeSearchContext(TreeSearchContextOptions{
		ProjectDir:     dir,
		MaxTreeEntries: 20,
	})
	if err != nil {
		t.Fatalf("LoadTreeSearchContext returned error: %v", err)
	}

	rendered := RenderCodeContext(ctx)
	for _, want := range []string{
		"internal/",
		"internal/todo/store.go",
		"internal/todo/store_test.go",
		"docs/reference/todo.md",
		".chr/default/config.yml",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered code context missing %q:\n%s", want, rendered)
		}
	}

	for _, unwanted := range []string{"runtime.log", ".git/config"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("rendered code context included %q:\n%s", unwanted, rendered)
		}
	}
}

func TestLoadTreeSearchContextDerivesTermsAndSearchesFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "internal", "todo", "store.go"), "func CompleteTask(id int) error { return nil }\n")
	writeTestFile(t, filepath.Join(dir, "internal", "todo", "store_test.go"), "func TestCompleteTask(t *testing.T) {}\n")
	writeTestFile(t, filepath.Join(dir, "docs", "reference", "todo.md"), "Mark task completed from the todo list.\n")

	ctx, err := LoadTreeSearchContext(TreeSearchContextOptions{
		ProjectDir: dir,
		Ticket: ticketcore.Indexed{
			IndexedIdentity: ticketcore.IndexedIdentity{Key: "TODO-2"}, IndexedSummary: ticketcore.IndexedSummary{Title: "Complete task"}, IndexedClassification: ticketcore.IndexedClassification{
				UseCase:     "todo list",
				ProjectArea: "todo",
				Tags:        []string{"backend"},
			},
		},
		ValidationCommand: "go test ./...",
		MaxSearchHits:     10,
	})
	if err != nil {
		t.Fatalf("LoadTreeSearchContext returned error: %v", err)
	}

	for _, want := range []string{"todo", "complete", "task", "backend", "test"} {
		if !containsString(ctx.Terms, want) {
			t.Fatalf("Terms = %#v, want %q", ctx.Terms, want)
		}
	}

	rendered := RenderCodeContext(ctx)
	for _, want := range []string{"internal/todo/store.go", "internal/todo/store_test.go", "docs/reference/todo.md"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered search context missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderCodeContextIsDeterministic(t *testing.T) {
	ctx := TreeSearchContext{
		ProjectDir: "/tmp/project",
		Tree: []ContextTreeEntry{
			{Path: "a.go"},
			{Path: "docs", IsDir: true},
		},
		Terms: []string{"task"},
		Hits: []ContextSearchHit{
			{Term: "task", Path: "a.go", Line: 1, Excerpt: "task"},
		},
	}

	first := RenderCodeContext(ctx)
	second := RenderCodeContext(ctx)

	if first != second {
		t.Fatalf("RenderCodeContext is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}
