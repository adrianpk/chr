package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

const telegramWorkActivationLimit = 16

// This registry holds only in-flight calls. Work, eligibility, recovery and
// capacity remain owned by the shared native workflow operation.
type telegramWorkHost struct {
	receiver telegramReceiver
	active   map[string]bool
	finished chan telegramWorkCompletion
	after    string
}

func newTelegramWorkHost(receiver telegramReceiver) *telegramWorkHost {
	return &telegramWorkHost{receiver: receiver, active: map[string]bool{}, finished: make(chan telegramWorkCompletion, telegramWorkActivationLimit)}
}

func (h *telegramWorkHost) run(ctx context.Context) (err error) {
	owned, cancel := context.WithCancel(ctx)
	defer func() {
		if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
			err = nil
		}

		cancel()

		err = errors.Join(err, h.waitOwned(ctx))
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	if err := h.pump(owned); err != nil {
		return err
	}

	for {
		select {
		case <-owned.Done():
			return nil
		case result := <-h.finished:
			if err := h.complete(owned, result); err != nil {
				return err
			}
		case <-ticker.C:
			if err := h.pump(owned); err != nil {
				return err
			}
		}
	}
}

func (h *telegramWorkHost) pump(ctx context.Context) error {
	if err := h.collect(ctx); err != nil || len(h.active) == telegramWorkActivationLimit {
		return err
	}

	refs, err := h.receiver.store.TelegramWorkReferences(ctx, h.receiver.botID, h.after, telegramWorkActivationLimit)
	if err != nil {
		return err
	}

	h.after = ""
	for _, ref := range refs {
		h.after = ref.Request.Target.ConversationID

		if ctx.Err() != nil {
			return ctx.Err()
		}

		key := telegramWorkFlightKey(ref)
		if h.active[key] {
			continue
		}

		if len(h.active) == telegramWorkActivationLimit {
			break
		}

		h.active[key] = true
		go func() { h.finished <- h.receiver.activateWork(ctx, ref) }()
	}

	return nil
}

func (h *telegramWorkHost) collect(ctx context.Context) error {
	for {
		select {
		case result := <-h.finished:
			if err := h.complete(ctx, result); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (h *telegramWorkHost) complete(ctx context.Context, result telegramWorkCompletion) error {
	delete(h.active, telegramWorkFlightKey(result.Reference))

	return h.receiver.completeWorkActivation(ctx, result)
}

func (h *telegramWorkHost) waitOwned(ctx context.Context) (err error) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	for len(h.active) != 0 {
		result := <-h.finished
		if ctx.Err() != nil {
			// Native checkpoints survive shutdown. Reopen will project them;
			// cancelled authority readers cannot authorize a new notice here.
			delete(h.active, telegramWorkFlightKey(result.Reference))

			continue
		}

		err = errors.Join(err, h.complete(cleanup, result))
	}

	return err
}

func telegramWorkFlightKey(ref core.TelegramWorkReference) string {
	target := ref.Request.Target

	return target.ProjectID + ":" + target.ProjectBindingDigest + ":" + target.Namespace + ":" + ref.Invocation.ID
}
