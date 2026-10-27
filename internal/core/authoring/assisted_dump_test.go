package authoring

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func assistedDumpTestOptions(t *testing.T, root, id string) AssistedArtifactSynthesisOptions {
	t.Helper()

	repository := NewAuthoringSessionRepository(root)
	session := testAssistedContextSession(root)
	session.ID = id
	path := filepath.Join(AuthoringSessionLayout(root, id).SessionDir, "curated-summary.md")
	ref := ArtifactRef{ID: "curated-summary-" + id, Kind: ArtifactKindDraft, Role: ArtifactRoleOutput, Path: filepath.ToSlash(path)}

	session.Outputs = []ArtifactRef{ref}
	if _, err := repository.CreateSession(t.Context(), session); err != nil {
		t.Fatal(err)
	}

	summary := BuildAssistedCuratedAnswerSummary(id, nil, session.CreatedAt)
	summary.ModelSummary = id + " prepared material."
	writeTestFile(t, path, RenderAssistedCuratedAnswerSummary(summary))

	return AssistedArtifactSynthesisOptions{ProjectDir: root, Session: session, CuratedSummary: summary, SourceRef: ref, TargetKind: ArtifactKindDump, Repository: repository}
}

// Assisted and ordinary captures serialize through one canonical identity; a
// semantic retry returns its original receipt despite later source/body edits.
func TestAssistedDumpCapture(t *testing.T) {
	root := t.TempDir()
	opts := assistedDumpTestOptions(t, root, "first")

	first, err := SynthesizeAssistedArtifact(t.Context(), opts)
	if err != nil || first.Mutation == nil {
		t.Fatal("first synthesis failed", first, err)
	}

	if _, err := MutateDump(t.Context(), root, dumpCaptureRequest("ordinary", "Ordinary material."), nil); err != nil {
		t.Fatal(err)
	}

	opts.CuratedSummary.GeneratedAt = time.Now().UTC()
	writeTestFile(t, opts.SourceRef.Path, "Independent source edit.")

	replay, err := SynthesizeAssistedArtifact(t.Context(), opts)
	if err != nil || !reflect.DeepEqual(first.Mutation, replay.Mutation) || replay.Content != first.Content {
		t.Fatal("semantic retry duplicated output or lost proof", replay, err)
	}

	checkAssistedDumpCapture(t, root, first.Layout.Ref.Path)
}

// Independent sessions retain their own prepared sources while extending the
// same legacy destination across namespace changes and concurrent writers.
func TestAssistedDumpConcurrent(t *testing.T) {
	root := t.TempDir()
	reference := projectDumpTestArtifact(t, root, "legacy", "Retained", "# Original material\n")
	options := []AssistedArtifactSynthesisOptions{assistedDumpTestOptions(t, root, "first"), assistedDumpTestOptions(t, root, "second")}
	t.Setenv(opsnamespace.UserEnv, "other")

	var group sync.WaitGroup
	for _, opts := range options {
		group.Go(func() {
			result, err := SynthesizeAssistedArtifact(t.Context(), opts)
			if err != nil || result.Layout.Ref.Path != reference {
				t.Error("concurrent synthesis changed identity", result, err)
			}
		})
	}

	group.Wait()

	document, err := ReadDumpDocument(t.Context(), root, reference)
	if err != nil {
		t.Fatal(err)
	}

	for _, text := range []string{"Original material", "first prepared material.", "second prepared material."} {
		if strings.Count(document.Content, text) != 1 {
			t.Fatal("concurrent contribution lost or duplicated", document.Content)
		}
	}
}

// Every changed publication image, including session and receipt, recovers as
// one transaction. The source guard is a no-op journal image, never a rewrite.
func TestAssistedDumpRecovery(t *testing.T) {
	for _, test := range []struct {
		name   string
		faults []int
	}{{"first", []int{0, 1, 2, 3, 4, 5, 6, 7, 9}}, {"later", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 10}}} {
		for _, fault := range test.faults {
			t.Run(test.name+"-"+string(rune('a'+fault)), func(t *testing.T) {
				checkAssistedDumpRecovery(t, test.name, fault)
			})
		}
	}
}

