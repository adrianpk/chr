package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// WorkflowIntegrationBinding pins an invocation to one guarded queue image.
type WorkflowIntegrationBinding struct {
	EntryID  string              `json:"entry_id"`
	Unit     WorkflowRef         `json:"unit"`
	Revision uint64              `json:"revision"`
	EntryRef WorkflowEvidenceRef `json:"entry_ref"`
}

// WorkflowIntegrationBlocker records the only permitted reconciliation resume.
type WorkflowIntegrationBlocker struct {
	Reason         string `json:"reason"`
	RequiredAction string `json:"required_action"`
	ResumeState    string `json:"resume_state"`
}

// WorkflowIntegrationEntry is owned by an invocation, never by a queue index.
type WorkflowIntegrationEntry struct {
	WorkflowIdentity
	SchemaVersion   int                         `json:"schema_version"`
	InvocationID    string                      `json:"invocation_id"`
	Unit            WorkflowRef                 `json:"unit"`
	CoordinatorID   string                      `json:"coordinator_id"`
	Repository      string                      `json:"repository_id"`
	TargetRef       string                      `json:"target_ref"`
	QueueSequence   uint64                      `json:"queue_sequence"`
	AdmissionKey    string                      `json:"admission_key"`
	FrozenScopeHash string                      `json:"frozen_scope_hash"`
	State           string                      `json:"state"`
	StateRevision   uint64                      `json:"state_revision"`
	ClaimGeneration uint64                      `json:"claim_generation"`
	BudgetRef       string                      `json:"budget_ref"`
	CandidateRef    *WorkflowEvidenceRef        `json:"candidate_ref"`
	SubmissionRef   *WorkflowSubmissionBinding  `json:"submission_ref"`
	ReceiptRef      *WorkflowEvidenceRef        `json:"receipt_ref"`
	Blocker         *WorkflowIntegrationBlocker `json:"blocker"`
	StopIntent      string                      `json:"stop_intent"`
	Extensions      map[string]json.RawMessage  `json:"extensions,omitempty"`
}

// WorkflowIntegrationAdmission requires an explicit frozen unit and full target.
type WorkflowIntegrationAdmission struct {
	TokenLimit       *int64
	InvocationID     string
	ExpectedRevision uint64
	Unit             WorkflowRef
	AdmissionKey     string
	TargetRef        string
}

// WorkflowIntegrationTransition fences callbacks by both authoritative records.
// CandidateHash is the observed candidate image hash, empty before preparation.
type WorkflowIntegrationTransition struct {
	InvocationID       string
	EntryID            string
	InvocationRevision uint64
	EntryRevision      uint64
	ClaimGeneration    uint64
	CandidateHash      string
	State              string
	CandidateRef       *WorkflowEvidenceRef
	Blocker            *WorkflowIntegrationBlocker
	StopIntent         string
	ClearStopIntent    bool
}

func workflowIntegrationEntryID(run workflowInvocationFile, unit WorkflowRef) string {
	return workflowHash([]byte(run.ProjectID + "\x00" + run.Namespace + "\x00" + run.ID + "\x00" + unit.Kind + "\x00" + unit.ID))
}

func workflowIntegrationEntryPath(invocationID, entryID string) string {
	return filepath.ToSlash(filepath.Join("run", invocationID, "extension", "chr", "integration", entryID, "queue.json"))
}

func workflowIntegrationBudgetPath(id string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "integration", "budget.json"))
}

func workflowIntegrationScopeHash(run workflowInvocationFile, unit WorkflowRef) (string, error) {
	members := []WorkflowPreviewMember{}

	for _, member := range run.Members {
		if workflowMemberOwnsUnit(member, unit) {
			members = append(members, member.Admission)
		}
	}

	data, err := json.Marshal(struct {
		RequestHash string
		Unit        WorkflowRef
		Members     []WorkflowPreviewMember
	}{run.RequestHash, unit, members})

	return workflowHash(data), err
}

func newWorkflowIntegrationEntry(run workflowInvocationFile, request WorkflowIntegrationAdmission, coordinator workflowCoordinator, sequence uint64) (WorkflowIntegrationEntry, error) {
	hash, err := workflowIntegrationScopeHash(run, request.Unit)
	if err != nil {
		return WorkflowIntegrationEntry{}, err
	}

	repository, err := workflowCanonicalRepository(run.ExecutionSelection().Target.Repository)
	if err != nil {
		return WorkflowIntegrationEntry{}, err
	}

	now := time.Now().UTC()
	entry := WorkflowIntegrationEntry{
		WorkflowIdentity: WorkflowIdentity{ID: workflowIntegrationEntryID(run, request.Unit), ProjectID: run.ProjectID, Namespace: run.Namespace, Revision: 1, CreatedAt: now, UpdatedAt: now},
		SchemaVersion:    1, InvocationID: run.ID, Unit: request.Unit, CoordinatorID: coordinator.ID,
		Repository: repository, TargetRef: request.TargetRef,
		QueueSequence: sequence, AdmissionKey: request.AdmissionKey, FrozenScopeHash: hash,
		State: "queued", StateRevision: 1, BudgetRef: workflowIntegrationBudgetPath(run.ID),
	}

	return entry, validateWorkflowIntegrationEntry(entry)
}

// AdmitWorkflowIntegration attaches retries before charging any new queue slot.
func (s *WorkflowStore) AdmitWorkflowIntegration(ctx context.Context, request WorkflowIntegrationAdmission) (entry WorkflowIntegrationEntry, err error) {
	err = s.withLock(ctx, func() error {
		runs, readErr := s.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		entry, readErr = s.admitWorkflowIntegration(ctx, runs, request)

		return readErr
	})

	return entry, err
}

// TransitionWorkflowIntegration journals the entry and its exact invocation pointer.
func (s *WorkflowStore) TransitionWorkflowIntegration(ctx context.Context, request WorkflowIntegrationTransition) (entry WorkflowIntegrationEntry, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		entry, readErr = owner.transitionWorkflowIntegration(ctx, run, request)

		return readErr
	})

	return entry, err
}

func workflowIntegrationTerminal(state string) bool {
	return state == "integrated" || state == "cancelled"
}

func workflowIntegrationConflict(reason string) error {
	return fmt.Errorf("%w: integration entry %s", ErrWorkflowConflict, reason)
}
