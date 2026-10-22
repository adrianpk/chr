package core

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGuidedUseCaseCheckpointPreserves keeps the caller's branch, index and dirty work.
func TestGuidedUseCaseCheckpointPreserves(t *testing.T) {
	root, command := guidedCheckpointFixture(t)
	guidedWriteFile(t, root, "staged.txt", "staged change")
	guidedGitCommand(t, root, "add", "staged.txt")
	guidedWriteFile(t, root, "staged.txt", "unstaged change")
	guidedWriteFile(t, root, "untracked.txt", "untracked material")

	head := guidedGitCommand(t, root, "rev-parse", "HEAD")

	index, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}

	session, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil || !bytes.Equal(index, after) {
		t.Fatalf("caller index changed: %v", err)
	}

	if head != guidedGitCommand(t, root, "rev-parse", "HEAD") || guidedGitCommand(t, root, "symbolic-ref", "HEAD") != "refs/heads/dev" {
		t.Fatal("caller branch changed")
	}

	for path, expected := range map[string]string{"staged.txt": "unstaged change", "untracked.txt": "untracked material"} {
		content, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || string(content) != expected {
			t.Fatalf("unrelated %s changed: %v", path, err)
		}
	}

	assertGuidedCheckpointInput(t, root, session)
}

func assertGuidedCheckpointInput(t *testing.T, root string, session AuthoringSession) {
	t.Helper()

	if session.Guided.Stage != "quiz" || session.Guided.Status != "blocked" || len(session.Guided.Checkpoints) != 1 {
		t.Fatalf("incorrect sealed state: %#v", session.Guided)
	}

	checkpoint := session.Guided.Checkpoints[0]
	path := checkpoint.Artifacts[0].Path

	committed, err := ReadGuidedCommittedArtifact(t.Context(), root, checkpoint, path)
	if err != nil || !strings.Contains(string(committed), "Deliver a bounded workflow.") {
		t.Fatalf("missing committed capture: %v", err)
	}

	guidedWriteFile(t, root, path, "Later working-copy edit")

	retained, err := ReadGuidedCommittedArtifact(t.Context(), root, checkpoint, path)
	if err != nil || !bytes.Equal(committed, retained) {
		t.Fatalf("working copy replaced checkpoint: %v", err)
	}

	if _, err := ReadGuidedCommittedArtifact(t.Context(), root, checkpoint, "staged.txt"); err == nil {
		t.Fatal("read unrelated input")
	}
}

// TestGuidedUseCaseCheckpointRecovery retains the original commit after a lost response.
func TestGuidedUseCaseCheckpointRecovery(t *testing.T) {
	root, command := guidedCheckpointFixture(t)
	lost := errors.New("lost Git acknowledgement")
	r := &fileAuthoringSessionRepository{projectDir: root, afterGuidedCheckpoint: func() error { return lost }}

	err := WithAuthoringMutation(t.Context(), root, func(ctx context.Context) error { _, err := r.finishGuidedUseCaseDump(ctx, command); return err })
	if !errors.Is(err, lost) {
		t.Fatalf("checkpoint fault not reached: %v", err)
	}

	pending, err := ReadGuidedUseCase(t.Context(), root, command.GuidedUseCaseScope)
	if err != nil {
		t.Fatal(err)
	}

	original := guidedGitCommand(t, root, "rev-parse", pending.Guided.PendingCheckpoint.Ref)

	recovered, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil {
		t.Fatal(err)
	}

	if len(recovered.Guided.Checkpoints) != 1 || recovered.Guided.Checkpoints[0].Commit != original {
		t.Fatal("recovery created replacement commit")
	}

	replay, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil || replay.Revision != recovered.Revision {
		t.Fatalf("completion replay mutated run: %v", err)
	}
}

// TestGuidedUseCaseCheckpointUnavailable rejects missing objects and invalid content bindings.
func TestGuidedUseCaseCheckpointUnavailable(t *testing.T) {
	for _, variant := range []string{"hash", "missing", "corrupt"} {
		t.Run(variant, func(t *testing.T) {
			root, command := guidedCheckpointFixture(t)

			session, err := FinishGuidedUseCaseDump(t.Context(), root, command)
			if err != nil {
				t.Fatal(err)
			}

			checkpoint := session.Guided.Checkpoints[0]

			damageGuidedCheckpoint(t, root, &checkpoint, variant)

			if _, err := ReadGuidedCommittedArtifact(t.Context(), root, checkpoint, checkpoint.Artifacts[0].Path); err == nil {
				t.Fatal("invalid committed source accepted")
			}
		})
	}
}

// TestGuidedUseCaseCheckpointEmpty cannot turn an empty capture into a sealed stage.
func TestGuidedUseCaseCheckpointEmpty(t *testing.T) {
	root, session := guidedCaptureFixture(t)

	command := GuidedUseCaseCommand{GuidedUseCaseScope: GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, ExpectedRevision: session.Revision, OperationID: "finish", Action: "finish"}
	if _, err := FinishGuidedUseCaseDump(t.Context(), root, command); err == nil {
		t.Fatal("empty capture sealed")
	}
}

func guidedCheckpointFixture(t *testing.T) (string, GuidedUseCaseCommand) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git checkpoint fixture")
	}

	root, session := guidedCaptureFixture(t)
	guidedGitCommand(t, root, "init", "-b", "dev")
	guidedGitCommand(t, root, "config", "user.name", "Checkpoint Test")
	guidedGitCommand(t, root, "config", "user.email", "checkpoint@example.invalid")
	guidedWriteFile(t, root, "staged.txt", "baseline")
	guidedGitCommand(t, root, "add", "staged.txt")
	guidedGitCommand(t, root, "commit", "-m", "Initial fixture")

	scope := GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}

	session, err := MutateGuidedUseCase(t.Context(), root, GuidedUseCaseCommand{GuidedUseCaseScope: scope, OperationID: "capture-input", ExpectedRevision: session.Revision, Action: "append", Content: "Deliver a bounded workflow."})
	if err != nil {
		t.Fatal(err)
	}

	return root, GuidedUseCaseCommand{GuidedUseCaseScope: scope, ExpectedRevision: session.Revision, OperationID: "finish", Action: "finish"}
}

func guidedGitCommand(t *testing.T, root string, args ...string) string {
	t.Helper()

	command := exec.CommandContext(t.Context(), "git", append([]string{"-C", root}, args...)...)

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}

	return strings.TrimSpace(string(output))
}

func guidedWriteFile(t *testing.T, root, path, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path)), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func damageGuidedCheckpoint(t *testing.T, root string, checkpoint *GuidedStageCheckpoint, variant string) {
	t.Helper()

	switch variant {
	case "hash":
		checkpoint.Artifacts[0].Hash = strings.Repeat("0", 64)
	case "missing":
		checkpoint.Commit = strings.Repeat("0", 40)
	case "corrupt":
		object := filepath.Join(root, ".git", "objects", checkpoint.Commit[:2], checkpoint.Commit[2:])
		if err := os.Chmod(object, 0o600); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(object, []byte("corrupt object"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
