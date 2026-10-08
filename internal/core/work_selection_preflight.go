package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	// AgentOperationDevelopment defines a package boundary constant.
	AgentOperationDevelopment = "development"
	// AgentOperationReview defines a package boundary constant.
	AgentOperationReview = "review"
	// AgentOperationMerge defines a package boundary constant.
	AgentOperationMerge = "merge"

	// WorkSelectionBlockReviewRepairPending defines a package boundary constant.
	WorkSelectionBlockReviewRepairPending = "review_repair_pending"
	// WorkSelectionBlockBranchPolicy defines a package boundary constant.
	WorkSelectionBlockBranchPolicy = "branch_policy_blocked"
)

// AgentOperationLockOptions describes a package boundary value.
type AgentOperationLockOptions struct {
	ProjectDir string
	Operation  string
	Owner      string
	Now        time.Time
}

// AgentOperationLock describes a package boundary value.
type AgentOperationLock struct {
	path        string
	guard       string
	reservation *os.File
}

type agentOperationLockRecord struct {
	Operation string                   `json:"operation"`
	Owner     string                   `json:"owner"`
	PID       int                      `json:"pid"`
	CreatedAt string                   `json:"created_at"`
	AttemptID string                   `json:"attempt_id,omitempty"`
	Process   *WorkflowProcessIdentity `json:"process,omitempty"`
}

// WorkSelectionPreflightOptions describes a package boundary value.
type WorkSelectionPreflightOptions struct {
	ProjectDir          string
	Tickets             []IndexedTicket
	SelectedTicketKey   string
	BranchPolicy        WorkBranchPolicy
	Workspace           WorkWorkspaceConfig
	RequireBranchPolicy bool
	Git                 WorkSelectionGit
}

// WorkSelectionPreflightResult describes a package boundary value.
type WorkSelectionPreflightResult struct {
	Allowed         bool
	SelectedTickets []IndexedTicket
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

// AcquireAgentOperationLock executes the corresponding package operation.
func AcquireAgentOperationLock(ctx context.Context, opts AgentOperationLockOptions) (*AgentOperationLock, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	projectDir, operation, createdAt, err := normalizeAgentOperationLockOptions(opts)
	if err != nil {
		return nil, err
	}

	lockPath, err := agentOperationLockPath(projectDir)
	if err != nil {
		return nil, err
	}

	reservation, err := acquireAgentRepositoryCapacity(ctx, projectDir)
	if err != nil {
		return nil, err
	}

	lock, err := acquireAgentOperationFile(lockPath, operation, opts.Owner, createdAt)
	if err != nil {
		if reservation != nil {
			err = errors.Join(err, reservation.Close())
		}

		return nil, err
	}

	lock.reservation = reservation

	return lock, nil
}

func acquireAgentOperationFile(lockPath, operation, owner string, createdAt time.Time) (*AgentOperationLock, error) {
	file, err := createAgentOperationLockFile(lockPath)
	if err != nil {
		return nil, err
	}

	record := agentOperationLockRecord{
		Operation: operation,
		Owner:     strings.TrimSpace(owner),
		PID:       os.Getpid(),
		CreatedAt: createdAt.Format(time.RFC3339),
	}

	err = writeAgentOperationLockFile(file, lockPath, record)
	if err != nil {
		return nil, err
	}

	data, err := readWorkflowFile(lockPath)
	if err != nil {
		return nil, err
	}

	return &AgentOperationLock{path: lockPath, guard: workflowHash(data)}, nil
}

func normalizeAgentOperationLockOptions(opts AgentOperationLockOptions) (string, string, time.Time, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return "", "", time.Time{}, err
	}

	operation := strings.TrimSpace(opts.Operation)
	if operation == "" {
		return "", "", time.Time{}, fmt.Errorf("agent operation is required")
	}

	createdAt := opts.Now.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return projectDir, operation, createdAt, nil
}

func createAgentOperationLockFile(lockPath string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(lockPath), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create agent operation lock directory: %w", err)
	}

	file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("agent operation lock is already held: %s", lockPath)
		}

		return nil, fmt.Errorf("create agent operation lock: %w", err)
	}

	return file, nil
}

func writeAgentOperationLockFile(file *os.File, lockPath string, record agentOperationLockRecord) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(record)
	closeErr := file.Close()

	if err != nil {
		removeErr := os.Remove(lockPath)
		if removeErr != nil {
			return fmt.Errorf("remove failed agent operation lock: %w", removeErr)
		}

		return fmt.Errorf("write agent operation lock: %w", err)
	}

	if closeErr != nil {
		removeErr := os.Remove(lockPath)
		if removeErr != nil {
			return fmt.Errorf("remove failed agent operation lock: %w", removeErr)
		}

		return fmt.Errorf("close agent operation lock: %w", closeErr)
	}

	return nil
}

