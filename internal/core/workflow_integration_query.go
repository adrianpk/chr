package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// WorkflowIntegrationRecordError identifies retained data requiring reconciliation.
// Readers never replace the failing source image with a runnable default.
type WorkflowIntegrationRecordError struct {
	Path  string
	Cause error
}

func (e *WorkflowIntegrationRecordError) Error() string {
	return fmt.Sprintf("integration records require reconciliation at %s: %v", e.Path, e.Cause)
}

func (e *WorkflowIntegrationRecordError) Unwrap() error {
	return errors.Join(ErrWorkflowConflict, e.Cause)
}

// WorkflowIntegrationView resolves an entry's sealed ledger and immutable evidence.
type WorkflowIntegrationView struct {
	Entry        WorkflowIntegrationEntry
	Budget       WorkflowIntegrationBudget
	CurrentUsage WorkflowIntegrationUsage
}

// WorkflowPublicationBarrier is derived from a common journal, never its index.
type WorkflowPublicationBarrier struct {
	JournalKey   string
	JournalHash  string
	TargetKey    string
	Namespace    string
	InvocationID string
	Submission   WorkflowSubmission
}

// WorkflowIntegrationQueueQuery selects a bounded page of authoritative units.
// EligibleOnly means queued, without stop intent or uncertain target publication.
type WorkflowIntegrationQueueQuery struct {
	Repository   string
	TargetRef    string
	Namespace    string
	InvocationID string
	Unit         WorkflowRef
	States       []string
	EligibleOnly bool
	Offset       int
	Limit        int
	SnapshotHash string
}

// WorkflowIntegrationQueuePage is observational; it confers no execution claim.
type WorkflowIntegrationQueuePage struct {
	Entries      []WorkflowIntegrationView
	SnapshotHash string
	NextOffset   int
}

type workflowIntegrationSnapshot struct {
	Views    []WorkflowIntegrationView
	Journals []workflowSubmissionJournal
	Runs     []workflowInvocationFile
	Hash     string
}

// QueryWorkflowIntegrationQueue does not replay a namespace or publication journal.
func (s *WorkflowStore) QueryWorkflowIntegrationQueue(ctx context.Context, query WorkflowIntegrationQueueQuery) (page WorkflowIntegrationQueuePage, err error) {
	if err := validateWorkflowIntegrationQueueQuery(query); err != nil {
		return page, err
	}

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		snapshot, readErr := s.readWorkflowIntegrationSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		page, readErr = workflowIntegrationQueuePage(snapshot, query)

		return readErr
	})

	return page, integrationRecordError("repository", err)
}

func integrationRecordError(path string, err error) error {
	if err == nil {
		return nil
	}

	var retained *WorkflowIntegrationRecordError
	if errors.As(err, &retained) {
		return err
	}

	return &WorkflowIntegrationRecordError{Path: path, Cause: err}
}

