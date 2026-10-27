package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Independent bound sessions and ordinary captures share the same identity and
// serialize their contributions, while raw session inputs retain their namespace.
func TestAuthoringDumpCapture(t *testing.T) {
	root := t.TempDir()

	initial, err := MutateDump(t.Context(), root, dumpCaptureRequest("initial", "Initial material."), nil)
	if err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for _, namespace := range []string{"alpha", "beta"} {
		group.Go(func() { captureAuthoringDumpNamespace(t, root, namespace, initial.Destination) })
	}

	group.Wait()

	observed, err := ObserveProjectDump(t.Context(), root)
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 3 {
		t.Fatal("canonical inventory fragmented", observed, err)
	}

	document, err := ReadDumpDocument(t.Context(), root, initial.Destination)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"Initial material.", "alpha contribution.", "beta contribution."} {
		if strings.Count(document.Content, want) != 1 {
			t.Fatal("lost or duplicated contribution", document.Content)
		}
	}
}

// Every publication replacement is recoverable through a fresh owner, including
// the project pointer and the receipt; retry never calls the provider again.
func TestAuthoringDumpRecovery(t *testing.T) {
	for _, test := range []struct {
		name  string
		count int
	}{{"first", 8}, {"later", 9}, {"refine", 7}} {
		for fault := range test.count {
			if test.name == "refine" && fault == 1 {
				continue // Refinement does not change the transcript image.
			}

			t.Run(fmt.Sprintf("%s-%d", test.name, fault), func(t *testing.T) {
				checkAuthoringDumpRecovery(t, test.name, fault)
			})
		}
	}
}

func checkAuthoringDumpRecovery(t *testing.T, mode string, fault int) {
	t.Helper()

	root := t.TempDir()
	repo := &fileAuthoringSessionRepository{projectDir: root}
	request := authoringDumpRecoveryRequest(t, repo, mode)

	repo.afterMutationReplace = func(index int) error {
		if index == fault {
			return errors.New("interrupted publication")
		}

		return nil
	}
	calls := 0
	generate := func(ctx context.Context, snapshot AuthoringSnapshot, request AuthoringContinuationRequest) (AuthoringDraft, error) {
		calls++

		return telegramAuthoringOwnerDraft(ctx, snapshot, request)
	}

	first, err := repo.ContinueAuthoring(t.Context(), request, generate)
	if err == nil {
		t.Fatal("publication fault not exercised", mode, fault)
	}

	reopened := NewAuthoringSessionRepository(root)
	if _, err := reopened.ObserveContinuation(t.Context(), first.Session.ID); err != nil {
		t.Fatal("session/dump recovery failed", err)
	}

	replay, err := reopened.ContinueAuthoring(t.Context(), request, generate)
	if err != nil || replay.SourceHash != first.SourceHash || calls != 1 {
		t.Fatal("receipt retry regenerated content", replay, calls, err)
	}

	observed, err := ObserveProjectDump(t.Context(), root)
	if err != nil || len(observed.Candidates) != 1 || observed.Project.ProjectDumpRef != first.Session.Outputs[0].Path {
		t.Fatal("recovery lost canonical pointer", observed, err)
	}
}

// A source guard covers edits made by another adapter before provider dispatch.
// Re-observation permits replacement and preserves the exact prior body/history.
func TestAuthoringDumpRefine(t *testing.T) {
	root := t.TempDir()
	repo := NewAuthoringSessionRepository(root)
	first := telegramAuthoringCreateOwner(t, repo, "initial")
	reference := first.Session.Outputs[0].Path

	if _, err := MutateDump(t.Context(), root, dumpCaptureRequest("outside", "Outside contribution."), nil); err != nil {
		t.Fatal(err)
	}

	request := telegramAuthoringNextOwner(first, "stale", "refine", authoringcore.ArtifactKindDump)
	if _, err := repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		t.Fatal("stale continuation invoked its provider")

		return AuthoringDraft{}, nil
	}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale shared body accepted", err)
	}

	before, err := ReadDumpDocument(t.Context(), root, reference)
	if err != nil {
		t.Fatal(err)
	}

	current, err := repo.ObserveContinuation(t.Context(), first.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	request.Binding.RequestID, request.ExpectedSourceHash = "current", current.SourceHash
	if _, err := repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		return AuthoringDraft{Body: "Replacement context."}, nil
	}); err != nil {
		t.Fatal(err)
	}

	after, err := ReadDumpDocument(t.Context(), root, reference)
	if err != nil || !strings.HasSuffix(after.Content, "Replacement context.") || strings.Contains(after.Content, "Outside contribution.") {
		t.Fatal("refine did not replace editable body", after, err)
	}

	checkAuthoringDumpHistory(t, root, before)
}

