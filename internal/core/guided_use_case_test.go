package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
)

// TestGuidedUseCaseCaptureStart preserves identity and rejects changed start scope.
func TestGuidedUseCaseCaptureStart(t *testing.T) {
	root := t.TempDir()
	scope := GuidedUseCaseScope{RunID: "capture", ProjectID: "project", Namespace: "default"}
	request := GuidedUseCaseStart{GuidedUseCaseScope: scope, Actor: "operator", ProgressionIntent: true}

	first, err := StartGuidedUseCase(t.Context(), root, request)
	if err != nil {
		t.Fatal(err)
	}

	again, err := StartGuidedUseCase(t.Context(), root, request)
	if err != nil || !reflect.DeepEqual(first, again) {
		t.Fatalf("replay changed session: %v", err)
	}

	variants := map[string]GuidedUseCaseStart{"project": request, "namespace": request, "actor": request, "intent": request}
	changed := request
	changed.ProjectID = "other"
	variants["project"] = changed
	changed = request
	changed.Namespace = "other"
	variants["namespace"] = changed
	changed = request
	changed.Actor = "other"
	variants["actor"] = changed
	changed = request
	changed.ProgressionIntent = false

	variants["intent"] = changed
	for name, changed := range variants {
		t.Run(name, func(t *testing.T) {
			if _, err := StartGuidedUseCase(t.Context(), root, changed); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatalf("changed start accepted: %v", err)
			}
		})
	}
}

// TestGuidedUseCaseCaptureManual preserves legacy sessions and prevents implicit adoption.
func TestGuidedUseCaseCaptureManual(t *testing.T) {
	root, first := guidedCaptureFixture(t)

	first.ID, first.Guided = "manual", nil
	if _, err := NewAuthoringSessionRepository(root).CreateSession(t.Context(), first); err != nil {
		t.Fatal(err)
	}

	scope := GuidedUseCaseScope{RunID: first.ID, ProjectID: first.ProjectID, Namespace: first.Namespace}
	if _, err := StartGuidedUseCase(t.Context(), root, GuidedUseCaseStart{GuidedUseCaseScope: scope, Actor: "operator", ProgressionIntent: true}); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("adopted manual session: %v", err)
	}

	saved, err := NewAuthoringSessionRepository(root).GetSession(t.Context(), first.ID)
	if err != nil || saved.Guided != nil {
		t.Fatalf("manual session changed: %v", err)
	}
}

// TestGuidedUseCaseCaptureWrites allows one concurrent revision and replays its receipt.
func TestGuidedUseCaseCaptureWrites(t *testing.T) {
	root, session := guidedCaptureFixture(t)
	scope := GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}
	command := GuidedUseCaseCommand{GuidedUseCaseScope: scope, OperationID: "first", ExpectedRevision: session.Revision, Action: "append", Content: "A meaningful requirement."}
	assertGuidedConcurrentWrites(t, root, command)

	current, err := ReadGuidedUseCase(t.Context(), root, scope)
	if err != nil {
		t.Fatal(err)
	}

	command.OperationID = current.Guided.CompletedOperations[len(current.Guided.CompletedOperations)-1].ID

	again, err := MutateGuidedUseCase(t.Context(), root, command)
	if err != nil || !reflect.DeepEqual(current, again) {
		t.Fatalf("duplicate append: %v", err)
	}

	command.Content = "Replacement"
	if _, err := MutateGuidedUseCase(t.Context(), root, command); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("changed replay accepted: %v", err)
	}
}

func assertGuidedConcurrentWrites(t *testing.T, root string, command GuidedUseCaseCommand) {
	t.Helper()

	var wait sync.WaitGroup

	results := make(chan error, 2)

	for _, id := range []string{"first", "second"} {
		wait.Add(1)
		go func() {
			defer wait.Done()

			request := command
			request.OperationID = id

			_, err := MutateGuidedUseCase(t.Context(), root, request)
			results <- err
		}()
	}

	wait.Wait()
	close(results)

	successes, conflicts := 0, 0

	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrWorkflowConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent writes: %d successes, %d conflicts", successes, conflicts)
	}
}

// TestGuidedUseCaseCaptureInspect does not allocate files for a read-only query.
func TestGuidedUseCaseCaptureInspect(t *testing.T) {
	root := t.TempDir()

	scope := GuidedUseCaseScope{RunID: "missing", ProjectID: "project", Namespace: "default"}
	if _, err := ReadGuidedUseCase(t.Context(), root, scope); !errors.Is(err, ErrAuthoringSessionNotFound) {
		t.Fatalf("missing query: %v", err)
	}

	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Fatalf("query created files: %v %v", entries, err)
	}
}

func guidedCaptureFixture(t *testing.T) (string, AuthoringSession) {
	t.Helper()

	root := t.TempDir()
	scope := GuidedUseCaseScope{RunID: "capture", ProjectID: "project", Namespace: "default"}

	session, err := StartGuidedUseCase(t.Context(), root, GuidedUseCaseStart{GuidedUseCaseScope: scope, Actor: "operator", ProgressionIntent: true})
	if err != nil {
		t.Fatal(err)
	}

	session, err = MutateGuidedUseCase(t.Context(), root, GuidedUseCaseCommand{GuidedUseCaseScope: scope, OperationID: "name", ExpectedRevision: session.Revision, Action: "name", Content: "Named work"})
	if err != nil {
		t.Fatal(err)
	}

	return root, session
}

// TestGuidedUseCaseCaptureRecovery replays partial owner writes without duplicate input.
func TestGuidedUseCaseCaptureRecovery(t *testing.T) {
	for _, phase := range []int{0, 1} {
		t.Run(fmt.Sprintf("replace-%d", phase), func(t *testing.T) {
			root, session := guidedCaptureFixture(t)
			command := GuidedUseCaseCommand{GuidedUseCaseScope: GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, OperationID: "append", ExpectedRevision: session.Revision, Action: "append", Content: "Retained requirement."}

			operation, err := guidedOperation(command)
			if err != nil {
				t.Fatal(err)
			}

			interrupted := errors.New("lost owner acknowledgement")
			r := &fileAuthoringSessionRepository{projectDir: root, afterMutationReplace: func(index int) error {
				if index == phase {
					return interrupted
				}

				return nil
			}}

			err = WithAuthoringMutation(t.Context(), root, func(ctx context.Context) error { return r.applyGuidedInput(ctx, &session, command, operation) })
			if !errors.Is(err, interrupted) {
				t.Fatalf("fault not reached: %v", err)
			}

			recovered, err := MutateGuidedUseCase(t.Context(), root, command)
			if err != nil || recovered.Guided.Capture != "Retained requirement.\n\n" {
				t.Fatalf("recovery duplicated or lost input: %v", err)
			}
		})
	}
}

// TestGuidedUseCaseCaptureLegacyJSON preserves checksum input for old manual receipts.
func TestGuidedUseCaseCaptureLegacyJSON(t *testing.T) {
	legacy := []byte(`{"ID":"legacy","Namespace":"","Revision":0,"ProjectID":"project","Status":"draft","Mode":"assisted","Target":"dump","Topic":"Legacy","Inputs":null,"Outputs":null,"Provenance":null,"CreatedAt":"2026-10-22T00:00:00Z","StartedAt":null,"FinishedAt":null,"UpdatedAt":"2026-10-22T00:00:00Z","Description":""}`)

	var session AuthoringSession
	if err := json.Unmarshal(legacy, &session); err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(session)
	if err != nil || !bytes.Equal(legacy, encoded) {
		t.Fatalf("legacy receipt checksum input changed: %v\n%s", err, encoded)
	}
}
