package app

import (
	"context"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramControlActivationBound runs real inbox calls with all ordinary
// activations occupied. Exactly two decision and two urgent calls can enter,
// produce durable responses and finish before any ordinary slot is released.
func TestTelegramControlActivationBound(t *testing.T) {
	f := newTelegramConversationAppFixture(t)

	f.receiver.slots = newTelegramActivationSlots()
	for range 12 {
		if !f.receiver.slots.acquire(core.TelegramProcessingOrdinary) {
			t.Fatal("ordinary allowance unavailable")
		}
		defer f.receiver.slots.release(core.TelegramProcessingOrdinary)
	}

	for _, text := range []string{"/help", "/status", "/status", "/status", "/answer unknown confirm", "/answer unknown confirm", "/answer unknown confirm"} {
		telegramOperationRetainInput(t, f, text)
	}

	ctx, cancel := context.WithCancel(t.Context())
	host := newTelegramApplicationHost(f.receiver)

	defer func() { cancel(); finishTelegramControlApplications(t, host) }()

	if err := host.pump(ctx); err != nil {
		t.Fatal(err)
	}

	if len(host.active) != 4 {
		t.Fatal("controls did not retain their bounded allowance", len(host.active))
	}

	checkTelegramControlSlotCeiling(t, host.receiver.slots)

	finishTelegramControlApplications(t, host)

	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	telegramWaitNotice(t, f, "Receiver:", time.Second)
	telegramWaitNotice(t, f, "unavailable, expired or stale", time.Second)

	pending, err := f.receiver.store.PendingUpdates(t.Context(), f.receiver.botID)
	if err != nil || len(pending) != 3 {
		t.Fatal("ordinary work advanced or a control escaped its ceiling", pending, err)
	}
}

func checkTelegramControlSlotCeiling(t *testing.T, slots *telegramActivationSlots) {
	t.Helper()

	for _, lane := range []core.TelegramProcessingLane{core.TelegramProcessingOrdinary, core.TelegramProcessingDecision, core.TelegramProcessingUrgent} {
		if slots.acquire(lane) {
			slots.release(lane)
			t.Fatal("application ceiling exceeded", lane)
		}
	}
}

func finishTelegramControlApplications(t *testing.T, host *telegramApplicationHost) {
	t.Helper()

	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	for len(host.active) > 0 {
		select {
		case result := <-host.finished:
			if err := host.complete(result); err != nil {
				t.Error(err)
			}
		case <-timer.C:
			t.Error("owned control applications did not finish")

			return
		}
	}
}
