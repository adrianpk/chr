package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func guidedIdeasFixture(t *testing.T) (string, AuthoringSession, GuidedUseCaseCommand) {
	t.Helper()

	root, finish := guidedCheckpointFixture(t)

	session, err := FinishGuidedUseCaseDump(t.Context(), root, finish)
	if err != nil {
		t.Fatal(err)
	}

	command := GuidedUseCaseCommand{GuidedUseCaseScope: finish.GuidedUseCaseScope, ExpectedRevision: session.Revision, OperationID: "later-idea", Action: "idea", Content: "Add a separately reviewed export later."}

	return root, session, command
}

// TestGuidedUseCaseIdeasSeparate retains ideas outside the frozen capture and
// preserves their original request through duplicate, stale and foreign replies.
func TestGuidedUseCaseIdeasSeparate(t *testing.T) {
	root, original, command := guidedIdeasFixture(t)

	saved, err := RetainGuidedIdea(t.Context(), root, command)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Guided.Capture != original.Guided.Capture || !reflect.DeepEqual(saved.Guided.Checkpoints, original.Guided.Checkpoints) || saved.Guided.Stage != original.Guided.Stage || saved.Guided.Status != original.Guided.Status {
		t.Fatal("idea replaced frozen inputs or interrupted progression")
	}

	assertGuidedIdeaProvenance(t, root, saved, original.Guided.Actor, command.Content)

	again, err := RetainGuidedIdea(t.Context(), root, command)
	if err != nil || again.Revision != saved.Revision {
		t.Fatal("duplicate idea repeated a write", err)
	}

	assertGuidedIdeaGuards(t, root, command)
}

func assertGuidedIdeaProvenance(t *testing.T, root string, saved AuthoringSession, actor, content string) {
	t.Helper()

	ideas, err := ReadGuidedIdeas(t.Context(), root, saved)
	if err != nil || len(ideas) != 1 || ideas[0].Content != content || ideas[0].State != "pending_review" || ideas[0].Actor != actor {
		t.Fatal("idea was not retained with provenance", err)
	}
}

func assertGuidedIdeaGuards(t *testing.T, root string, command GuidedUseCaseCommand) {
	t.Helper()

	for _, variant := range []string{"content", "stale", "namespace", "empty"} {
		changed := command

		switch variant {
		case "content":
			changed.Content = "Changed duplicate"
		case "stale":
			changed.OperationID = "stale"
		case "namespace":
			changed.Namespace = "foreign"
		case "empty":
			changed.Content = " "
		}

		if _, err := RetainGuidedIdea(t.Context(), root, changed); !errors.Is(err, ErrWorkflowConflict) {
			t.Fatal("invalid idea accepted", variant, err)
		}
	}
}

// TestGuidedUseCaseIdeasRecovery recovers a journal interrupted after either the
// separate idea entry or the session image without duplicating either object.
func TestGuidedUseCaseIdeasRecovery(t *testing.T) {
	for _, position := range []int{0, 2} {
		t.Run(string(rune('0'+position)), func(t *testing.T) {
			testGuidedIdeaRecovery(t, position)
		})
	}
}

func testGuidedIdeaRecovery(t *testing.T, position int) {
	t.Helper()

	root, session, command := guidedIdeasFixture(t)

	operation, err := guidedOperation(command)
	if err != nil {
		t.Fatal(err)
	}

	lost := errors.New("lost idea journal acknowledgement")
	r := &fileAuthoringSessionRepository{projectDir: root, afterMutationReplace: func(index int) error {
		if index == position {
			return lost
		}

		return nil
	}}

	err = WithAuthoringMutation(t.Context(), root, func(context.Context) error { return r.retainGuidedIdea(&session, operation) })
	if !errors.Is(err, lost) {
		t.Fatal("idea fault not reached", err)
	}

	recovered, err := RetainGuidedIdea(t.Context(), root, command)
	if err != nil || len(recovered.Guided.LaterIdeas) != 1 {
		t.Fatal("idea recovery duplicated or lost its reference", err)
	}

	if _, err := ReadGuidedIdeas(t.Context(), root, recovered); err != nil {
		t.Fatal(err)
	}
}

// TestGuidedUseCaseIdeasControl keeps deferred requests available during a stop,
// while preventing a retained control identity from being reused for an idea.
func TestGuidedUseCaseIdeasControl(t *testing.T) {
	root, session, idea := guidedIdeasFixture(t)
	pause := idea
	pause.Action, pause.Content, pause.OperationID = "pause", "", "pause-before-idea"

	stopped, err := ApplyGuidedUseCaseControl(t.Context(), root, pause, func(context.Context, *AuthoringSession) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	idea.ExpectedRevision = stopped.Revision

	idea.OperationID = pause.OperationID
	if _, err := RetainGuidedIdea(t.Context(), root, idea); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("idea reused an existing control identity", err)
	}

	idea.OperationID = "idea-during-pause"

	saved, err := RetainGuidedIdea(t.Context(), root, idea)
	if err != nil || GuidedDispatchAllowed(saved) || !reflect.DeepEqual(saved.Guided.Control, stopped.Guided.Control) || saved.Guided.Capture != session.Guided.Capture {
		t.Fatal("deferred idea changed the stop or frozen capture", err)
	}
}
