package runtime

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"
)

// Reservation extends an admitted slot to an independently owned
// runtime. It names the existing capacity files, never a second worker pool.
type Reservation struct {
	Slot  string
	Guard string
}

// HoldReservation holds only shared references to an admitted
// reservation. A competing exclusive admission blocks it; cancellation prevents
// abandoned startup from running after a later claim acquired the slot.
func HoldReservation(ctx context.Context, reservation Reservation) (release func() error, err error) {
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
		file, err := openReservationFile(path)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
		if err := holdReservationFile(ctx, file); err != nil {
			return nil, err
		}
	}

	return closeFiles, nil
}

func holdReservationFile(ctx context.Context, file *os.File) error {
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

func openReservationFile(path string) (*os.File, error) {
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
