package core

import (
	"fmt"
	"strings"
)

// WorktreePlanOptions describes a package boundary value.
type WorktreePlanOptions struct {
	ProjectDir   string
	TicketKey    string
	Branch       string
	BranchPolicy WorkBranchPolicy
	Workspace    WorkWorkspaceConfig
	BaseSHA      string
	WorktreePath string
}

// WorktreePlan describes a package boundary value.
type WorktreePlan struct {
	ProjectDir   string
	TicketKey    string
	Branch       string
	BaseRef      string
	BaseSHA      string
	WorktreeRoot string
	WorktreePath string
}

// NewWorktreePlan executes the corresponding package operation.
func NewWorktreePlan(opts WorktreePlanOptions) (WorktreePlan, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return WorktreePlan{}, err
	}

	ticketKey := strings.TrimSpace(opts.TicketKey)
	if ticketKey == "" {
		return WorktreePlan{}, fmt.Errorf("ticket key is required")
	}

	branch := strings.TrimSpace(opts.Branch)
	if branch == "" {
		return WorktreePlan{}, fmt.Errorf("ticket branch is required")
	}

	branchPolicy, workspace, err := normalizedWorktreePlanConfig(opts)
	if err != nil {
		return WorktreePlan{}, err
	}

	baseRef, err := worktreeBaseRef(branchPolicy, workspace)
	if err != nil {
		return WorktreePlan{}, err
	}

	return WorktreePlan{
		ProjectDir:   projectDir,
		TicketKey:    ticketKey,
		Branch:       branch,
		BaseRef:      baseRef,
		BaseSHA:      strings.TrimSpace(opts.BaseSHA),
		WorktreeRoot: resolveProjectPath(projectDir, workspace.WorktreeRoot),
		WorktreePath: strings.TrimSpace(opts.WorktreePath),
	}, nil
}

func normalizedWorktreePlanConfig(opts WorktreePlanOptions) (WorkBranchPolicy, WorkWorkspaceConfig, error) {
	branchPolicy := opts.BranchPolicy
	if branchPolicy == (WorkBranchPolicy{}) {
		branchPolicy = DefaultWorkBranchPolicy()
	}

	if err := ValidateWorkBranchPolicy(branchPolicy); err != nil {
		return WorkBranchPolicy{}, WorkWorkspaceConfig{}, err
	}

	workspace := opts.Workspace
	if workspace == (WorkWorkspaceConfig{}) {
		workspace = DefaultWorkWorkspaceConfig()
	} else {
		workspace = fillWorkWorkspaceConfigDefaults(workspace)
	}

	if err := ValidateWorkWorkspaceConfig(workspace); err != nil {
		return WorkBranchPolicy{}, WorkWorkspaceConfig{}, err
	}

	return branchPolicy, workspace, nil
}

func worktreeBaseRef(policy WorkBranchPolicy, workspace WorkWorkspaceConfig) (string, error) {
	switch workspace.WorktreeBase {
	case WorktreeBasePolicyRemoteIntegrationBranch:
		remote := strings.TrimSpace(policy.CanonicalRemote)
		branch := strings.TrimSpace(policy.IntegrationBranch)
		if remote == "" || branch == "" {
			return "", fmt.Errorf("remote integration branch base requires canonical remote and integration branch")
		}

		return remote + "/" + branch, nil
	case WorktreeBasePolicyLocalIntegrationBranch:
		branch := strings.TrimSpace(policy.IntegrationBranch)
		if branch == "" {
			return "", fmt.Errorf("local integration branch base requires integration branch")
		}

		return branch, nil
	case WorktreeBasePolicyCurrentBranch:
		return "HEAD", nil
	default:
		return "", fmt.Errorf("invalid work worktree_base: %s", workspace.WorktreeBase)
	}
}
