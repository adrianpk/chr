package workflow

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// UpgradeOptions explicitly selects a project policy and, optionally,
// one approved legacy plan. Omitting ReviewMode preserves its effective policy.
type UpgradeOptions struct {
	ExpectedRevision string                             `json:"expected_revision"`
	StateAuthority   workspaceconfig.WorkStateAuthority `json:"state_authority"`
	ReviewMode       workflowstate.ReviewMode           `json:"review_mode,omitempty"`
	Adoption         *PlanAdoptionOptions               `json:"adoption,omitempty"`
	CheckQuiescent   func(context.Context) error        `json:"-"`
}

// UpgradeResult retains the policy observed before the version marker.
// Snapshot is relative to the project, outside the workflow record namespace.
type UpgradeResult struct {
	Snapshot string                                `json:"snapshot"`
	Prior    workspaceconfig.WorkflowConfiguration `json:"prior"`
	Current  workspaceconfig.WorkflowConfiguration `json:"current"`
	Adopted  *MaterializationResult                `json:"adopted,omitempty"`
}

// UpgradeWorkflowConfiguration upgrades only the selected, quiescent project.
// Repeating the same observed revision and explicit choices resumes its journal.
func (s *Store) UpgradeWorkflowConfiguration(ctx context.Context, opts UpgradeOptions) (result UpgradeResult, err error) {
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
		return workspacecore.WithConfigLock(ctx, s.layout.ProjectDir(), func() error {
			if err := workspaceconfig.RequireNoHarnessSelectionChange(s.layout.ProjectDir()); err != nil {
				return err
			}

			if err := s.checkWorkflowUpgradeQuiescent(ctx, opts); err != nil {
				return err
			}

			result, err = s.upgradeWorkflowConfiguration(ctx, opts)

			return err
		})
	})

	return result, err
}

func validateWorkflowUpgradeOptions(opts UpgradeOptions) error {
	if len(opts.ExpectedRevision) != 64 || !slices.Contains([]string{"tickets", "tracker_compat"}, string(opts.StateAuthority)) {
		return fmt.Errorf("upgrade requires an observed configuration revision and explicit state authority")
	}

	if _, err := hex.DecodeString(opts.ExpectedRevision); err != nil {
		return fmt.Errorf("invalid configuration revision: %w", err)
	}

	if opts.ReviewMode != "" && !workflowstate.IntegrationCycleMode(opts.ReviewMode).Valid() {
		return fmt.Errorf("unsupported upgrade review mode")
	}

	if opts.CheckQuiescent == nil {
		return fmt.Errorf("legacy project ownership observation is required")
	}

	if opts.Adoption != nil {
		if opts.StateAuthority != workspaceconfig.WorkStateAuthorityTickets || opts.Adoption.CheckQuiescent == nil {
			return fmt.Errorf("selected adoption requires ticket authority and legacy task ownership observation")
		}

		return generationcore.ValidateGeneratedWorkflowPlan(opts.Adoption.Plan)
	}

	return nil
}

func (s *Store) checkWorkflowUpgradeQuiescent(ctx context.Context, opts UpgradeOptions) error {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if !workflowstate.InvocationTerminal(run.Status) {
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

func (s *Store) acquireWorkflowUpgradeLocks(ctx context.Context) ([]*AgentOperationLock, error) {
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
