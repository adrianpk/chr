package core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkflowTrackerRestart repairs every interrupted export boundary without
// an execution journal, new domain revision or replacement of operator bytes.
func TestWorkflowTrackerRestart(t *testing.T) {
	for _, stage := range []int{0, 1, 2} {
		t.Run([]string{"pending", "output", "confirmation"}[stage], func(t *testing.T) {
			checkWorkflowTrackerRestart(t, stage)
		})
	}
}

func checkWorkflowTrackerRestart(t *testing.T, stage int) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	store.afterTrackerReplace = func(index int) error {
		if index == stage {
			return errors.New("lost response")
		}

		return nil
	}
	if _, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{}); err == nil {
		t.Fatal("injected interruption was ignored")
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)

	repaired, err := restarted.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil || repaired.Status != "current" {
		t.Fatal("restart required domain replay", repaired, err)
	}

	set, err := restarted.Read(t.Context(), hierarchy.Set.Path)
	if err != nil || set.Revision != hierarchy.Set.Revision {
		t.Fatal("repair changed authoritative revision", err)
	}
}

// TestWorkflowTrackerProcess exits an owned child after output replacement,
// leaving the pending projection marker without deferred cleanup.
func TestWorkflowTrackerProcess(t *testing.T) {
	project := os.Getenv("CHR_TRACKER_TEST_PROJECT")
	if project == "" {
		return
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: project, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	store.afterTrackerReplace = func(stage int) error {
		if stage == 1 {
			os.Exit(17)
		}

		return nil
	}
	if _, err := store.ExportWorkflowTracker(t.Context(), os.Getenv("CHR_TRACKER_TEST_SET"), WorkflowTrackerOptions{}); err != nil {
		t.Fatal(err)
	}

	t.Fatal("owned crash boundary not reached")
}

// TestWorkflowTrackerCrash resumes real process loss without a domain journal,
// another task allocation or a new invocation.
func TestWorkflowTrackerCrash(t *testing.T) {
	store, opts := workflowMaterializationFixture(t)

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	command := exec.Command(os.Args[0], "-test.run=^TestWorkflowTrackerProcess$")

	command.Env = append(os.Environ(), "CHR_TRACKER_TEST_PROJECT="+store.layout.ProjectDir(), "CHR_TRACKER_TEST_SET="+hierarchy.Set.ID)
	output, err := command.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned crash: %v: %s", err, output)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)

	stale, err := restarted.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
	if err != nil || stale.Status != "stale" {
		t.Fatal("lost confirmation did not remain stale", stale, err)
	}

	repaired, err := restarted.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil || repaired.Status != "current" || len(repaired.Sources) != 7 {
		t.Fatal("lost response did not repair", repaired, err)
	}
}

// TestWorkflowTrackerIO rejects unsafe or oversized output and edited metadata;
// broken derived paths must not prevent successful authoritative transitions.
func TestWorkflowTrackerIO(t *testing.T) {
	cases := []struct {
		name      string
		breakPath func(*testing.T, *WorkflowStore, WorkflowTrackerStatus)
	}{
		{"symlink", func(t *testing.T, s *WorkflowStore, p WorkflowTrackerStatus) {
			path := filepath.Join(s.layout.Root(), p.Path)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}

			if err := os.Symlink(filepath.Join(t.TempDir(), "foreign"), path); err != nil {
				t.Fatal(err)
			}
		}},
		{"oversize", func(t *testing.T, s *WorkflowStore, p WorkflowTrackerStatus) {
			workflowTrackerWrite(t, filepath.Join(s.layout.Root(), p.Path), []byte(strings.Repeat("x", WorkflowFileLimit+1)))
		}},
		{"metadata", func(t *testing.T, s *WorkflowStore, p WorkflowTrackerStatus) {
			path := workflowTrackerTestMetadata(t, p.SetID)
			workflowTrackerWrite(t, filepath.Join(s.layout.Root(), path), []byte(`{"schema":"edited"}`))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy, projection := workflowTrackerFixture(t)
			tc.breakPath(t, store, projection)

			if _, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{}); err == nil {
				t.Fatal("unsafe projection accepted")
			}

			task := hierarchy.Tasks[0]
			task.Body = append(task.Body, []byte("Preserved domain edit\n")...)

			written, err := store.Write(t.Context(), task, task.Revision)
			if err != nil || written.Revision != task.Revision+1 {
				t.Fatal("projection failure blocked domain progress", err)
			}
		})
	}
}

// TestWorkflowTrackerConcurrentEdit observes output again after pending metadata;
// an operator edit in that interval survives even with prior replace permission.
func TestWorkflowTrackerConcurrentEdit(t *testing.T) {
	store, hierarchy, projection := workflowTrackerFixture(t)
	task := hierarchy.Tasks[0]

	task.Body = append(task.Body, []byte("New description\n")...)
	if _, err := store.Write(t.Context(), task, task.Revision); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.layout.Root(), projection.Path)
	edited := []byte("Concurrent operator notes\n")
	store.afterTrackerReplace = func(index int) error {
		if index == 0 {
			workflowTrackerWrite(t, path, edited)
		}

		return nil
	}

	_, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: projection.OutputHash})
	if !errors.Is(err, ErrWorkflowConflict) || string(workflowTrackerRead(t, path)) != string(edited) {
		t.Fatal("concurrent edit was replaced", err)
	}
}

// TestWorkflowTrackerSourceChange detects uncoordinated source edits and missing
// output without writing during inspection; repair rebuilds only the view.
func TestWorkflowTrackerSourceChange(t *testing.T) {
	for _, source := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing output", true: "source body"}[source], func(t *testing.T) {
			store, hierarchy, projection := workflowTrackerFixture(t)

			path := filepath.Join(store.layout.Root(), projection.Path)
			if source {
				body := filepath.Join(store.layout.Root(), filepath.Dir(hierarchy.Tasks[0].Path), "artifact.md")
				workflowTrackerWrite(t, body, []byte("Edited task description\n"))
			} else if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}

			status, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
			if err != nil || status.Status != "stale" {
				t.Fatal("status trusted retained marker", status, err)
			}

			repaired, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
			if err != nil || repaired.Status != "current" || source && repaired.SourceHash == projection.SourceHash {
				t.Fatal("repair did not observe authoritative bytes", repaired, err)
			}
		})
	}
}
