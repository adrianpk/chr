package workflow

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestNativePreparationIsolation preserves staged, unstaged and untracked
// content, checkout metadata and original evidence in separate worker/user lanes.
func TestNativePreparationIsolation(t *testing.T) {
	t.Setenv("GIT_OPTIONAL_LOCKS", "0")

	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { testNativePreparationIsolation(t, method) })
	}
}

func testNativePreparationIsolation(t *testing.T, method string) {
	t.Helper()

	f := newWorkflowGitFixture(t, method)
	user := f.request.Directory
	worker := filepath.Join(t.TempDir(), "original-worker")
	workflowIntegrationCommand(t, user, "worktree", "add", "-b", "original-worker", worker, f.request.Git.SourceCommit)
	nativeDirtyCheckout(t, user)
	nativeDirtyCheckout(t, worker)

	beforeUser, beforeWorker := nativeCheckoutSnapshot(t, user), nativeCheckoutSnapshot(t, worker)
	beforeRemote := workflowIntegrationCommand(t, f.remote, "show-ref")
	f.request.Directory = worker
	proof, lane := nativeCombinedPreparation(t, f)

	nativeCheckCheckout(t, user, beforeUser)
	nativeCheckCheckout(t, worker, beforeWorker)

	if workflowIntegrationCommand(t, f.remote, "show-ref") != beforeRemote {
		t.Fatal("preparation mutated canonical refs")
	}

	if workflowIntegrationCommand(t, lane.Worktree, "rev-parse", "HEAD") != proof.ResultCommit || workflowIntegrationCommand(t, lane.Worktree, "status", "--porcelain") != "" {
		t.Fatal("owned validation lane does not contain the exact clean result")
	}

	for _, name := range []string{"source.txt", "target.txt"} {
		if _, err := os.Stat(filepath.Join(lane.Worktree, name)); err != nil {
			t.Fatal("prepared result lost source or target content", err)
		}
	}
}

func nativeCheckCheckout(t *testing.T, directory string, before map[string]string) {
	t.Helper()

	after := nativeCheckoutSnapshot(t, directory)
	for name, content := range before {
		if after[name] != content {
			t.Fatalf("preparation mutated original checkout evidence: %s", name)
		}
	}

	if len(after) != len(before) {
		t.Fatal("preparation added or removed original checkout evidence")
	}
}

func nativeDirtyCheckout(t *testing.T, directory string) {
	t.Helper()

	path := filepath.Join(directory, "initial.txt")
	if err := os.WriteFile(path, []byte("Staged original intent\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, directory, "add", "initial.txt")

	for name, content := range map[string]string{"initial.txt": "Unstaged original intent\n", "original-evidence.json": "{\"candidate\":\"original\"}\n"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	fetch := workflowIntegrationCommand(t, directory, "rev-parse", "--path-format=absolute", "--git-path", "FETCH_HEAD")
	if err := os.WriteFile(fetch, []byte("Retained original fetch evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func nativeCheckoutSnapshot(t *testing.T, directory string) map[string]string {
	t.Helper()

	snapshot := map[string]string{"commit": workflowIntegrationCommand(t, directory, "rev-parse", "HEAD")}

	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		data, err := os.ReadFile(path)
		snapshot[path] = string(data)

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"HEAD", "index", "FETCH_HEAD", "ORIG_HEAD"} {
		path := workflowIntegrationCommand(t, directory, "rev-parse", "--path-format=absolute", "--git-path", name)

		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}

		if err == nil {
			snapshot["git:"+name] = string(data)
		}
	}

	return snapshot
}
