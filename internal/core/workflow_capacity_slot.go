package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

func (s *WorkflowStore) acquireWorkflowWorkerSlot(ctx context.Context) (slot, guard *os.File, err error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return nil, nil, err
	}

	limits, err := s.workflowRepositoryCapacity(ctx, runs)
	if err != nil {
		return nil, nil, err
	}

	common, err := workflowRepositoryDirectory(ctx, s.layout.ProjectDir())
	if err != nil {
		return nil, nil, err
	}

	guard, err = workflowCapacityFile(common, "chr-workflow-worker.lock", syscall.LOCK_SH)
	if err != nil {
		return nil, nil, err
	}

	owned := guard

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.Close())
			if slot != nil {
				err = errors.Join(err, slot.Close())
				slot = nil
			}

			guard = nil
		}
	}()

	slot, err = reserveWorkflowWorkerSlot(ctx, common, limits.Workers)

	return slot, guard, err
}

func workflowCapacityFile(common, name string, mode int) (*os.File, error) {
	file, err := os.OpenFile(filepath.Join(common, name), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), mode|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) {
			err = ErrWorkflowCapacity
		}

		return nil, errors.Join(err, file.Close())
	}

	return file, nil
}

func reserveWorkflowWorkerSlot(ctx context.Context, common string, workers int) (slot *os.File, err error) {
	lock, err := os.OpenFile(filepath.Join(common, "chr-workflow-capacity.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	if err := opsfile.Lock(ctx, lock); err != nil {
		return nil, err
	}

	occupied, err := workflowOccupiedProcessSlots(common)
	if err != nil {
		return nil, err
	}

	free, err := workflowFreeWorkerSlots(common, occupied)
	defer func() { err = errors.Join(err, closeWorkflowSlotFiles(free, slot)) }()

	if err != nil {
		return nil, err
	}

	if len(occupied) >= workers {
		return nil, ErrWorkflowCapacity
	}

	for index := 1; index <= workers; index++ {
		if file := free[index]; file != nil {
			return file, nil
		}
	}

	return nil, ErrWorkflowCapacity
}

// Probe all eight slots, including slots above a reduced limit. Live work is
// neither killed nor omitted from the count when capacity is reduced.
func workflowFreeWorkerSlots(common string, occupied map[int]bool) (map[int]*os.File, error) {
	free := map[int]*os.File{}

	for slot := 1; slot <= workflowWorkerSlotLimit; slot++ {
		if occupied[slot] {
			continue
		}

		file, err := workflowCapacityFile(common, workflowWorkerSlotName(slot), syscall.LOCK_EX)
		if errors.Is(err, ErrWorkflowCapacity) {
			occupied[slot] = true

			continue
		}

		if err != nil {
			return free, err
		}

		free[slot] = file
	}

	return free, nil
}

func closeWorkflowSlotFiles(files map[int]*os.File, keep *os.File) error {
	var result error

	for _, file := range files {
		if file != keep {
			result = errors.Join(result, file.Close())
		}
	}

	return result
}

func workflowWorkerSlotName(slot int) string {
	return fmt.Sprintf("chr-workflow-slot-%d.lock", slot)
}

func workflowWorkerSlotNumber(path string) int {
	name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "chr-workflow-slot-"), ".lock")

	number, err := strconv.Atoi(name)
	if err != nil {
		return 0
	}

	return number
}
