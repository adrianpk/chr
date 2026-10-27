package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
)

// TestTelegramRecoveryApplicationError keeps a real checkpoint failure visible
// when an application completion also contains its canceled call context.
func TestTelegramRecoveryApplicationError(t *testing.T) {
	checkpoint := errors.New("application checkpoint failed")
	for _, test := range []struct {
		name    string
		failure error
		want    error
	}{
		{"canceled", context.Canceled, nil},
		{"wrapped cancellation", fmt.Errorf("canceled call: %w", context.Canceled), nil},
		{"checkpoint", checkpoint, checkpoint},
		{"joined", errors.Join(context.Canceled, checkpoint), checkpoint},
		{"wrapped join", fmt.Errorf("application call: %w", errors.Join(context.Canceled, checkpoint)), checkpoint},
		{"deferred join", errors.Join(context.Canceled, modelcore.ErrCapacity), nil},
		{"checkpoint and wait", errors.Join(context.Canceled, modelcore.ErrCapacity, checkpoint), checkpoint},
	} {
		t.Run(test.name, func(t *testing.T) {
			host := newTelegramApplicationHost(telegramReceiver{})

			err := host.complete(telegramApplicationCompletion{updateID: 1, err: test.failure})
			if !errors.Is(err, test.want) {
				t.Fatal("application completion hid checkpoint failure", err)
			}
		})
	}
}
