package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"
)

type telegramShutdownFailure struct{ quiesced <-chan struct{} }

func (*telegramShutdownFailure) Error() string {
	return "telegram shutdown deadline exceeded; owned activations remain live"
}

// One grace interval covers intake, application calls, native supervision and
// outbox delivery. A host is done only after all of its owned calls return.
func runTelegramReceiverHosts(ctx context.Context, hosts []func(context.Context) error, grace time.Duration) error {
	owned, cancel := context.WithCancel(ctx)
	defer cancel()

	finished, quiesced := startTelegramReceiverHosts(owned, hosts)
	remaining := len(hosts)

	var result error
	select {
	case result = <-finished:
		remaining--
	case <-ctx.Done():
	}

	result = telegramStorageShutdownFailure(result, owned.Err())
	cancel()

	result = telegramHostFailure(result, owned.Err())

	timer := time.NewTimer(grace)
	defer timer.Stop()

	for remaining > 0 {
		select {
		case err := <-finished:
			err = telegramStorageShutdownFailure(err, owned.Err())
			result = errors.Join(result, telegramHostFailure(err, owned.Err()))
			remaining--
		case <-timer.C:
			return errors.Join(result, &telegramShutdownFailure{quiesced: quiesced})
		}
	}

	return result
}

func telegramHostFailure(err, cancellation error) error {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var retained error
		for _, cause := range joined.Unwrap() {
			retained = errors.Join(retained, telegramHostFailure(cause, cancellation))
		}

		return retained
	}

	if cancellation != nil && errors.Is(err, cancellation) {
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			return telegramHostFailure(wrapped.Unwrap(), cancellation)
		}

		return nil
	}

	return err
}

func telegramStorageShutdownFailure(err, cancellation error) error {
	// SQLite reports SQLITE_INTERRUPT (9) for canceled calls. Only discard that
	// exact driver cause after cancellation, preserving joined checkpoint errors.
	var databaseError interface {
		error
		Code() int
	}
	if cancellation != nil && errors.As(err, &databaseError) && databaseError.Code() == 9 {
		return telegramHostFailure(err, databaseError)
	}

	return err
}

func startTelegramReceiverHosts(ctx context.Context, hosts []func(context.Context) error) (<-chan error, <-chan struct{}) {
	finished, quiesced := make(chan error, len(hosts)), make(chan struct{})

	var active sync.WaitGroup
	for _, host := range hosts {
		active.Go(func() { finished <- host(ctx) })
	}

	go func() { active.Wait(); close(quiesced) }()

	return finished, quiesced
}

type telegramReceiverResources struct{ closers []io.Closer }

// Uncooperative calls keep their database and OS exclusions until quiescence or
// process exit. A bounded error return must not permit a second live receiver.
func (r telegramReceiverResources) finish(failure error) error {
	var pending *telegramShutdownFailure
	if errors.As(failure, &pending) {
		go func() {
			<-pending.quiesced

			if r.close() != nil {
				slog.Error("Telegram receiver resource release failed", "code", "shutdown_release_failed")
			}
		}()

		return failure
	}

	return errors.Join(failure, r.close())
}

func (r telegramReceiverResources) close() (err error) {
	for i := len(r.closers) - 1; i >= 0; i-- {
		err = errors.Join(err, r.closers[i].Close())
	}

	return err
}
