package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// WorkSelectionPreflightOptions describes a package boundary value.
type WorkSelectionPreflightOptions struct {
	ProjectDir          string
	Tickets             []ticketcore.Indexed
	SelectedTicketKey   string
	BranchPolicy        executioncore.WorkBranchPolicy
	Workspace           workspacecore.ExecutionConfig
	RequireBranchPolicy bool
	Git                 WorkSelectionGit
}

// WorkSelectionPreflightResult describes a package boundary value.
type WorkSelectionPreflightResult struct {
	Allowed         bool
	SelectedTickets []ticketcore.Indexed
	BlockReasons    []WorkSelectionBlockReason
}

// WorkSelectionBlockReason describes a package boundary value.
type WorkSelectionBlockReason struct {
	Code      string
	TicketKey string
	Detail    string
}

// WorkSelectionGit defines a package boundary interface.
type WorkSelectionGit interface {
	CurrentBranch(ctx context.Context, projectDir string) (string, error)
	WorktreeClean(ctx context.Context, projectDir string) (bool, error)
	BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error)
}

// OSWorkSelectionGit describes a package boundary value.
type OSWorkSelectionGit struct{}

// RunWorkSelectionPreflight executes the corresponding package operation.
func RunWorkSelectionPreflight(ctx context.Context, opts WorkSelectionPreflightOptions) (WorkSelectionPreflightResult, error) {
	err := ctx.Err()
	if err != nil {
		return WorkSelectionPreflightResult{}, err
	}

	result := WorkSelectionPreflightResult{
		Allowed:         true,
		SelectedTickets: slices.Clone(opts.Tickets),
	}

	repairs := reviewRepairTickets(opts.Tickets)
	if len(repairs) > 0 {
		result.SelectedTickets = repairs
		if selected := strings.TrimSpace(opts.SelectedTicketKey); selected != "" && !hasTicketKey(repairs, selected) {
			result.Allowed = false
			result.BlockReasons = append(result.BlockReasons, WorkSelectionBlockReason{
				Code:      workflowcore.WorkSelectionBlockReviewRepairPending,
				TicketKey: repairs[0].Key,
				Detail:    "review repair must be selected before later development work",
			})
		}
	}

	if shouldEvaluateWorkBranchPolicy(opts) {
		blocked, err := workBranchPolicyBlockReason(ctx, opts)
		if err != nil {
			return WorkSelectionPreflightResult{}, err
		}

		if blocked != nil {
			result.Allowed = false
			result.BlockReasons = append(result.BlockReasons, *blocked)
		}
	}

	return result, nil
}

// CurrentBranch executes the corresponding package operation.
func (OSWorkSelectionGit) CurrentBranch(ctx context.Context, projectDir string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "branch", "--show-current").Output()
	if err != nil {
		return "", normalizeGitPreflightError(err)
	}

	return strings.TrimSpace(string(output)), nil
}

// WorktreeClean executes the corresponding package operation.
func (OSWorkSelectionGit) WorktreeClean(ctx context.Context, projectDir string) (bool, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return false, normalizeGitPreflightError(err)
	}

	return strings.TrimSpace(string(output)) == "", nil
}

// BranchContainsRef executes the corresponding package operation.
func (OSWorkSelectionGit) BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error) {
	err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--verify", ref).Run()
	if err != nil {
		return true, normalizeGitPreflightError(err)
	}

	err = exec.CommandContext(ctx, "git", "-C", projectDir, "merge-base", "--is-ancestor", ref, branch).Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, normalizeGitPreflightError(err)
}

func reviewRepairTickets(tickets []ticketcore.Indexed) []ticketcore.Indexed {
	repairs := make([]ticketcore.Indexed, 0)

	for _, ticket := range tickets {
		if isReviewRepairTicket(ticket) {
			repairs = append(repairs, ticket)
		}
	}

	return repairs
}

func isReviewRepairTicket(ticket ticketcore.Indexed) bool {
	return ticketcore.NeedsReviewRepair(ticket)
}

func hasTicketKey(tickets []ticketcore.Indexed, key string) bool {
	for _, ticket := range tickets {
		if ticket.Key == key {
			return true
		}
	}

	return false
}

func shouldEvaluateWorkBranchPolicy(opts WorkSelectionPreflightOptions) bool {
	return opts.RequireBranchPolicy || opts.BranchPolicy != (executioncore.WorkBranchPolicy{})
}

func workBranchPolicyBlockReason(ctx context.Context, opts WorkSelectionPreflightOptions) (*WorkSelectionBlockReason, error) {
	projectDir, err := authoringcore.CleanContextProjectDir(opts.ProjectDir)
	if err != nil {
		return nil, err
	}

	git := opts.Git
	if git == nil {
		git = OSWorkSelectionGit{}
	}

	policy := opts.BranchPolicy
	if policy == (executioncore.WorkBranchPolicy{}) {
		policy = executioncore.DefaultWorkBranchPolicy()
	}

	workspace := opts.Workspace
	if workspace == (workspacecore.ExecutionConfig{}) {
		workspace = workspacecore.DefaultExecutionConfig()
	} else {
		workspace = workspacecore.FillExecutionDefaults(workspace)
	}

	return workBranchPolicyBlockReasonForGit(ctx, workBranchPolicyCheck{
		projectDir: projectDir,
		git:        git,
		policy:     policy,
		workspace:  workspace,
	})
}

