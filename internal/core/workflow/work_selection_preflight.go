package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
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
	Operation string           `json:"operation"`
	Owner     string           `json:"owner"`
	PID       int              `json:"pid"`
	CreatedAt string           `json:"created_at"`
	AttemptID string           `json:"attempt_id,omitempty"`
	Process   *ProcessIdentity `json:"process,omitempty"`
}

// AcquireAgentOperationLock executes the corresponding package operation.
func AcquireAgentOperationLock(ctx context.Context, opts AgentOperationLockOptions) (*AgentOperationLock, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	settings, err := normalizeAgentOperationLockOptions(opts)
	if err != nil {
		return nil, err
	}

	lockPath, err := agentOperationLockPath(settings.projectDir)
	if err != nil {
		return nil, err
	}

	reservation, err := acquireAgentRepositoryCapacity(ctx, settings.projectDir)
	if err != nil {
		return nil, err
	}

	lock, err := acquireAgentOperationFile(lockPath, settings.operation, opts.Owner, settings.createdAt)
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

	data, err := opsfile.Read(lockPath)
	if err != nil {
		return nil, err
	}

	return &AgentOperationLock{path: lockPath, guard: contenthash.Sum(data)}, nil
}

type agentOperationSettings struct {
	projectDir, operation string
	createdAt             time.Time
}

func normalizeAgentOperationLockOptions(opts AgentOperationLockOptions) (agentOperationSettings, error) {
	projectDir, err := authoringcore.CleanContextProjectDir(opts.ProjectDir)
	if err != nil {
		return agentOperationSettings{}, err
	}

	operation := strings.TrimSpace(opts.Operation)
	if operation == "" {
		return agentOperationSettings{}, fmt.Errorf("agent operation is required")
	}

	createdAt := opts.Now.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return agentOperationSettings{projectDir: projectDir, operation: operation, createdAt: createdAt}, nil
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

func agentOperationLockPath(projectDir string) (string, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return "", err
	}

	return filepath.Join(layout.Root(), "lock", "agent-operation.lock"), nil
}
