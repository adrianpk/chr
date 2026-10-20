package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// WorkerRuntimeReservation extends an admitted slot to an independently owned
// runtime. It names the existing capacity files, never a second worker pool.
type WorkerRuntimeReservation struct {
	Slot  string
	Guard string
}

// DelegateWorkerRuntimeReservation converts the inherited slot to shared mode
// under the existing capacity coordinator. New admissions still require an
// exclusive slot. The runtime can therefore retain that same slot after the
// host process dies, without a free interval during lock conversion.
func DelegateWorkerRuntimeReservation(ctx context.Context, directory string) (WorkerRuntimeReservation, error) {
	common, err := workflowRepositoryDirectory(ctx, directory)
	if err != nil {
		return WorkerRuntimeReservation{}, err
	}

	reservation, err := inheritedWorkerRuntimeReservation(common)
	if err != nil {
		return WorkerRuntimeReservation{}, err
	}

	lock, err := os.OpenFile(filepath.Join(common, "chr-workflow-capacity.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return WorkerRuntimeReservation{}, err
	}
	defer func() { _ = lock.Close() }() //nolint:errcheck // Closing releases only this coordinator descriptor.

	if err := acquireWorkflowLock(ctx, lock); err != nil {
		return WorkerRuntimeReservation{}, err
	}

	if err := syscall.Flock(3, syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		return WorkerRuntimeReservation{}, err
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

// HoldWorkerRuntimeReservation holds only shared references to an admitted
// reservation. A competing exclusive admission blocks it; cancellation prevents
// abandoned startup from running after a later claim acquired the slot.
func HoldWorkerRuntimeReservation(ctx context.Context, reservation WorkerRuntimeReservation) (release func() error, err error) {
	files := []*os.File{}
	closeFiles := func() error {
		var err error
		for _, file := range files {
			err = errors.Join(err, file.Close())
		}

		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, closeFiles())
		}
	}()

	for _, path := range []string{reservation.Slot, reservation.Guard} {
		file, err := openRuntimeReservationFile(path)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
		if err := holdRuntimeReservationFile(ctx, file); err != nil {
			return nil, err
		}
	}

	return closeFiles, nil
}

func holdRuntimeReservationFile(ctx context.Context, file *os.File) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
		if err == nil {
			return nil
		}

		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func inheritedWorkerRuntimeReservation(common string) (WorkerRuntimeReservation, error) {
	slot, err := inheritedRuntimeFile(3)
	if err != nil || filepath.Dir(slot) != common || workflowWorkerSlotNumber(slot) < 1 || workflowWorkerSlotNumber(slot) > workflowWorkerSlotLimit || filepath.Base(slot) != workflowWorkerSlotName(workflowWorkerSlotNumber(slot)) {
		return WorkerRuntimeReservation{}, errors.New("worker runtime requires its inherited admitted capacity slot")
	}

	guard := filepath.Join(common, "chr-workflow-worker.lock")

	actual, err := inheritedRuntimeFile(5)
	if err != nil || actual != guard {
		return WorkerRuntimeReservation{}, errors.New("worker runtime requires its inherited capacity guard")
	}

	return WorkerRuntimeReservation{Slot: slot, Guard: guard}, nil
}

func openRuntimeReservationFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return nil, errors.Join(errors.New("runtime capacity reservation must be a private regular file"), file.Close())
	}

	return file, nil
}
