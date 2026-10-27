package core

import (
	"errors"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedDumpContent seals the final form body once, including after replay.
func TestGuidedDumpContent(t *testing.T) {
	root, command := guidedCheckpointFixture(t)
	command.Content = "A contributor needs accurate project onboarding instructions."

	first, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil || first.Guided.Capture != command.Content || len(first.Guided.Checkpoints) != 1 {
		t.Fatal("final dump input was not sealed", err)
	}

	replay, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil || replay.Revision != first.Revision || replay.Guided.Capture != command.Content {
		t.Fatal("replay repeated or replaced the capture", err)
	}

	checkpoint := first.Guided.Checkpoints[0]

	content, err := ReadGuidedCommittedArtifact(t.Context(), root, checkpoint, checkpoint.Artifacts[0].Path)
	if err != nil || !strings.Contains(string(content), command.Content) {
		t.Fatal("committed dump omitted the submitted body", err)
	}

	command.Content = "A changed request"
	if _, err := FinishGuidedUseCaseDump(t.Context(), root, command); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed replay replaced the agreed input", err)
	}
}
