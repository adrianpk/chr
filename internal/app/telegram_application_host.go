package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramApplicationCompletion struct {
	updateID int64
	lane     core.TelegramProcessingLane
	err      error
}

// The inbox is durable; this host remembers only calls owned by this process.
type telegramApplicationHost struct {
	receiver telegramReceiver
	active   map[int64]bool
	finished chan telegramApplicationCompletion
}

func newTelegramApplicationHost(receiver telegramReceiver) *telegramApplicationHost {
	return &telegramApplicationHost{receiver: receiver, active: map[int64]bool{}, finished: make(chan telegramApplicationCompletion, telegramWorkActivationLimit)}
}

func (h *telegramApplicationHost) run(ctx context.Context) (err error) {
	owned, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		for len(h.active) > 0 {
			err = errors.Join(err, h.complete(<-h.finished))
		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for ctx.Err() == nil {
		if err := h.pump(owned); err != nil {
			return err
		}

		select {
		case result := <-h.finished:
			if err := h.complete(result); err != nil {
				return err
			}
		case <-ticker.C:
		case <-ctx.Done():
		}
	}

	return nil
}

func (h *telegramApplicationHost) pump(ctx context.Context) error {
	for _, lane := range []core.TelegramProcessingLane{core.TelegramProcessingUrgent, core.TelegramProcessingDecision, core.TelegramProcessingOrdinary} {
		updates, err := h.receiver.store.PendingApplicationInputs(ctx, h.receiver.botID, lane, time.Now())
		if err != nil {
			return err
		}

		for _, update := range updates {
			if h.active[update.UpdateID] {
				continue
			}

			if !h.receiver.slots.acquire(lane) {
				break
			}

			h.active[update.UpdateID] = true
			receiver := h.receiver
			receiver.processingLane = &lane

			go func() {
				h.finished <- telegramApplicationCompletion{update.UpdateID, lane, receiver.respond(ctx, update)}
			}()
		}
	}

	return nil
}

func (h *telegramApplicationHost) complete(result telegramApplicationCompletion) error {
	delete(h.active, result.updateID)
	h.receiver.slots.release(result.lane)

	var pending *telegramControlReceiptPending
	if errors.As(result.err, &pending) && telegramControlExpectedFailure(pending.failure) {
		return nil
	}

	if telegramApplicationDeferred(result.err) {
		return nil
	}

	var apiErr *core.TelegramAPIError
	if errors.As(result.err, &apiErr) && apiErr.Code == "local_receiver_conflict" {
		return nil
	}

	return result.err
}

func telegramApplicationDeferred(err error) bool {
	for _, expected := range []error{core.ErrStructuredModelCapacity, core.ErrTelegramReplyPending, errTelegramLaneChanged, context.Canceled} {
		err = telegramHostFailure(err, expected)
	}

	return err == nil
}

func (r telegramReceiver) runDeliveries(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for ctx.Err() == nil {
		if err := r.drainDeliveries(ctx); err != nil {
			return err
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
		}
	}

	return nil
}
