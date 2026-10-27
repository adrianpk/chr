package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

// DelegateWorkerRuntimeReservation converts the inherited slot to shared mode
// under the existing capacity coordinator. New admissions still require an
// exclusive slot. The runtime can therefore retain that same slot after the
// host process dies, without a free interval during lock conversion.
func DelegateWorkerRuntimeReservation(ctx context.Context, directory string) (harnessruntime.WorkerRuntimeReservation, error) {
	common, err := projectcore.ResolveRepositoryDirectory(ctx, directory)
	if err != nil {
		return harnessruntime.WorkerRuntimeReservation{}, err
	}

	reservation, err := inheritedWorkerRuntimeReservation(common)
	if err != nil {
		return harnessruntime.WorkerRuntimeReservation{}, err
	}

	lock, err := os.OpenFile(filepath.Join(common, "chr-workflow-capacity.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return harnessruntime.WorkerRuntimeReservation{}, err
	}
	defer func() { _ = lock.Close() }() //nolint:errcheck // Closing releases only this coordinator descriptor.

	if err := opsfile.Lock(ctx, lock); err != nil {
		return harnessruntime.WorkerRuntimeReservation{}, err
	}

	if err := syscall.Flock(3, syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		return harnessruntime.WorkerRuntimeReservation{}, err
	}

	return reservation, nil
}

func inheritedRuntimeFile(fd int) (string, error) {
	name := filepath.Join("/proc/self/fd", strconv.Itoa(fd))

	path, err := os.Readlink(name)
	if err != nil {
		return "", err
	}

	var inherited, current syscall.Stat_t
	if err := syscall.Fstat(fd, &inherited); err != nil {
		return "", err
	}

	if err := syscall.Lstat(path, &current); err != nil || inherited.Ino != current.Ino || inherited.Dev != current.Dev || current.Mode&syscall.S_IFMT != syscall.S_IFREG {
		return "", errors.New("inherited runtime reservation changed")
	}

	return path, nil
}

func inheritedWorkerRuntimeReservation(common string) (harnessruntime.WorkerRuntimeReservation, error) {
	slot, err := inheritedRuntimeFile(3)
	if err != nil || filepath.Dir(slot) != common || workflowWorkerSlotNumber(slot) < 1 || workflowWorkerSlotNumber(slot) > workflowWorkerSlotLimit || filepath.Base(slot) != workflowWorkerSlotName(workflowWorkerSlotNumber(slot)) {
		return harnessruntime.WorkerRuntimeReservation{}, errors.New("worker runtime requires its inherited admitted capacity slot")
	}

	guard := filepath.Join(common, "chr-workflow-worker.lock")

	actual, err := inheritedRuntimeFile(5)
	if err != nil || actual != guard {
		return harnessruntime.WorkerRuntimeReservation{}, errors.New("worker runtime requires its inherited capacity guard")
	}

	return harnessruntime.WorkerRuntimeReservation{Slot: slot, Guard: guard}, nil
}
