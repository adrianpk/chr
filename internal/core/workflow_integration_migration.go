package core

import (
	"context"
	"math"
	"time"
)

// MigrateWorkflowIntegration explicitly admits a missing unit, or adds the
// sealed ledger to an earlier queue entry. It never upgrades a candidate,
// clears a claim, seals uncertain publication or rewrites historical receipts.
func (s *WorkflowStore) MigrateWorkflowIntegration(ctx context.Context, request WorkflowIntegrationAdmission) (entry WorkflowIntegrationEntry, err error) {
	err = s.withLock(ctx, func() error {
		runs, readErr := s.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		for _, run := range runs {
			if run.ID == request.InvocationID && run.Namespace == s.layout.Namespace() {
				owner, readErr := s.workflowOwnerStore(run)
				if readErr != nil {
					return readErr
				}

				entry, readErr = owner.migrateWorkflowIntegration(ctx, run, runs, request)

				return readErr
			}
		}

		return ErrWorkflowConflict
	})

	return entry, integrationRecordError(workflowInvocationPath(request.InvocationID), err)
}

func (s *WorkflowStore) migrateWorkflowIntegration(ctx context.Context, run workflowInvocationFile, runs []workflowInvocationFile, request WorkflowIntegrationAdmission) (WorkflowIntegrationEntry, error) {
	if run.Revision != request.ExpectedRevision || run.Revision == math.MaxUint64 {
		return WorkflowIntegrationEntry{}, ErrWorkflowConflict
	}

	if len(run.IntegrationEntries) == 0 {
		if err := s.checkWorkflowIntegrationFiles(run); err != nil {
			return WorkflowIntegrationEntry{}, err
		}

		return s.admitWorkflowIntegration(ctx, runs, request)
	}

	return s.migrateWorkflowIntegrationLedger(ctx, run, request)
}

func (s *WorkflowStore) migrateWorkflowIntegrationLedger(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationAdmission) (WorkflowIntegrationEntry, error) {
	if run.Integration == nil || run.Integration.Unit != request.Unit {
		return WorkflowIntegrationEntry{}, ErrWorkflowConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return entry, err
	}

	if entry.TargetRef != request.TargetRef || entry.AdmissionKey != request.AdmissionKey {
		return entry, ErrWorkflowConflict
	}

	if run.IntegrationBudget != nil {
		_, err := s.readWorkflowIntegrationBudget(run)

		return entry, err
	}

	if workflowIntegrationMigrationHasEvidence(entry) {
		return entry, workflowIntegrationConflict("retained candidate requires compatible preparation before ledger migration")
	}

	budget, err := s.workflowIntegrationBudgetAdmission(ctx, run, entry, request.TokenLimit)
	if err != nil {
		return entry, err
	}

	return entry, s.commitWorkflowIntegrationMigration(ctx, &run, &budget)
}

func (s *WorkflowStore) commitWorkflowIntegrationMigration(ctx context.Context, run *workflowInvocationFile, budget *WorkflowIntegrationBudget) error {
	file, err := s.workflowIntegrationBudgetChange(ctx, run, budget)
	if err != nil {
		return err
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	invocation, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles([]workflowFileChange{file, invocation})
}

func workflowIntegrationMigrationHasEvidence(entry WorkflowIntegrationEntry) bool {
	return entry.CandidateRef != nil || entry.ReceiptRef != nil || entry.SubmissionRef != nil
}
