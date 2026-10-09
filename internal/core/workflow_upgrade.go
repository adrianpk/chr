package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
)

// WorkflowUpgradeOptions explicitly selects a project policy and, optionally,
// one approved legacy plan. Omitting ReviewMode preserves its effective policy.
type WorkflowUpgradeOptions struct {
	ExpectedRevision string                       `json:"expected_revision"`
	StateAuthority   WorkStateAuthority           `json:"state_authority"`
	ReviewMode       ReviewWorkflowMode           `json:"review_mode,omitempty"`
	Adoption         *WorkflowPlanAdoptionOptions `json:"adoption,omitempty"`
	CheckQuiescent   func(context.Context) error  `json:"-"`
}

// WorkflowUpgradeResult retains the policy observed before the version marker.
// Snapshot is relative to the project, outside the workflow record namespace.
type WorkflowUpgradeResult struct {
	Snapshot string                         `json:"snapshot"`
	Prior    WorkflowConfiguration          `json:"prior"`
	Current  WorkflowConfiguration          `json:"current"`
	Adopted  *WorkflowMaterializationResult `json:"adopted,omitempty"`
}

// UpgradeWorkflowConfiguration upgrades only the selected, quiescent project.
// Repeating the same observed revision and explicit choices resumes its journal.
func (s *WorkflowStore) UpgradeWorkflowConfiguration(ctx context.Context, opts WorkflowUpgradeOptions) (result WorkflowUpgradeResult, err error) {
	if err := validateWorkflowUpgradeOptions(opts); err != nil {
		return result, err
	}

	capacity, err := acquireWorkflowCapacity(ctx, s.layout.ProjectDir())
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, capacity.Close()) }()

	locks, err := s.acquireWorkflowUpgradeLocks(ctx)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, releaseWorkflowAdoptionLocks(locks)) }()

	err = s.withLock(ctx, func() (err error) {
		if err := s.checkWorkflowUpgradeQuiescent(ctx, opts); err != nil {
			return err
		}

		result, err = s.upgradeWorkflowConfiguration(ctx, opts)

		return err
	})

	return result, err
}

func validateWorkflowUpgradeOptions(opts WorkflowUpgradeOptions) error {
	if len(opts.ExpectedRevision) != 64 || !workflowConfigChoice(string(opts.StateAuthority), []string{"tickets", "tracker_compat"}) {
		return fmt.Errorf("upgrade requires an observed configuration revision and explicit state authority")
	}

	if _, err := hex.DecodeString(opts.ExpectedRevision); err != nil {
		return fmt.Errorf("invalid configuration revision: %w", err)
	}

	if opts.ReviewMode != "" && !IntegrationCycleMode(opts.ReviewMode).Valid() {
		return fmt.Errorf("unsupported upgrade review mode")
	}

	if opts.CheckQuiescent == nil {
		return fmt.Errorf("legacy project ownership observation is required")
	}

	if opts.Adoption != nil {
		if opts.StateAuthority != WorkStateAuthorityTickets || opts.Adoption.CheckQuiescent == nil {
			return fmt.Errorf("selected adoption requires ticket authority and legacy task ownership observation")
		}

		return ValidateGeneratedWorkflowPlan(opts.Adoption.Plan)
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowUpgradeQuiescent(ctx context.Context, opts WorkflowUpgradeOptions) error {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if !workflowInvocationTerminal(run.Status) {
			return fmt.Errorf("unfinished workflow %s blocks project upgrade", run.ID)
		}

		for _, attempt := range run.Attempts {
			if err := workflowAttemptQuiescent(attempt); err != nil {
				return err
			}
		}
	}

	return opts.CheckQuiescent(ctx)
}

func (s *WorkflowStore) acquireWorkflowUpgradeLocks(ctx context.Context) ([]*AgentOperationLock, error) {
	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return nil, err
	}

	locks := []*AgentOperationLock{}

	for _, layout := range layouts {
		lock, err := acquireWorkflowAdoptionLock(layout)
		if err != nil {
			return nil, errors.Join(err, releaseWorkflowAdoptionLocks(locks))
		}

		locks = append(locks, lock)
	}

	return locks, nil
}
