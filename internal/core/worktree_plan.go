package core

import (
	"fmt"
	"path/filepath"
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

	worktreeRoot, err := resolvedWorktreeRoot(projectDir)
	if err != nil {
		return WorktreePlan{}, err
	}

	worktreePath, err := worktreeTargetPath(worktreeRoot, ticketKey, opts.WorktreePath)
	if err != nil {
		return WorktreePlan{}, err
	}

	return WorktreePlan{
		ProjectDir:   projectDir,
		TicketKey:    ticketKey,
		Branch:       branch,
		BaseRef:      baseRef,
		BaseSHA:      strings.TrimSpace(opts.BaseSHA),
		WorktreeRoot: worktreeRoot,
		WorktreePath: worktreePath,
	}, nil
}

func resolvedWorktreeRoot(projectDir string) (string, error) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return "", err
	}
	return layout.WorktreeRoot(), nil
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

func worktreeTargetPath(worktreeRoot, ticketKey, configuredPath string) (string, error) {
	path := strings.TrimSpace(configuredPath)
	if path != "" {
		return filepath.Clean(path), nil
	}

	component := WorktreePathComponent(ticketKey)
	if component == "" {
		return "", fmt.Errorf("ticket key %q cannot be used as a worktree path component", ticketKey)
	}

	return filepath.Join(worktreeRoot, component), nil
}

// WorktreePathComponent executes the corresponding package operation.
func WorktreePathComponent(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	output := strings.Builder{}
	lastDash := false

	for _, char := range value {
		if worktreePathRuneAllowed(char) {
			output.WriteRune(char)

			lastDash = false

			continue
		}

		if !lastDash {
			output.WriteRune('-')

			lastDash = true
		}
	}

	return strings.Trim(output.String(), "-_.")
}

func worktreePathRuneAllowed(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '-' ||
		char == '_' ||
		char == '.'
}