func checkAssistedDumpRecovery(t *testing.T, mode string, fault int) {
	t.Helper()

	root := t.TempDir()
	if mode == "later" {
		if _, err := MutateDump(t.Context(), root, dumpCaptureRequest("initial", "Initial material."), nil); err != nil {
			t.Fatal(err)
		}
	}

	opts := assistedDumpTestOptions(t, root, "session")
	opts.Repository = &fileAuthoringSessionRepository{projectDir: root, afterMutationReplace: func(index int) error {
		if index == fault {
			return errors.New("interrupted publication")
		}

		return nil
	}}

	first, err := SynthesizeAssistedArtifact(t.Context(), opts)
	if err == nil {
		t.Fatal("fault not exercised", mode, fault)
	}

	opts.Repository = NewAuthoringSessionRepository(root)

	replayed, err := SynthesizeAssistedArtifact(t.Context(), opts)
	if err != nil || !reflect.DeepEqual(first.Mutation, replayed.Mutation) {
		t.Fatal("publication did not recover original result", replayed, err)
	}

	session, err := opts.Repository.GetSession(t.Context(), opts.Session.ID)
	if err != nil || !assistedSessionHasOutput(session, first.Layout.Ref.Path) || len(session.Provenance) != 1 {
		t.Fatal("dump published without its session/provenance", session, err)
	}
}

// Changed arguments, stale session images and prepared sources cannot silently
// replace retained work; another owner cannot reuse a pending request identity.
func TestAssistedDumpGuards(t *testing.T) {
	for _, name := range []string{"session", "source", "reserved", "arguments"} {
		t.Run(name, func(t *testing.T) {
			checkAssistedDumpGuard(t, name)
		})
	}
}

func checkAssistedDumpGuard(t *testing.T, name string) {
	t.Helper()

	root := t.TempDir()
	opts := assistedDumpTestOptions(t, root, "session")
	opts.OperationID = "retained"

	switch name {
	case "session":
		changed := opts.Session

		changed.Description = "Independent description."
		if _, err := opts.Repository.UpdateSession(t.Context(), changed); err != nil {
			t.Fatal(err)
		}
	case "source":
		writeTestFile(t, opts.SourceRef.Path, "Independent source.")
	case "reserved":
		writeTestFile(t, filepath.Join(root, authoringRequestRef(opts.OperationID)), "reserved identity")
	case "arguments":
		if _, err := SynthesizeAssistedArtifact(t.Context(), opts); err != nil {
			t.Fatal(err)
		}

		opts.CuratedSummary.ModelSummary = "Changed input."
	}

	if _, err := SynthesizeAssistedArtifact(t.Context(), opts); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed or reserved input accepted", name, err)
	}
}

// A pending transaction retains all sources when an independent edit diverges.
// Recovery prechecks the guarded source before publishing any remaining image.
func TestAssistedDumpDivergence(t *testing.T) {
	root := t.TempDir()
	opts := assistedDumpTestOptions(t, root, "session")

	opts.Repository = &fileAuthoringSessionRepository{projectDir: root, afterMutationReplace: func(int) error { return errors.New("interrupt") }}
	if _, err := SynthesizeAssistedArtifact(t.Context(), opts); err == nil {
		t.Fatal("fault not exercised")
	}

	writeTestFile(t, opts.SourceRef.Path, "Independent source.")

	opts.Repository = NewAuthoringSessionRepository(root)
	if _, err := opts.Repository.GetSession(t.Context(), opts.Session.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("divergent source did not block recovery", err)
	}

	if _, err := os.Stat(filepath.Join(root, authoringRuntimeRef("pending.json"))); err != nil {
		t.Fatal("pending transaction discarded", err)
	}
}

func checkAssistedDumpCapture(t *testing.T, root, reference string) {
	t.Helper()

	observed, err := ObserveProjectDump(t.Context(), root)
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 2 || observed.Destination != reference {
		t.Fatal("synthesis fragmented destination", observed, err)
	}

	document, err := ReadDumpDocument(t.Context(), root, reference)
	if err != nil || strings.Count(document.Content, "first prepared material.") != 1 || !strings.Contains(document.Content, "Ordinary material.") {
		t.Fatal("capture material lost or duplicated", document, err)
	}
}

// A copied receipt cannot replay in another project, even with identical session
// and summary inputs and the same explicit operation identity.
func TestAssistedDumpRoot(t *testing.T) {
	root := t.TempDir()
	opts := assistedDumpTestOptions(t, root, "session")
	opts.OperationID = "same-operation"

	if _, err := SynthesizeAssistedArtifact(t.Context(), opts); err != nil {
		t.Fatal(err)
	}

	path := authoringReceiptRef(opts.OperationID)

	receipt, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}

	other := assistedDumpTestOptions(t, t.TempDir(), "session")
	other.OperationID = opts.OperationID
	writeTestFile(t, filepath.Join(other.ProjectDir, path), string(receipt))

	if _, err := SynthesizeAssistedArtifact(t.Context(), other); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("receipt crossed project roots", err)
	}
}
