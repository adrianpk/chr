package core

import (
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeSubmissionResponses retains the original unknown exchange when
// later evidence verifies Git application, without dispatching or charging again.
func TestNativeSubmissionResponses(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)
	receipt := nativeRecordReceipt(t, f.store, f.candidate)
	applied := nativeSubmissionApplied(t, f, submission, receipt)

	var result WorkflowGitPublicationResult
	if err := f.store.readWorkflowNativeEvidence(submission.InvocationID, *applied.Facts.TransportResponseRef, &result); err != nil {
		t.Fatal(err)
	}

	result.State, result.PublishedSourceCommit = "unknown", ""
	ref := retainNativeRecordEvidence(t, f.store, submission.InvocationID, "lost-response", result)
	unknown := applied
	unknown.Facts.ApplicationState, unknown.Facts.TransportResponseRef = "unknown", &ref

	retained, err := f.store.RetainWorkflowNativeSubmission(t.Context(), unknown)
	if err != nil {
		t.Fatal(err)
	}

	request := nativeSubmissionUpdate(t, f, retained)
	request.Facts.ApplicationState = "git_applied"

	request.Facts.TransportResponseRef = applied.Facts.TransportResponseRef
	if _, err := f.store.RetainWorkflowNativeSubmission(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("unknown response erased", err)
	}

	request.Facts.TransportResponseHistory = []WorkflowEvidenceRef{ref}

	retained, err = f.store.RetainWorkflowNativeSubmission(t.Context(), request)
	if err != nil || retained.State != "awaiting_observation" || len(retained.Native.TransportResponseHistory) != 1 {
		t.Fatal("reconciled evidence not retained", err)
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), submission.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("response consumed another attempt", err)
	}
}

// TestNativeSubmissionFences rejects a stale callback or replacement intent and
// leaves the actual common journal image intact for independent reconciliation.
func TestNativeSubmissionFences(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)
	request := nativeSubmissionUpdate(t, f, submission)

	run, err := f.store.readWorkflowInvocation(submission.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	path, err := workflowSubmissionPath(t.Context(), f.store.layout.ProjectDir(), run.Submission.JournalKey)
	if err != nil {
		t.Fatal(err)
	}

	before := mustReadNativeEntry(t, path)
	for _, test := range []struct {
		name   string
		change func(*WorkflowNativeSubmissionUpdate)
	}{
		{"revision", func(r *WorkflowNativeSubmissionUpdate) { r.InvocationRevision++; r.Facts.ApplicationState = "unknown" }},
		{"generation", func(r *WorkflowNativeSubmissionUpdate) { r.ClaimGeneration++; r.Facts.ApplicationState = "unknown" }},
		{"candidate", func(r *WorkflowNativeSubmissionUpdate) {
			r.Facts.CandidateHash = contenthash.Sum([]byte("replacement"))
		}},
		{"deadline", func(r *WorkflowNativeSubmissionUpdate) {
			r.Facts.ObservationDeadline = r.Facts.ObservationDeadline.Add(1)
		}},
		{"schema", func(r *WorkflowNativeSubmissionUpdate) { r.Facts.SchemaVersion = 2 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := request
			test.change(&changed)

			if _, err := f.store.RetainWorkflowNativeSubmission(t.Context(), changed); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale or changed intent accepted", err)
			}

			if after := mustReadNativeEntry(t, path); string(after) != string(before) {
				t.Fatal("invalid callback rewrote retained intent")
			}
		})
	}
}
