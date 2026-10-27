package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
)

// Explicit project instructions must not become a route to host files through
// links, a blocking FIFO, invalid text or an oversized resource.
func TestPiWorkerInstructions(t *testing.T) {
	for _, scenario := range []string{"text", "missing", "symlink", "hard link", "FIFO", "invalid text", "oversized"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			prepareWorkerInstructionFixture(t, root, scenario)

			data, err := WorkerRuntimeInstructions(root)
			if scenario == "text" || scenario == "missing" {
				if err != nil || (scenario == "text" && string(data) != "Project instructions") {
					t.Fatalf("explicit instructions: %q, %v", data, err)
				}
			} else if err == nil || len(data) != 0 {
				t.Fatal("unsafe project instructions were imported")
			}
		})
	}
}

func prepareWorkerInstructionFixture(t *testing.T, root, scenario string) {
	t.Helper()

	path := filepath.Join(root, "AGENTS.md")

	var err error

	switch scenario {
	case "missing":
		return
	case "symlink", "hard link":
		outside := filepath.Join(t.TempDir(), "outside")
		if err := os.WriteFile(outside, []byte("outside fixture"), 0o600); err != nil {
			t.Fatal(err)
		}

		if scenario == "symlink" {
			err = os.Symlink(outside, path)
		} else {
			err = os.Link(outside, path)
		}
	case "FIFO":
		err = syscall.Mkfifo(path, 0o600)
	case "invalid text":
		err = os.WriteFile(path, []byte{0xff}, 0o600)
	case "oversized":
		err = os.WriteFile(path, make([]byte, harness.MaxInput+1), 0o600)
	default:
		err = os.WriteFile(path, []byte("Project instructions"), 0o600)
	}

	if err != nil {
		t.Fatal(err)
	}
}

// Even an otherwise explicit checkout cannot grant the entire developer home.
func TestPiWorkerHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := WorkerRuntimeMounts(t.Context(), home); err == nil || !strings.Contains(err.Error(), "home") {
		t.Fatalf("home mount was accepted: %v", err)
	}
}