func validateWorkflowIntegrationQueueQuery(q WorkflowIntegrationQueueQuery) error {
	if !workflowIntegrationQueryBounds(q) {
		return ErrWorkflowConflict
	}

	if q.Namespace != "" && !IsValidOpsNamespace(q.Namespace) {
		return ErrWorkflowConflict
	}

	for _, state := range q.States {
		if !validWorkflowIntegrationState(state) {
			return ErrWorkflowConflict
		}
	}

	if q.Repository != "" {
		if _, err := workflowCanonicalRepository(q.Repository); err != nil {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func workflowIntegrationQueuePage(snapshot workflowIntegrationSnapshot, q WorkflowIntegrationQueueQuery) (WorkflowIntegrationQueuePage, error) {
	page := WorkflowIntegrationQueuePage{SnapshotHash: snapshot.Hash, NextOffset: -1}
	if q.SnapshotHash != "" && q.SnapshotHash != snapshot.Hash {
		return page, ErrWorkflowConflict
	}

	views := []WorkflowIntegrationView{}

	for _, view := range snapshot.Views {
		if workflowIntegrationQueryMatches(view.Entry, q) && (!q.EligibleOnly || workflowIntegrationEligible(view.Entry, snapshot.Journals)) {
			views = append(views, view)
		}
	}

	start := min(q.Offset, len(views))
	end := min(start+q.Limit, len(views))

	page.Entries = views[start:end]
	if end < len(views) {
		page.NextOffset = end
	}

	return page, nil
}

func workflowIntegrationQueryMatches(entry WorkflowIntegrationEntry, q WorkflowIntegrationQueueQuery) bool {
	identity := workflowIntegrationQueryTarget(entry, q)
	owner := (q.Namespace == "" || entry.Namespace == q.Namespace) && (q.InvocationID == "" || entry.InvocationID == q.InvocationID)
	unit := q.Unit.ID == "" || q.Unit == entry.Unit

	return identity && owner && unit && workflowIntegrationQueryState(entry, q)
}

func workflowIntegrationEligible(entry WorkflowIntegrationEntry, journals []workflowSubmissionJournal) bool {
	if entry.State != "queued" || entry.StopIntent != "" {
		return false
	}

	target := workflowHash([]byte(entry.Repository + "\n" + entry.TargetRef))

	return !slices.ContainsFunc(journals, func(j workflowSubmissionJournal) bool {
		return j.TargetKey == target && j.Attempts[len(j.Attempts)-1].State == "awaiting_observation"
	})
}

// WorkflowSubmissionQuery resolves a retained attempt by stable submission key,
// candidate hash or owning unit. Limit bounds historical attempts as well as heads.
type WorkflowSubmissionQuery struct {
	SubmissionID  string
	CandidateHash string
	InvocationID  string
	Namespace     string
	Unit          WorkflowRef
	Limit         int
}

// QueryWorkflowIntegrationSubmissions includes uncertain and sealed history.
func (s *WorkflowStore) QueryWorkflowIntegrationSubmissions(ctx context.Context, q WorkflowSubmissionQuery) (result []WorkflowPublicationBarrier, err error) {
	if q.Limit < 1 || q.Limit > WorkflowPageLimit {
		return nil, ErrWorkflowConflict
	}

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		snapshot, readErr := s.readWorkflowIntegrationSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		result, readErr = workflowIntegrationSubmissionPage(snapshot, q)

		return readErr
	})
	if err != nil {
		result = nil
	}

	return result, integrationRecordError("publication journals", err)
}

func workflowSubmissionQueryMatches(j workflowSubmissionJournal, s WorkflowSubmission, q WorkflowSubmissionQuery) bool {
	owner := workflowSubmissionQueryOwner(j, q)
	identity := (q.SubmissionID == "" || q.SubmissionID == s.ID) && (q.Unit.ID == "" || q.Unit == j.Unit)

	hash := s.CandidateHash
	if s.Native != nil {
		hash = s.Native.CandidateHash
	}

	return owner && identity && (q.CandidateHash == "" || q.CandidateHash == hash)
}

func workflowPublicationBarrier(j workflowSubmissionJournal, submission WorkflowSubmission) WorkflowPublicationBarrier {
	return WorkflowPublicationBarrier{JournalKey: j.Key, JournalHash: j.sourceHash, TargetKey: j.TargetKey, Namespace: j.Namespace, InvocationID: j.InvocationID, Submission: submission}
}

func workflowIntegrationQueryBounds(q WorkflowIntegrationQueueQuery) bool {
	return q.Limit >= 1 && q.Limit <= WorkflowPageLimit && q.Offset >= 0 && len(q.States) <= WorkflowPageLimit
}

func workflowIntegrationQueryTarget(entry WorkflowIntegrationEntry, q WorkflowIntegrationQueueQuery) bool {
	repository := q.Repository
	if repository != "" {
		var err error

		repository, err = workflowCanonicalRepository(repository)
		if err != nil {
			return false
		}
	}

	return (repository == "" || entry.Repository == repository) && (q.TargetRef == "" || entry.TargetRef == q.TargetRef)
}

func workflowIntegrationQueryState(entry WorkflowIntegrationEntry, q WorkflowIntegrationQueueQuery) bool {
	return len(q.States) == 0 || slices.Contains(q.States, entry.State)
}

func workflowSubmissionQueryOwner(j workflowSubmissionJournal, q WorkflowSubmissionQuery) bool {
	return (q.InvocationID == "" || q.InvocationID == j.InvocationID) && (q.Namespace == "" || q.Namespace == j.Namespace)
}

func workflowIntegrationSubmissionPage(snapshot workflowIntegrationSnapshot, q WorkflowSubmissionQuery) ([]WorkflowPublicationBarrier, error) {
	result := []WorkflowPublicationBarrier{}

	for _, journal := range snapshot.Journals {
		for _, submission := range journal.Attempts {
			if workflowSubmissionQueryMatches(journal, submission, q) {
				if len(result) == q.Limit {
					return nil, fmt.Errorf("submission query exceeds limit; narrow its identity")
				}

				result = append(result, workflowPublicationBarrier(journal, submission))
			}
		}
	}

	return result, nil
}
