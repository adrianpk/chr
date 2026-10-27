package core

import (
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// Real SQLite projection freezes one bounded view under its original native
// revision, including uncertain-send recovery after reopening the store.
func TestTelegramPresentationProgressRetention(t *testing.T) {
	f := newTelegramConversationFixture(t)
	ref := telegramWorkReferenceFixture(t, f, "run-presentation")
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	telegramWorkDeliverInitial(t, f, now)

	message := tgmessage.Blocks{Family: "33", InitialText: "README work in Example project", Content: strings.Repeat("literal `*` 😀\n", 500), SecondaryText: "No confirmed integration receipt. Inspect: chr work status --json run-presentation"}
	original := tgmessage.Render(message).PlainFallback

	observation := TelegramWorkObservation{Revision: 1, Phase: "executing", Kind: "progress", Text: original, Message: &message}
	if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	first := telegramWorkDueNotice(t, f, now, original)
	if first.Presentation == nil || first.Presentation.Family != "33" || len(utf16.Encode([]rune(first.Text))) > 1800 {
		t.Fatal("unfrozen or unbounded projection", first)
	}

	checkTelegramProgressRecovery(t, f, ref, observation, first, now)
}

func checkTelegramProgressRecovery(t *testing.T, f *telegramConversationFixture, ref TelegramWorkReference, observation TelegramWorkObservation, first TelegramReceiverDelivery, now time.Time) {
	t.Helper()

	original := first.Text

	if err := f.store.ReserveDelivery(t.Context(), first.ID, now); err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliveryFailed(t.Context(), first, &TelegramAPIError{Code: "transport_failure", Uncertain: true}, now); err != nil {
		t.Fatal(err)
	}

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, f.env)
	observation.Message.InitialText = "Changed mutable display"

	if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	retry := telegramWorkDueNotice(t, f, now.Add(time.Minute), original)
	if retry.ID != first.ID || retry.Attempts != 1 || retry.State != "uncertain" || retry.Presentation == nil || retry.Presentation.PlainFallback != original {
		t.Fatal("recovery replaced revision, view or allowance", retry)
	}
}