// A retained question's own raw input can bootstrap the first general output;
// unrelated orphan captures continue to require explicit reconciliation.
func TestAuthoringDumpQuestion(t *testing.T) {
	root := t.TempDir()
	repo := NewAuthoringSessionRepository(root)
	request := telegramAuthoringOwnerRequest()

	first, err := repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		return AuthoringDraft{Question: "Which limit?"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	rawPath := filepath.Join(root, first.Session.Inputs[0].Path)

	raw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatal(err)
	}

	next := telegramAuthoringNextOwner(first, "response", "respond", authoringcore.ArtifactKindDump)
	next.Question, next.Instructions = first.Question, "Five attempts."

	result, err := repo.ContinueAuthoring(t.Context(), next, telegramAuthoringOwnerDraft)
	if err != nil || len(result.Session.Outputs) != 1 {
		t.Fatal("question continuation did not publish first dump", result, err)
	}

	retained, err := os.ReadFile(rawPath)
	if err != nil || string(raw) != string(retained) {
		t.Fatal("raw input rewritten", err)
	}

	if _, err := MutateDump(t.Context(), root, dumpCaptureRequest("ordinary", "More context."), nil); err != nil {
		t.Fatal("ordinary capture cannot use bound destination", err)
	}
}

// The returned guard must be immediately usable after publication, including a
// shared destination outside the session's namespace.
func TestAuthoringDumpSnapshot(t *testing.T) {
	repo := NewAuthoringSessionRepository(t.TempDir())
	request := telegramAuthoringOwnerRequest()
	request.Namespace = "team"

	result, err := repo.ContinueAuthoring(t.Context(), request, telegramAuthoringOwnerDraft)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := repo.ObserveContinuation(t.Context(), result.Session.ID)
	if err != nil || snapshot.SourceHash != result.SourceHash {
		t.Fatalf("published guard mismatch: receipt=%s snapshot=%+v err=%v", result.SourceHash, snapshot, err)
	}
}

// Reconciliation retains an old output reference while observation, refinement
// and work descriptions use its resolved canonical body in another namespace.
func TestAuthoringDumpAlias(t *testing.T) {
	root := t.TempDir()
	repo := NewAuthoringSessionRepository(root)
	first := telegramAuthoringCreateOwner(t, repo, "initial")
	old := first.Session.Outputs[0].Path
	destination := reconcileAuthoringDump(t, root)

	observed, err := repo.ObserveContinuation(t.Context(), first.Session.ID)
	if err != nil || authoringArtifactObservation(observed, destination) == nil || authoringArtifactObservation(observed, old) != nil {
		t.Fatal("session context did not resolve retained alias", observed, err)
	}

	request := telegramAuthoringNextOwner(first, "alias-refine", "refine", authoringcore.ArtifactKindDump)
	request.ExpectedSourceHash = observed.SourceHash

	result, err := repo.ContinueAuthoring(t.Context(), request, telegramAuthoringOwnerDraft)
	if err != nil || len(result.Session.Outputs) != 2 || result.Session.Outputs[1].Path != destination {
		t.Fatal("refinement replaced alias instead of canonical body", result, err)
	}

	if _, err := repo.ObserveContinuation(t.Context(), first.Session.ID); err != nil {
		t.Fatal("updated session cannot reopen", err)
	}

	original, err := os.ReadFile(filepath.Join(root, old))
	if err != nil || !strings.Contains(string(original), "Keep bounded retries.") {
		t.Fatal("retained alias body overwritten", err)
	}
}

// Independent file edits during provider preparation are detected before any
// session or dump image is published, even when they bypass the shared owner.
func TestAuthoringDumpProviderGuard(t *testing.T) {
	root := t.TempDir()

	first, err := MutateDump(t.Context(), root, dumpCaptureRequest("initial", "Initial material."), nil)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewAuthoringSessionRepository(root)
	request := telegramAuthoringOwnerRequest()

	_, err = repo.ContinueAuthoring(t.Context(), request, func(ctx context.Context, snapshot AuthoringSnapshot, request AuthoringContinuationRequest) (AuthoringDraft, error) {
		if err := os.WriteFile(filepath.Join(root, first.Destination), []byte("Independent edited image."), 0o600); err != nil {
			t.Fatal(err)
		}

		return telegramAuthoringOwnerDraft(ctx, snapshot, request)
	})
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed preparation source accepted", err)
	}

	content, err := os.ReadFile(filepath.Join(root, first.Destination))
	if err != nil || string(content) != "Independent edited image." {
		t.Fatal("independent source overwritten", err)
	}

	sessions, err := repo.ListSessions(t.Context())
	if err != nil || len(sessions) != 0 {
		t.Fatal("partial session published ahead of dump", sessions, err)
	}
}