type workBranchPolicyCheck struct {
	projectDir string
	git        WorkSelectionGit
	policy     executioncore.WorkBranchPolicy
	workspace  workspacecore.ExecutionConfig
}

func workBranchPolicyBlockReasonForGit(ctx context.Context, check workBranchPolicyCheck) (*WorkSelectionBlockReason, error) {
	err := executioncore.ValidateWorkBranchPolicy(check.policy)
	if err != nil {
		return nil, err
	}

	err = workspacecore.ValidateExecutionConfig(check.workspace)
	if err != nil {
		return nil, err
	}

	if check.policy.Mode == executioncore.WorkBranchPolicyModeNoGitEnforcement {
		return nil, nil
	}

	return runWorkBranchPolicyChecks(ctx, check)
}

func runWorkBranchPolicyChecks(ctx context.Context, check workBranchPolicyCheck) (*WorkSelectionBlockReason, error) {
	if check.workspace.Mode == workspacecore.ExecutionWorktree {
		return nil, nil
	}

	branch, err := check.git.CurrentBranch(ctx, check.projectDir)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	expectedBranch := strings.TrimSpace(check.policy.IntegrationBranch)

	blocked := workBranchPolicyCurrentBranchBlockReason(check.policy, branch, expectedBranch)
	if blocked != nil {
		return blocked, nil
	}

	return workBranchPolicyRemainingBlockReason(ctx, check, expectedBranch)
}

func workBranchPolicyRemainingBlockReason(ctx context.Context, check workBranchPolicyCheck, expectedBranch string) (*WorkSelectionBlockReason, error) {
	if blocked, err := workBranchPolicyCleanBlockReasonIfNeeded(ctx, check); blocked != nil || err != nil {
		return blocked, err
	}

	return workBranchPolicyRemoteBlockReasonIfNeeded(ctx, check, expectedBranch)
}

func workBranchPolicyCleanBlockReasonIfNeeded(ctx context.Context, check workBranchPolicyCheck) (*WorkSelectionBlockReason, error) {
	if !check.policy.RequireCleanWorktree {
		return nil, nil
	}

	return workBranchPolicyCleanBlockReason(ctx, check.projectDir, check.git, check.policy)
}

func workBranchPolicyRemoteBlockReasonIfNeeded(ctx context.Context, check workBranchPolicyCheck, expectedBranch string) (*WorkSelectionBlockReason, error) {
	if check.policy.Mode != executioncore.WorkBranchPolicyModeIntegrationBranch || !check.policy.RequireRemoteTracking {
		return nil, nil
	}

	return workBranchPolicyRemoteBlockReason(ctx, check.projectDir, check.git, check.policy, expectedBranch)
}

func workBranchPolicyCurrentBranchBlockReason(policy executioncore.WorkBranchPolicy, branch, expectedBranch string) *WorkSelectionBlockReason {
	if policy.Mode != executioncore.WorkBranchPolicyModeIntegrationBranch || branch == expectedBranch {
		return nil
	}

	return &WorkSelectionBlockReason{
		Code: workflowcore.WorkSelectionBlockBranchPolicy,
		Detail: fmt.Sprintf(
			"work selection requires local %s branch, current branch is %q",
			expectedBranch,
			branch,
		),
	}
}

func workBranchPolicyCleanBlockReason(ctx context.Context, projectDir string, git WorkSelectionGit, policy executioncore.WorkBranchPolicy) (*WorkSelectionBlockReason, error) {
	clean, err := git.WorktreeClean(ctx, projectDir)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if clean {
		return nil, nil
	}

	return &WorkSelectionBlockReason{
		Code:   workflowcore.WorkSelectionBlockBranchPolicy,
		Detail: workSelectionCleanWorktreeDetail(policy),
	}, nil
}

func workBranchPolicyRemoteBlockReason(ctx context.Context, projectDir string, git WorkSelectionGit, policy executioncore.WorkBranchPolicy, expectedBranch string) (*WorkSelectionBlockReason, error) {
	remoteRef := strings.TrimSpace(policy.CanonicalRemote) + "/" + expectedBranch

	contains, err := git.BranchContainsRef(ctx, projectDir, expectedBranch, remoteRef)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if contains {
		return nil, nil
	}

	return &WorkSelectionBlockReason{
		Code: workflowcore.WorkSelectionBlockBranchPolicy,
		Detail: fmt.Sprintf(
			"work selection requires local %s to contain %s",
			expectedBranch,
			remoteRef,
		),
	}, nil
}

func workSelectionCleanWorktreeDetail(policy executioncore.WorkBranchPolicy) string {
	if policy.Mode == executioncore.WorkBranchPolicyModeIntegrationBranch {
		return fmt.Sprintf("work selection requires a clean local %s worktree", policy.IntegrationBranch)
	}

	return "work selection requires a clean local worktree"
}

func nilIfGitUnavailable(err error) (*WorkSelectionBlockReason, error) {
	if errors.Is(err, errWorkSelectionGitUnavailable) {
		return nil, nil
	}

	return nil, err
}

var errWorkSelectionGitUnavailable = errors.New("work selection git unavailable")

func normalizeGitPreflightError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "not a git repository") {
		return errWorkSelectionGitUnavailable
	}

	return err
}
