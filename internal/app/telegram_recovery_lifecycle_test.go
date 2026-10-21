package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramRecoveryGrace bounds a caller that ignores cancellation and keeps
// real configuration exclusion until that caller actually returns.
func TestTelegramRecoveryGrace(t *testing.T) {
	receiver, _ := telegramAppTestReceiver(t, core.TelegramReceiverAPI{})

	lock, err := core.AcquireTelegramConfigurationLock(receiver.env)
	if err != nil {
		t.Fatal(err)
	}

	entered, release := make(chan struct{}), make(chan struct{})

	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() { <-entered; cancel() }()

	host := func(context.Context) error { close(entered); <-release; return nil }
	failure := runTelegramReceiverHosts(ctx, []func(context.Context) error{host}, 10*time.Millisecond)

	var pending *telegramShutdownFailure
	if !errors.As(failure, &pending) {
		t.Fatal("unquiesced shutdown was reported successful", failure)
	}

	resources := telegramReceiverResources{closers: []io.Closer{lock}}
	if !errors.Is(resources.finish(failure), failure) {
		t.Fatal("shutdown failure was lost")
	}

	if other, err := core.AcquireTelegramConfigurationLock(receiver.env); err == nil {
		if err := other.Close(); err != nil {
			t.Error(err)
		}

		t.Fatal("receiver exclusion released before quiescence")
	}

	unblock()

	select {
	case <-pending.quiesced:
	case <-time.After(time.Second):
		t.Fatal("owned host did not finish")
	}

	waitTelegramRecoveryLock(t, receiver.env)
}

func waitTelegramRecoveryLock(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		lock, err := core.AcquireTelegramConfigurationLock(env)
		if err == nil {
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}

			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("quiesced receiver retained its resources")
}

// TestTelegramRecoveryHostFailure preserves a real failure joined with normal
// cancellation, including a failure returned by a different draining host.
func TestTelegramRecoveryHostFailure(t *testing.T) {
	failure := errors.New("fixture checkpoint failure")

	for _, fixture := range []struct {
		name           string
		first, wrapped bool
	}{
		{"first", true, false}, {"draining", false, false}, {"wrapped_first", true, true}, {"wrapped_draining", false, true},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			cause := errors.Join(context.Canceled, failure)
			if fixture.wrapped {
				cause = fmt.Errorf("fixture host failure: %w", cause)
			}

			failed := func(context.Context) error { return cause }
			draining := func(ctx context.Context) error { <-ctx.Done(); return cause }

			hosts := []func(context.Context) error{func(context.Context) error { return nil }, draining}
			if fixture.first {
				hosts = []func(context.Context) error{failed}
			}

			err := runTelegramReceiverHosts(t.Context(), hosts, time.Second)
			if !errors.Is(err, failure) || errors.Is(err, context.Canceled) {
				t.Fatal("host failure was hidden by cancellation", err)
			}
		})
	}
}

type telegramTestDatabaseError int

func (e telegramTestDatabaseError) Error() string { return fmt.Sprintf("database error (%d)", e) }

func (e telegramTestDatabaseError) Code() int { return int(e) }

// TestTelegramRecoveryStorageCancel covers the driver's interrupt result during
// shutdown without hiding an unrelated database or checkpoint failure.
func TestTelegramRecoveryStorageCancel(t *testing.T) {
	interrupted, busy := telegramTestDatabaseError(9), telegramTestDatabaseError(5)
	checkpoint := errors.New("checkpoint failed")
	plain := errors.New("interrupted (9)")

	for _, test := range []struct {
		name    string
		failure error
		cancel  bool
		want    error
	}{
		{"interrupted", interrupted, true, nil},
		{"wrapped", fmt.Errorf("query: %w", interrupted), true, nil},
		{"not canceled", interrupted, false, interrupted},
		{"other database failure", busy, true, busy},
		{"joined checkpoint", errors.Join(interrupted, checkpoint), true, checkpoint},
		{"wrapped checkpoint", fmt.Errorf("host: %w", errors.Join(interrupted, checkpoint)), true, checkpoint},
		{"plain message", plain, true, plain},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			host := func(owned context.Context) error {
				if test.cancel {
					cancel()
					<-owned.Done()
				}

				return test.failure
			}
			err := runTelegramReceiverHosts(ctx, []func(context.Context) error{host}, time.Second)
			if !errors.Is(err, test.want) {
				t.Fatal("shutdown lost the original failure semantics", err)
			}
		})
	}
}