// A new session cannot claim unrelated orphaned raw captures to bypass the
// explicit reconciliation boundary.
func TestAuthoringDumpOrphan(t *testing.T) {
	root := t.TempDir()
	projectDumpTestArtifact(t, root, "team", "Orphan", "---\ntype: dump_raw\n---\nRetained original.\n")

	repo := NewAuthoringSessionRepository(root)

	_, err := repo.ContinueAuthoring(t.Context(), telegramAuthoringOwnerRequest(), func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		t.Fatal("unresolved ownership invoked provider")

		return AuthoringDraft{}, nil
	})
	if !errors.Is(err, ErrProjectDumpReconciliation) {
		t.Fatal("orphaned input adopted implicitly", err)
	}
}

func captureAuthoringDumpNamespace(t *testing.T, root, namespace, destination string) {
	t.Helper()

	request := telegramAuthoringOwnerRequest()
	request.Binding.RequestID, request.Namespace = namespace, namespace

	result, err := NewAuthoringSessionRepository(root).ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		return AuthoringDraft{Body: namespace + " contribution."}, nil
	})
	if err != nil || len(result.Session.Outputs) != 1 || result.Session.Outputs[0].Path != destination {
		t.Error("bound session allocated another dump", result, err)
	}
}

func authoringDumpRecoveryRequest(t *testing.T, repo AuthoringSessionRepository, mode string) AuthoringContinuationRequest {
	t.Helper()

	request := telegramAuthoringOwnerRequest()
	if mode == "first" {
		return request
	}

	previous := telegramAuthoringCreateOwner(t, repo, "initial")
	if mode == "refine" {
		return telegramAuthoringNextOwner(previous, "refine", "refine", authoringcore.ArtifactKindDump)
	}

	return request
}

func checkAuthoringDumpHistory(t *testing.T, root string, before DumpDocumentObservation) {
	t.Helper()

	history, err := os.ReadFile(filepath.Join(root, projectDumpHistoryRef(before.Artifact), "artifact.md"))
	if err != nil || string(history) != before.Content {
		t.Fatal("refinement lost prior bytes", err)
	}
}

func reconcileAuthoringDump(t *testing.T, root string) string {
	t.Helper()

	destination := projectDumpTestArtifact(t, root, "team", "Earlier", "# Earlier material\n")

	preview, err := PreviewProjectDumpReconciliation(t.Context(), root)
	if err != nil || preview.Destination != destination {
		t.Fatal("reconciliation fixture has wrong target", preview, err)
	}

	if _, err := ReconcileProjectDump(t.Context(), root, ProjectDumpReconciliationRequest{OperationID: "reconcile", ExpectedRevision: preview.Revision, Destination: destination}); err != nil {
		t.Fatal(err)
	}

	return destination
}

// A sole legacy general document is adopted in its existing namespace; provider
// context can read it before the pointer is published with the session.
func TestAuthoringDumpAdopt(t *testing.T) {
	root := t.TempDir()
	reference := projectDumpTestArtifact(t, root, "legacy", "Existing", "# Existing context\nRetain this material.\n")
	repo := NewAuthoringSessionRepository(root)
	request := telegramAuthoringOwnerRequest()
	request.Namespace = "another"

	result, err := repo.ContinueAuthoring(t.Context(), request, func(ctx context.Context, snapshot AuthoringSnapshot, request AuthoringContinuationRequest) (AuthoringDraft, error) {
		if authoringArtifactObservation(snapshot, reference) == nil {
			t.Fatal("existing canonical context was omitted")
		}

		return telegramAuthoringOwnerDraft(ctx, snapshot, request)
	})
	if err != nil || result.Session.Outputs[0].Path != reference {
		t.Fatal("bound capture did not adopt existing identity", result, err)
	}

	project, err := ObserveProjectDump(t.Context(), root)
	if err != nil || project.Project.ProjectDumpRef != reference {
		t.Fatal("adoption did not publish pointer", project, err)
	}
}

// Refinement requires this session's observed output, rather than creating one
// from a pending raw input or borrowing another session's dump context.
func TestAuthoringDumpMissingOutput(t *testing.T) {
	repo := NewAuthoringSessionRepository(t.TempDir())

	first, err := repo.ContinueAuthoring(t.Context(), telegramAuthoringOwnerRequest(), func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		return AuthoringDraft{Question: "Which limit?"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	answer := telegramAuthoringNextOwner(first, "answer", "answer", authoringcore.ArtifactKindDump)
	answer.Question, answer.Instructions = first.Question, "Five attempts."

	answered, err := repo.ContinueAuthoring(t.Context(), answer, nil)
	if err != nil {
		t.Fatal(err)
	}

	request := telegramAuthoringNextOwner(answered, "refine", "refine", authoringcore.ArtifactKindDump)

	_, err = repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		t.Fatal("refinement without an output invoked provider")

		return AuthoringDraft{}, nil
	})
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("refinement created a new output", err)
	}
}