// Release executes the corresponding package operation.
func (lock *AgentOperationLock) Release() error {
	if lock == nil || lock.path == "" {
		return nil
	}

	if err := guardedAgentOperationRelease(lock); err != nil {
		return err
	}

	if lock.reservation != nil {
		err := lock.reservation.Close()
		lock.reservation = nil

		return err
	}

	return nil
}

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
				Code:      WorkSelectionBlockReviewRepairPending,
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

func agentOperationLockPath(projectDir string) (string, error) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return "", err
	}

	return filepath.Join(layout.Root(), "lock", "agent-operation.lock"), nil
}

func reviewRepairTickets(tickets []IndexedTicket) []IndexedTicket {
	repairs := make([]IndexedTicket, 0)

	for _, ticket := range tickets {
		if isReviewRepairTicket(ticket) {
			repairs = append(repairs, ticket)
		}
	}

	return repairs
}

func isReviewRepairTicket(ticket IndexedTicket) bool {
	return IsTicketReviewRepair(ticket)
}

func hasTicketKey(tickets []IndexedTicket, key string) bool {
	for _, ticket := range tickets {
		if ticket.Key == key {
			return true
		}
	}

	return false
}

func shouldEvaluateWorkBranchPolicy(opts WorkSelectionPreflightOptions) bool {
	return opts.RequireBranchPolicy || opts.BranchPolicy != (WorkBranchPolicy{})
}

func workBranchPolicyBlockReason(ctx context.Context, opts WorkSelectionPreflightOptions) (*WorkSelectionBlockReason, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return nil, err
	}

	git := opts.Git
	if git == nil {
		git = OSWorkSelectionGit{}
	}

	policy := opts.BranchPolicy
	if policy == (WorkBranchPolicy{}) {
		policy = DefaultWorkBranchPolicy()
	}

	workspace := opts.Workspace
	if workspace == (WorkWorkspaceConfig{}) {
		workspace = DefaultWorkWorkspaceConfig()
	} else {
		workspace = fillWorkWorkspaceConfigDefaults(workspace)
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
	policy     WorkBranchPolicy
	workspace  WorkWorkspaceConfig
}

func workBranchPolicyBlockReasonForGit(ctx context.Context, check workBranchPolicyCheck) (*WorkSelectionBlockReason, error) {
	err := ValidateWorkBranchPolicy(check.policy)
	if err != nil {
		return nil, err
	}

	err = ValidateWorkWorkspaceConfig(check.workspace)
	if err != nil {
		return nil, err
	}

	if check.policy.Mode == WorkBranchPolicyModeNoGitEnforcement {
		return nil, nil
	}

	return runWorkBranchPolicyChecks(ctx, check)
}

func runWorkBranchPolicyChecks(ctx context.Context, check workBranchPolicyCheck) (*WorkSelectionBlockReason, error) {
	if check.workspace.Mode == WorkWorkspaceModeWorktree {
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
	if check.policy.Mode != WorkBranchPolicyModeIntegrationBranch || !check.policy.RequireRemoteTracking {
		return nil, nil
	}

	return workBranchPolicyRemoteBlockReason(ctx, check.projectDir, check.git, check.policy, expectedBranch)
}

func workBranchPolicyCurrentBranchBlockReason(policy WorkBranchPolicy, branch, expectedBranch string) *WorkSelectionBlockReason {
	if policy.Mode != WorkBranchPolicyModeIntegrationBranch || branch == expectedBranch {
		return nil
	}

	return &WorkSelectionBlockReason{
		Code: WorkSelectionBlockBranchPolicy,
		Detail: fmt.Sprintf(
			"work selection requires local %s branch, current branch is %q",
			expectedBranch,
			branch,
		),
	}
}

func workBranchPolicyCleanBlockReason(ctx context.Context, projectDir string, git WorkSelectionGit, policy WorkBranchPolicy) (*WorkSelectionBlockReason, error) {
	clean, err := git.WorktreeClean(ctx, projectDir)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if clean {
		return nil, nil
	}

	return &WorkSelectionBlockReason{
		Code:   WorkSelectionBlockBranchPolicy,
		Detail: workSelectionCleanWorktreeDetail(policy),
	}, nil
}

func workBranchPolicyRemoteBlockReason(ctx context.Context, projectDir string, git WorkSelectionGit, policy WorkBranchPolicy, expectedBranch string) (*WorkSelectionBlockReason, error) {
	remoteRef := strings.TrimSpace(policy.CanonicalRemote) + "/" + expectedBranch

	contains, err := git.BranchContainsRef(ctx, projectDir, expectedBranch, remoteRef)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if contains {
		return nil, nil
	}

	return &WorkSelectionBlockReason{
		Code: WorkSelectionBlockBranchPolicy,
		Detail: fmt.Sprintf(
			"work selection requires local %s to contain %s",
			expectedBranch,
			remoteRef,
		),
	}, nil
}

func workSelectionCleanWorktreeDetail(policy WorkBranchPolicy) string {
	if policy.Mode == WorkBranchPolicyModeIntegrationBranch {
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
