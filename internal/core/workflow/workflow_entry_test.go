package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowEntry restores only equivalent unfinished frozen inputs. A new
// compatibility start after completion gets a distinct bounded invocation.
func TestWorkflowEntry(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	request.RequestKey = "entry-retry"
	request.Selection.Target.Commit = ""

	found, matched, err := store.FindWorkflowEntry(t.Context(), request)
	if err != nil || !matched || found.ID != run.Invocation.ID || found.Selection.Target != run.Invocation.Selection.Target {
		t.Fatal("entry lost frozen target", err)
	}

	checkWorkflowEntryChanges(t, store, hierarchy, request)
	workflowTestExecute(t, store, run.Invocation.ID, "worker")

	if _, matched, err := store.FindWorkflowEntry(t.Context(), request); err != nil || matched {
		t.Fatal("terminal entry reopened", err)
	}
}

// TestWorkflowLegacyEntry excludes both native Markdown and JSON-backed views
// before a legacy worker, including malformed adoption markers.
func TestWorkflowLegacyEntry(t *testing.T) {
	store, hierarchy := workflowSelectionFixture(t)

	tickets, err := store.WorkflowTickets(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	selected, excluded, err := FilterLegacyWorkTickets(t.Context(), tickets)
	if err != nil || len(selected) != 0 || excluded != len(hierarchy.Tasks) {
		t.Fatal("JSON projection retained legacy authority", err)
	}

	dir := t.TempDir()

	cases := []struct {
		name, content string
		adopted       bool
	}{
		{"legacy", "---\nkey: PRJ-90\nstatus: ready_for_dev\n---\nWork\n", false},
		{"native", "---\nkey: PRJ-90\nworkflow_schema: chr-workflow/v1\n---\nWork\n", true},
		{"invalid schema", "---\nkey: PRJ-90\nworkflow_schema: broken\n---\nWork\n", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name+".md")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}

			if err := CheckLegacyWorkTicket(path); errors.Is(err, ErrWorkflowAuthority) != tc.adopted {
				t.Fatalf("legacy guard: %v", err)
			}
		})
	}
}

// TestWorkflowEntryRepository retains public identity and refuses host aliases,
// credential-bearing URLs and local paths instead of storing transport guesses.
func TestWorkflowEntryRepository(t *testing.T) {
	store, _ := workflowSelectionFixture(t)
	project := store.layout.ProjectDir()

	cases := []struct{ name, remote, want string }{
		{"HTTPS", "https://forge.example/owner/repo.git", "https://forge.example/owner/repo"},
		{"alias", "ssh://forge@private:2222/owner/repo.git", ""},
		{"credential", "https://user:secret@forge.example/owner/repo", ""},
		{"local", filepath.Join(t.TempDir(), "remote.git"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workflowSelectionGit(t, store, "remote", "add", tc.name, tc.remote)

			got, err := EntryRepository(t.Context(), project, tc.name)
			if got != tc.want || (err == nil) != (tc.want != "") {
				t.Fatal("untrusted identity accepted", err)
			}
		})
	}
}

func checkWorkflowEntryChanges(t *testing.T, store *Store, hierarchy MaterializationResult, request StartRequest) {
	t.Helper()

	for _, change := range []string{"policy", "repository", "selector"} {
		t.Run(change, func(t *testing.T) {
			changed := request

			switch change {
			case "policy":
				changed.Policy.MaxWorkerAttempts = 2
			case "repository":
				changed.Selection.Target.Repository = "https://other.example/test/repo"
			case "selector":
				changed.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key
			}

			if _, matched, err := store.FindWorkflowEntry(t.Context(), changed); err != nil || matched {
				t.Fatal("entry replaced frozen inputs", err)
			}
		})
	}
}
