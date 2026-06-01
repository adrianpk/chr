package chc

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
	AgentOperationDevelopment = "development"
	AgentOperationReview      = "review"

	WorkSelectionBlockReviewRepairPending = "review_repair_pending"
	WorkSelectionBlockStaleDev            = "stale_dev"
)

type AgentOperationLockOptions struct {
	ProjectDir string
	Operation  string
	Owner      string
	Now        time.Time
}

type AgentOperationLock struct {
	path string
}

type agentOperationLockRecord struct {
	Operation string `json:"operation"`
	Owner     string `json:"owner"`
	PID       int    `json:"pid"`
	CreatedAt string `json:"created_at"`
}

type WorkSelectionPreflightOptions struct {
	ProjectDir        string
	Tickets           []IndexedTicket
	SelectedTicketKey string
	RequireUpdatedDev bool
	Git               WorkSelectionGit
}

type WorkSelectionPreflightResult struct {
	Allowed         bool
	SelectedTickets []IndexedTicket
	BlockReasons    []WorkSelectionBlockReason
}

type WorkSelectionBlockReason struct {
	Code      string
	TicketKey string
	Detail    string
}

type WorkSelectionGit interface {
	CurrentBranch(ctx context.Context, projectDir string) (string, error)
	WorktreeClean(ctx context.Context, projectDir string) (bool, error)
	DevContainsOriginDev(ctx context.Context, projectDir string) (bool, error)
}

type OSWorkSelectionGit struct{}

func AcquireAgentOperationLock(ctx context.Context, opts AgentOperationLockOptions) (*AgentOperationLock, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	projectDir, operation, createdAt, err := normalizeAgentOperationLockOptions(opts)
	if err != nil {
		return nil, err
	}

	lockPath := agentOperationLockPath(projectDir)

	file, err := createAgentOperationLockFile(lockPath)
	if err != nil {
		return nil, err
	}

	record := agentOperationLockRecord{
		Operation: operation,
		Owner:     strings.TrimSpace(opts.Owner),
		PID:       os.Getpid(),
		CreatedAt: createdAt.Format(time.RFC3339),
	}

	err = writeAgentOperationLockFile(file, lockPath, record)
	if err != nil {
		return nil, err
	}

	return &AgentOperationLock{path: lockPath}, nil
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

func (lock *AgentOperationLock) Release() error {
	if lock == nil || lock.path == "" {
		return nil
	}

	err := os.Remove(lock.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("release agent operation lock: %w", err)
	}

	lock.path = ""

	return nil
}

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

	if opts.RequireUpdatedDev {
		blocked, err := staleDevBlockReason(ctx, opts)
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

func (OSWorkSelectionGit) CurrentBranch(ctx context.Context, projectDir string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "branch", "--show-current").Output()
	if err != nil {
		return "", normalizeGitPreflightError(err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (OSWorkSelectionGit) WorktreeClean(ctx context.Context, projectDir string) (bool, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return false, normalizeGitPreflightError(err)
	}

	return strings.TrimSpace(string(output)) == "", nil
}

func (OSWorkSelectionGit) DevContainsOriginDev(ctx context.Context, projectDir string) (bool, error) {
	err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--verify", "origin/dev").Run()
	if err != nil {
		return true, normalizeGitPreflightError(err)
	}

	err = exec.CommandContext(ctx, "git", "-C", projectDir, "merge-base", "--is-ancestor", "origin/dev", "dev").Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, normalizeGitPreflightError(err)
}

func agentOperationLockPath(projectDir string) string {
	return filepath.Join(projectDir, DirectoryName, UserDirectoryName, "runtime", "agent-operation.lock")
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
	return ticket.Status == TicketStatusReadyForDev &&
		(ticket.ReviewState == "changes_requested" || ticket.ReviewResult == "needs_changes")
}

func hasTicketKey(tickets []IndexedTicket, key string) bool {
	for _, ticket := range tickets {
		if ticket.Key == key {
			return true
		}
	}

	return false
}

func staleDevBlockReason(ctx context.Context, opts WorkSelectionPreflightOptions) (*WorkSelectionBlockReason, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return nil, err
	}

	git := opts.Git
	if git == nil {
		git = OSWorkSelectionGit{}
	}

	return staleDevBlockReasonForGit(ctx, projectDir, git)
}

func staleDevBlockReasonForGit(ctx context.Context, projectDir string, git WorkSelectionGit) (*WorkSelectionBlockReason, error) {
	branch, err := git.CurrentBranch(ctx, projectDir)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if branch != "dev" {
		return &WorkSelectionBlockReason{
			Code:   WorkSelectionBlockStaleDev,
			Detail: fmt.Sprintf("work selection requires local dev branch, current branch is %q", branch),
		}, nil
	}

	clean, err := git.WorktreeClean(ctx, projectDir)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if !clean {
		return &WorkSelectionBlockReason{
			Code:   WorkSelectionBlockStaleDev,
			Detail: "work selection requires a clean local dev worktree",
		}, nil
	}

	contains, err := git.DevContainsOriginDev(ctx, projectDir)
	if err != nil {
		return nilIfGitUnavailable(err)
	}

	if !contains {
		return &WorkSelectionBlockReason{
			Code:   WorkSelectionBlockStaleDev,
			Detail: "work selection requires local dev to contain origin/dev",
		}, nil
	}

	return nil, nil
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
