package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationRecordError identifies retained data requiring reconciliation.
// Readers never replace the failing source image with a runnable default.
type IntegrationRecordError struct {
	Path  string
	Cause error
}

func (e *IntegrationRecordError) Error() string {
	return fmt.Sprintf("integration records require reconciliation at %s: %v", e.Path, e.Cause)
}

func (e *IntegrationRecordError) Unwrap() error {
	return errors.Join(workflowstate.ErrConflict, e.Cause)
}

// IntegrationView resolves an entry's sealed ledger and immutable evidence.
type IntegrationView struct {
	Entry        IntegrationEntry
	Budget       IntegrationBudget
	CurrentUsage IntegrationUsage
}

// PublicationBarrier is derived from a common journal, never its index.
type PublicationBarrier struct {
	JournalKey   string
	JournalHash  string
	TargetKey    string
	Namespace    string
	InvocationID string
	Submission   workflowstate.Submission
}

// IntegrationQueueQuery selects a bounded page of authoritative units.
// EligibleOnly means queued, without stop intent or uncertain target publication.
type IntegrationQueueQuery struct {
	Repository   string
	TargetRef    string
	Namespace    string
	InvocationID string
	Unit         workflowstate.Ref
	States       []string
	EligibleOnly bool
	Offset       int
	Limit        int
	SnapshotHash string
}

// IntegrationQueuePage is observational; it confers no execution claim.
type IntegrationQueuePage struct {
	Entries      []IntegrationView
	SnapshotHash string
	NextOffset   int
}

type workflowIntegrationSnapshot struct {
	Views    []IntegrationView
	Journals []workflowSubmissionJournal
	Runs     []workflowInvocationFile
	Hash     string
}

// QueryWorkflowIntegrationQueue does not replay a namespace or publication journal.
func (s *Store) QueryWorkflowIntegrationQueue(ctx context.Context, query IntegrationQueueQuery) (page IntegrationQueuePage, err error) {
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

	var retained *IntegrationRecordError
	if errors.As(err, &retained) {
		return err
	}

	return &IntegrationRecordError{Path: path, Cause: err}
}

func validateWorkflowIntegrationQueueQuery(q IntegrationQueueQuery) error {
	if !workflowIntegrationQueryBounds(q) {
		return workflowstate.ErrConflict
	}

	if q.Namespace != "" && !opsnamespace.Valid(q.Namespace) {
		return workflowstate.ErrConflict
	}

	for _, state := range q.States {
		if !validWorkflowIntegrationState(IntegrationState(state)) {
			return workflowstate.ErrConflict
		}
	}

	if q.Repository != "" {
		if _, err := workflowstate.CanonicalRepository(q.Repository); err != nil {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func workflowIntegrationQueuePage(snapshot workflowIntegrationSnapshot, q IntegrationQueueQuery) (IntegrationQueuePage, error) {
	page := IntegrationQueuePage{SnapshotHash: snapshot.Hash, NextOffset: -1}
	if q.SnapshotHash != "" && q.SnapshotHash != snapshot.Hash {
		return page, workflowstate.ErrConflict
	}

	views := []IntegrationView{}

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

func workflowIntegrationQueryMatches(entry IntegrationEntry, q IntegrationQueueQuery) bool {
	identity := workflowIntegrationQueryTarget(entry, q)
	owner := (q.Namespace == "" || entry.Namespace == q.Namespace) && (q.InvocationID == "" || entry.InvocationID == q.InvocationID)
	unit := q.Unit.ID == "" || q.Unit == entry.Unit

	return identity && owner && unit && workflowIntegrationQueryState(entry, q)
}

func workflowIntegrationEligible(entry IntegrationEntry, journals []workflowSubmissionJournal) bool {
	if entry.State != "queued" || entry.StopIntent != "" {
		return false
	}

	target := contenthash.Sum([]byte(entry.Repository + "\n" + entry.TargetRef))

	return !slices.ContainsFunc(journals, func(j workflowSubmissionJournal) bool {
		return j.TargetKey == target && j.Attempts[len(j.Attempts)-1].State == "awaiting_observation"
	})
}

// SubmissionQuery resolves a retained attempt by stable submission key,
// candidate hash or owning unit. Limit bounds historical attempts as well as heads.
type SubmissionQuery struct {
	SubmissionID  string
	CandidateHash string
	InvocationID  string
	Namespace     string
	Unit          workflowstate.Ref
	Limit         int
}

// QueryWorkflowIntegrationSubmissions includes uncertain and sealed history.
func (s *Store) QueryWorkflowIntegrationSubmissions(ctx context.Context, q SubmissionQuery) (result []PublicationBarrier, err error) {
	if q.Limit < 1 || q.Limit > workflowstate.PageLimit {
		return nil, workflowstate.ErrConflict
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

func workflowSubmissionQueryMatches(j workflowSubmissionJournal, s workflowstate.Submission, q SubmissionQuery) bool {
	owner := workflowSubmissionQueryOwner(j, q)
	identity := (q.SubmissionID == "" || q.SubmissionID == s.ID) && (q.Unit.ID == "" || q.Unit == j.Unit)

	hash := s.CandidateHash
	if s.Native != nil {
		hash = s.Native.CandidateHash
	}

	return owner && identity && (q.CandidateHash == "" || q.CandidateHash == hash)
}

func workflowPublicationBarrier(j workflowSubmissionJournal, submission workflowstate.Submission) PublicationBarrier {
	return PublicationBarrier{JournalKey: j.Key, JournalHash: j.sourceHash, TargetKey: j.TargetKey, Namespace: j.Namespace, InvocationID: j.InvocationID, Submission: submission}
}

func workflowIntegrationQueryBounds(q IntegrationQueueQuery) bool {
	return q.Limit >= 1 && q.Limit <= workflowstate.PageLimit && q.Offset >= 0 && len(q.States) <= workflowstate.PageLimit
}

func workflowIntegrationQueryTarget(entry IntegrationEntry, q IntegrationQueueQuery) bool {
	repository := q.Repository
	if repository != "" {
		var err error

		repository, err = workflowstate.CanonicalRepository(repository)
		if err != nil {
			return false
		}
	}

	return (repository == "" || entry.Repository == repository) && (q.TargetRef == "" || entry.TargetRef == q.TargetRef)
}

func workflowIntegrationQueryState(entry IntegrationEntry, q IntegrationQueueQuery) bool {
	return len(q.States) == 0 || slices.Contains(q.States, string(entry.State))
}

func workflowSubmissionQueryOwner(j workflowSubmissionJournal, q SubmissionQuery) bool {
	return (q.InvocationID == "" || q.InvocationID == j.InvocationID) && (q.Namespace == "" || q.Namespace == j.Namespace)
}

func workflowIntegrationSubmissionPage(snapshot workflowIntegrationSnapshot, q SubmissionQuery) ([]PublicationBarrier, error) {
	result := []PublicationBarrier{}

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
