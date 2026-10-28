package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeEntryCandidate binds a real prepared Git result to its admission;
// foreign or replaced evidence cannot silently authorize the next phase.
func TestNativeEntryCandidate(t *testing.T) {
	store, candidate := nativeRecordCandidate(t, "merge")
	run := workflowTestRead(t, store, candidate.InvocationID)
	admission := IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: candidate.Native.Unit, AdmissionKey: run.RequestKey, TargetRef: candidate.Native.TargetRef}
	entry := nativeEntryAdmit(t, store, admission)

	candidate.Native.EntryID = entry.ID

	var err error

	candidate.Native.CandidateHash, err = NativeCandidateHash(candidate)
	if err != nil {
		t.Fatal(err)
	}

	ref := retainNativeRecordEvidence(t, store, run.ID, "entry-candidate", candidate)
	request := nativeEntryTransition(t, store, entry, "preparing")
	request.CandidateRef = &ref

	entry, err = store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil || entry.CandidateRef == nil || *entry.CandidateRef != ref {
		t.Fatal("owned native candidate was not retained", err)
	}

	request = nativeEntryTransition(t, store, entry, "validating")
	before := workflowTestRead(t, store, run.ID)
	candidate.Native.EntryID = "another-unit"
	wrong := retainNativeRecordEvidence(t, store, run.ID, "foreign-entry-candidate", candidate)

	request.CandidateRef = &wrong
	if _, err := store.TransitionWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("foreign entry candidate accepted", err)
	}

	request.CandidateRef = nil

	path, err := store.safePath(ref.Path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("substituted bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.TransitionWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed candidate bytes accepted", err)
	}

	if after := workflowTestRead(t, store, run.ID); !reflect.DeepEqual(before, after) {
		t.Fatal("candidate rejection changed the authoritative pointer")
	}
}

// TestNativeEntryMalformed preserves unsupported and broken source bytes;
// admission cannot turn a missing queue record into a fresh sequence.
func TestNativeEntryMalformed(t *testing.T) {
	for _, failure := range []string{"version", "missing", "symlink"} {
		t.Run(failure, func(t *testing.T) {
			nativeEntryMalformed(t, failure)
		})
	}
}

func nativeEntryMalformed(t *testing.T, failure string) {
	t.Helper()

	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)

	path, err := store.safePath(workflowIntegrationEntryPath(entry.InvocationID, entry.ID))
	if err != nil {
		t.Fatal(err)
	}

	switch failure {
	case "version":
		nativeEntryUnsupported(t, store, path)
	case "missing":
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	case "symlink":
		outside := filepath.Join(t.TempDir(), "queue.json")
		if err := os.Rename(path, outside); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(outside, path); err != nil {
			t.Fatal(err)
		}
	}

	before := optionalNativeEntryBytes(t, path)
	if _, err := store.AdmitWorkflowIntegration(t.Context(), admission); err == nil {
		t.Fatal("missing or unsupported authority was re-admitted")
	}

	after := optionalNativeEntryBytes(t, path)
	if string(after) != string(before) {
		t.Fatal("blocked admission changed malformed source bytes")
	}
}

func optionalNativeEntryBytes(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}

	return data
}

func nativeEntryUnsupported(t *testing.T, store *Store, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}

	fields["schema_version"] = json.RawMessage(`2`)

	data, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}

	var entry IntegrationEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatal(err)
	}

	run, err := store.readWorkflowInvocation(entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	// Reseal the owner pointer so version dispatch, rather than a stale hash,
	// is the reason this retained future record blocks mutation.
	run.Integration.EntryRef.Hash = contenthash.Sum(data)
	run.IntegrationEntries[0] = *run.Integration

	change, err := workflowInvocationChange(run)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.commitFiles([]workflowstate.FileChange{{Path: run.Integration.EntryRef.Path, Before: contenthash.Sum(mustReadNativeEntry(t, path)), After: data}, change}); err != nil {
		t.Fatal(err)
	}
}

func mustReadNativeEntry(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// TestNativeEntryUnknown retains the existing common publication authority.
// Neither an early queue phase nor missing journal bytes permit cancellation.
func TestNativeEntryUnknown(t *testing.T) {
	f, provider := newWorkflowSubmissionFixture(t, "merge")
	run := workflowTestRead(t, f.store, f.run.ID)
	request := IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/" + run.PullRequest.Request.BaseBranch}
	entry := nativeEntryAdmit(t, f.store, request)

	_, dispatch, err := f.store.beginWorkflowSubmission(t.Context(), run.ID, provider, workflowSubmissionObserver(f))
	if err != nil || !dispatch || provider.calls != 0 {
		t.Fatal("fixture did not retain intent before external dispatch", err)
	}

	transition := nativeEntryTransition(t, f.store, entry, "cancelled")

	before := workflowTestRead(t, f.store, run.ID)
	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), transition); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("unknown publication was cancelled", err)
	}

	path, err := workflowSubmissionPath(t.Context(), f.store.layout.ProjectDir(), before.Submission.JournalKey)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), transition); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("missing journal was interpreted as no publication", err)
	}

	if after := workflowTestRead(t, f.store, run.ID); !reflect.DeepEqual(before, after) {
		t.Fatal("uncertain publication rejection changed retained consumption")
	}
}

// TestNativeEntryAlias binds the canonical repository identity while retaining
// the invocation's original frozen request and retry image.
func TestNativeEntryAlias(t *testing.T) {
	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("true"))
	request.Selection.Target.Repository = "https://FORGE.example/test/repo.git/"
	run := workflowTestStart(t, store, request).Invocation
	admission := IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}

	entry := nativeEntryAdmit(t, store, admission)
	if entry.Repository != "https://forge.example/test/repo" || nativeEntryAdmit(t, store, admission).ID != entry.ID {
		t.Fatal("repository alias did not retain one canonical target and admission")
	}
}
