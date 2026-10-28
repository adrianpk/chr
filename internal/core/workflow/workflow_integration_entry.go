package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationBinding pins an invocation to one guarded queue image.
type IntegrationBinding struct {
	EntryID  string                    `json:"entry_id"`
	Unit     workflowstate.Ref         `json:"unit"`
	Revision uint64                    `json:"revision"`
	EntryRef workflowstate.EvidenceRef `json:"entry_ref"`
}

// IntegrationBlocker records the only permitted reconciliation resume.
type IntegrationBlocker struct {
	Reason         string `json:"reason"`
	RequiredAction string `json:"required_action"`
	ResumeState    string `json:"resume_state"`
}

// IntegrationEntry is owned by an invocation, never by a queue index.
type IntegrationEntry struct {
	DecisionRef *workflowstate.EvidenceRef `json:"decision_ref,omitempty"`
	Identity
	SchemaVersion       int                        `json:"schema_version"`
	InvocationID        string                     `json:"invocation_id"`
	Unit                workflowstate.Ref          `json:"unit"`
	CoordinatorID       string                     `json:"coordinator_id"`
	Repository          string                     `json:"repository_id"`
	TargetRef           string                     `json:"target_ref"`
	QueueSequence       uint64                     `json:"queue_sequence"`
	AdmissionKey        string                     `json:"admission_key"`
	FrozenScopeHash     string                     `json:"frozen_scope_hash"`
	State               string                     `json:"state"`
	StateRevision       uint64                     `json:"state_revision"`
	PreparationReserved bool                       `json:"preparation_reserved,omitempty"`
	ClaimGeneration     uint64                     `json:"claim_generation"`
	BudgetRef           string                     `json:"budget_ref"`
	CandidateRef        *workflowstate.EvidenceRef `json:"candidate_ref"`
	SubmissionRef       *SubmissionBinding         `json:"submission_ref"`
	ReceiptRef          *workflowstate.EvidenceRef `json:"receipt_ref"`
	Blocker             *IntegrationBlocker        `json:"blocker"`
	StopIntent          string                     `json:"stop_intent"`
	Extensions          map[string]json.RawMessage `json:"extensions,omitempty"`
}

// IntegrationAdmission requires an explicit frozen unit and full target.
type IntegrationAdmission struct {
	TokenLimit       *int64
	InvocationID     string
	ExpectedRevision uint64
	Unit             workflowstate.Ref
	AdmissionKey     string
	TargetRef        string
}

// IntegrationTransition fences callbacks by both authoritative records.
// CandidateHash is the observed candidate image hash, empty before preparation.
type IntegrationTransition struct {
	InvocationID       string
	EntryID            string
	InvocationRevision uint64
	EntryRevision      uint64
	ClaimGeneration    uint64
	CandidateHash      string
	State              string
	CandidateRef       *workflowstate.EvidenceRef
	Blocker            *IntegrationBlocker
	StopIntent         string
	ClearStopIntent    bool
}

func workflowIntegrationEntryID(run workflowInvocationFile, unit workflowstate.Ref) string {
	return contenthash.Sum([]byte(run.ProjectID + "\x00" + run.Namespace + "\x00" + run.ID + "\x00" + unit.Kind + "\x00" + unit.ID))
}

func workflowIntegrationEntryPath(invocationID, entryID string) string {
	return filepath.ToSlash(filepath.Join("run", invocationID, "extension", "chr", "integration", entryID, "queue.json"))
}

func workflowIntegrationBudgetPath(id string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "integration", "budget.json"))
}

func workflowIntegrationScopeHash(run workflowInvocationFile, unit workflowstate.Ref) (string, error) {
	members := []PreviewMember{}

	for _, member := range run.Members {
		if workflowMemberOwnsUnit(member, unit) {
			members = append(members, member.Admission)
		}
	}

	data, err := json.Marshal(struct {
		RequestHash string
		Unit        workflowstate.Ref
		Members     []PreviewMember
	}{run.RequestHash, unit, members})

	return contenthash.Sum(data), err
}

func newWorkflowIntegrationEntry(run workflowInvocationFile, request IntegrationAdmission, coordinator workflowCoordinator, sequence uint64) (IntegrationEntry, error) {
	hash, err := workflowIntegrationScopeHash(run, request.Unit)
	if err != nil {
		return IntegrationEntry{}, err
	}

	repository, err := workflowCanonicalRepository(run.ExecutionSelection().Target.Repository)
	if err != nil {
		return IntegrationEntry{}, err
	}

	now := time.Now().UTC()
	entry := IntegrationEntry{
		Identity:      Identity{ID: workflowIntegrationEntryID(run, request.Unit), ProjectID: run.ProjectID, Namespace: run.Namespace, Revision: 1, CreatedAt: now, UpdatedAt: now},
		SchemaVersion: 1, InvocationID: run.ID, Unit: request.Unit, CoordinatorID: coordinator.ID,
		Repository: repository, TargetRef: request.TargetRef,
		QueueSequence: sequence, AdmissionKey: request.AdmissionKey, FrozenScopeHash: hash,
		State: "queued", StateRevision: 1, BudgetRef: workflowIntegrationBudgetPath(run.ID),
	}

	return entry, validateWorkflowIntegrationEntry(entry)
}

// AdmitWorkflowIntegration attaches retries before charging any new queue slot.
func (s *Store) AdmitWorkflowIntegration(ctx context.Context, request IntegrationAdmission) (entry IntegrationEntry, err error) {
	err = s.withLock(ctx, func() error {
		runs, readErr := s.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		for _, run := range runs {
			owner, ownerErr := s.workflowOwnerStore(run)
			if ownerErr != nil {
				return ownerErr
			}

			if recordErr := owner.checkWorkflowIntegrationRetainedRecords(ctx, run); recordErr != nil {
				return recordErr
			}
		}

		entry, readErr = s.admitWorkflowIntegration(ctx, runs, request)

		return readErr
	})

	return entry, err
}

// TransitionWorkflowIntegration journals the entry and its exact invocation pointer.
func (s *Store) TransitionWorkflowIntegration(ctx context.Context, request IntegrationTransition) (entry IntegrationEntry, err error) {
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
	return fmt.Errorf("%w: integration entry %s", workflowstate.ErrConflict, reason)
}
