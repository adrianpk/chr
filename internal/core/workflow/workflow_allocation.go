package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowAllocationSchema = "chr-materialization-lock/v1"

type workflowAllocationOwner struct {
	Schema string `json:"schema"`
	PID    int    `json:"pid"`
	Nonce  int64  `json:"nonce"`
}

func (s *Store) acquireWorkflowAllocation() (string, error) {
	path := filepath.Join(s.layout.Root(), "lock", "id.lock")
	if err := s.reclaimWorkflowAllocation(path); err != nil {
		return "", err
	}

	if err := opscore.AcquireArtifactLock(path); err != nil {
		return "", err
	}

	data, err := json.Marshal(workflowAllocationOwner{Schema: workflowAllocationSchema, PID: os.Getpid(), Nonce: time.Now().UnixNano()})
	if err != nil {
		return "", err
	}

	if err := opsfile.WriteAtomic(path, data); err != nil {
		return "", err
	}

	return contenthash.Sum(data), nil
}

func (s *Store) reclaimWorkflowAllocation(path string) error {
	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	var owner workflowAllocationOwner
	if err := json.Unmarshal(data, &owner); err != nil || owner.Schema != workflowAllocationSchema || owner.PID < 1 {
		return fmt.Errorf("existing artifact allocation lock has no materialization ownership proof")
	}
	// An existing live PID, including a reused PID, conservatively blocks reclaim.
	if err := syscall.Kill(owner.PID, 0); !errors.Is(err, syscall.ESRCH) {
		return fmt.Errorf("materialization allocator is still owned")
	}

	return s.releaseWorkflowAllocation(contenthash.Sum(data))
}

func (s *Store) releaseWorkflowAllocation(guard string) error {
	if guard == "" {
		return nil
	}

	path := filepath.Join(s.layout.Root(), "lock", "id.lock")

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if contenthash.Sum(data) != guard {
		return fmt.Errorf("%w: allocation lock owner changed", workflowstate.ErrConflict)
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return opsfile.SyncDirectory(filepath.Dir(path))
}
