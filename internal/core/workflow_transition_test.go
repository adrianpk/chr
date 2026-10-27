package core

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowProcess is the owned subprocess used for lock and crash boundaries.
func TestWorkflowProcess(t *testing.T) {
	mode := os.Getenv("CHR_WORKFLOW_HELPER_MODE")
	if mode == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: os.Getenv("CHR_WORKFLOW_HELPER_DIR"), Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	s := NewWorkflowStore(layout, "project-1")

	ctx := context.Background()
	if mode == "hold" {
		err := s.withLock(ctx, func() error {
			if _, err := os.Stdout.WriteString("ready\n"); err != nil {
				return err
			}

			_, err := io.ReadFull(os.Stdin, make([]byte, 1))

			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	changes := []WorkflowChange{}

	for _, path := range strings.Split(os.Getenv("CHR_WORKFLOW_HELPER_PATHS"), "|") {
		r, err := s.Read(ctx, path)
		if err != nil {
			t.Fatal(err)
		}

		r.Task.Status = ticketcore.TicketStatusBacklog
		changes = append(changes, WorkflowChange{Record: r, ExpectedRevision: r.Revision})
	}

	s.afterReplace = func(int) error { os.Exit(17); return nil }
	if err := s.Transition(ctx, changes); err != nil {
		t.Fatal(err)
	}
}

func workflowProcess(t *testing.T, s *WorkflowStore, mode string, paths []string) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^TestWorkflowProcess$")

	cmd.Env = append(os.Environ(), "CHR_WORKFLOW_HELPER_MODE="+mode, "CHR_WORKFLOW_HELPER_DIR="+s.layout.ProjectDir(), "CHR_WORKFLOW_HELPER_PATHS="+strings.Join(paths, "|"))

	return cmd
}

// TestWorkflowCrash confirms matching replay and refuses incompatible manual edits.
func TestWorkflowCrash(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "recover", true: "conflict"}[conflict], func(t *testing.T) {
			s := workflowTestStore(t)
			a := workflowTestCreate(t, s, "CHR-1")
			b := workflowTestCreate(t, s, "CHR-2")

			workflowCrashWriter(t, s, []string{a.Path, b.Path})

			if conflict {
				workflowTestConflict(t, s, b)

				return
			}

			snapshot, err := s.Snapshot(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			for _, r := range snapshot.Records {
				if r.Revision != 2 || r.Task.Status != ticketcore.TicketStatusBacklog {
					t.Fatal("mixed transition survived recovery")
				}
			}
		})
	}
}

func workflowCrashWriter(t *testing.T, s *WorkflowStore, paths []string) {
	t.Helper()

	err := workflowProcess(t, s, "crash", paths).Run()

	var exit *exec.ExitError

	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("unexpected writer result: %v", err)
	}
}

func workflowTestCreate(t *testing.T, s *WorkflowStore, key string) WorkflowRecord {
	t.Helper()

	r, err := s.Create(context.Background(), key, workflowTestTask(key))
	if err != nil {
		t.Fatal(err)
	}

	return r
}

func workflowTestConflict(t *testing.T, s *WorkflowStore, r WorkflowRecord) {
	t.Helper()

	body := filepath.Join(s.layout.Root(), filepath.Dir(r.Path), "artifact.md")
	if err := os.WriteFile(body, []byte("manual edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Snapshot(context.Background()); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("incompatible image did not block: %v", err)
	}

	metadata, err := readWorkflowEnvelope(filepath.Join(s.layout.Root(), r.Path))
	if err != nil {
		t.Fatal(err)
	}

	if metadata.Revision != 1 {
		t.Fatal("conflicting replay applied remaining replacements")
	}
}

// TestWorkflowProcessLock proves exclusion by a separate process, not a Go mutex.
func TestWorkflowProcessLock(t *testing.T) {
	s := workflowTestStore(t)
	r := workflowTestCreate(t, s, "CHR-1")
	linked := workflowTestLinkedStore(t, s)
	cmd := workflowProcess(t, s, "hold", nil)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := stdin.Close(); err != nil {
			t.Error(err)
		}

		if err := cmd.Wait(); err != nil {
			t.Error(err)
		}
	})

	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		t.Fatal("lock holder was not ready")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := linked.Read(ctx, r.Path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("process-held lock failed exclusion: %v", err)
	}

	if _, err := stdin.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
}

func workflowTestLinkedStore(t *testing.T, store *WorkflowStore) *WorkflowStore {
	t.Helper()

	root := store.layout.ProjectDir()
	destination := filepath.Join(t.TempDir(), "linked")

	commands := [][]string{
		{"-C", root, "-c", "user.name=Example", "-c", "user.email=example@example.org", "commit", "--allow-empty", "-qm", "Fixture"},
		{"-C", root, "worktree", "add", "-qb", "linked", destination},
	}
	for _, argv := range commands {
		if output, err := exec.Command("git", argv...).CombinedOutput(); err != nil {
			t.Fatalf("fixture git command: %v: %s", err, output)
		}
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: destination, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, "project-1")
}

// TestWorkflowCompetingWriters admits one exact observed revision across goroutines.
func TestWorkflowCompetingWriters(t *testing.T) {
	s := workflowTestStore(t)
	r := workflowTestCreate(t, s, "CHR-1")
	results := make(chan error, 8)

	var group sync.WaitGroup

	for range 8 {
		group.Go(func() {
			copyOf := r
			task := *r.Task
			task.Status = ticketcore.TicketStatusBacklog
			copyOf.Task = &task

			_, err := s.Write(context.Background(), copyOf, 1)
			results <- err
		})
	}

	group.Wait()
	close(results)

	accepted := 0

	for err := range results {
		if err == nil {
			accepted++
		} else if !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal(err)
		}
	}

	if accepted != 1 {
		t.Fatalf("accepted %d competing writers", accepted)
	}
}
