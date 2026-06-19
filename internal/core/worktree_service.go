package core

import (
	"context"
	"fmt"
	"os"
)

// WorktreeService describes a package boundary value.
type WorktreeService struct {
	Git WorktreeGit
}

// WorktreePrepareResult describes a package boundary value.
type WorktreePrepareResult struct {
	Plan          WorktreePlan
	CreatedBranch bool
	AddedWorktree bool
}

// NewWorktreeService constructs the corresponding package service or value.
func NewWorktreeService() WorktreeService {
	return WorktreeService{}
}

// Prepare executes the corresponding package operation.
func (s WorktreeService) Prepare(ctx context.Context, opts WorktreePlanOptions) (WorktreePrepareResult, error) {
	settings, err := newWorktreeServiceSettings(opts, s.Git)
	if err != nil {
		return WorktreePrepareResult{}, err
	}

	if err := ctx.Err(); err != nil {
		return WorktreePrepareResult{}, err
	}

	return prepareWorktree(ctx, settings)
}

type worktreeServiceSettings struct {
	plan      WorktreePlan
	policy    WorkBranchPolicy
	workspace WorkWorkspaceConfig
	git       WorktreeGit
}

func newWorktreeServiceSettings(opts WorktreePlanOptions, git WorktreeGit) (worktreeServiceSettings, error) {
	policy, workspace, err := normalizedWorktreePlanConfig(opts)
	if err != nil {
		return worktreeServiceSettings{}, err
	}

	plan, err := NewWorktreePlan(opts)
	if err != nil {
		return worktreeServiceSettings{}, err
	}

	if git == nil {
		git = OSWorktreeGit{}
	}

	return worktreeServiceSettings{
		plan:      plan,
		policy:    policy,
		workspace: workspace,
		git:       git,
	}, nil
}

func prepareWorktree(ctx context.Context, settings worktreeServiceSettings) (WorktreePrepareResult, error) {
	result := WorktreePrepareResult{Plan: settings.plan}

	if err := settings.prepareBase(ctx, &result); err != nil {
		return WorktreePrepareResult{}, err
	}

	createdBranch, err := settings.ensureBranch(ctx)
	if err != nil {
		return WorktreePrepareResult{}, err
	}

	result.CreatedBranch = createdBranch

	if err := settings.ensureRoot(); err != nil {
		return WorktreePrepareResult{}, err
	}

	return settings.addWorktree(ctx, result)
}

func (settings worktreeServiceSettings) prepareBase(ctx context.Context, result *WorktreePrepareResult) error {
	if settings.workspace.WorktreeBase == WorktreeBasePolicyRemoteIntegrationBranch {
		if err := settings.git.FetchRemote(ctx, settings.plan.ProjectDir, settings.policy.CanonicalRemote); err != nil {
			return fmt.Errorf("fetch worktree base: %w", err)
		}
	}

	baseSHA, err := settings.git.ResolveRef(ctx, settings.plan.ProjectDir, settings.plan.BaseRef)
	if err != nil {
		return fmt.Errorf("resolve worktree base: %w", err)
	}

	result.Plan.BaseSHA = baseSHA

	return nil
}

func (settings worktreeServiceSettings) ensureBranch(ctx context.Context) (bool, error) {
	branchExists, err := settings.git.BranchExists(ctx, settings.plan.ProjectDir, settings.plan.Branch)
	if err != nil {
		return false, fmt.Errorf("inspect worktree branch: %w", err)
	}

	if branchExists {
		return false, nil
	}

	if err := settings.git.CreateBranch(ctx, settings.plan.ProjectDir, settings.plan.Branch, settings.plan.BaseRef); err != nil {
		return false, fmt.Errorf("create worktree branch: %w", err)
	}

	return true, nil
}

func (settings worktreeServiceSettings) ensureRoot() error {
	if err := os.MkdirAll(settings.plan.WorktreeRoot, 0o755); err != nil {
		return fmt.Errorf("create worktree root: %w", err)
	}

	return nil
}

func (settings worktreeServiceSettings) addWorktree(ctx context.Context, result WorktreePrepareResult) (WorktreePrepareResult, error) {
	if err := settings.git.AddWorktree(ctx, settings.plan.ProjectDir, settings.plan.WorktreePath, settings.plan.Branch); err != nil {
		return WorktreePrepareResult{}, fmt.Errorf("add worktree: %w", err)
	}

	result.AddedWorktree = true

	return result, nil
}
