package app

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestTelegramControlModelLoad runs the production receiver against HTTP model
// and Telegram fixtures. Status and cancellation bypass the blocked model call.
func TestTelegramControlModelLoad(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	f.send(t, "/prepare --project example-project original", 0)

	c := f.sessions(t)[0]
	entered, release := make(chan struct{}, 2), make(chan struct{})

	var once sync.Once

	calls := telegramAuthoringIntentHTTP(t, f, func(input ConversationIntentContext) string {
		entered <- struct{}{}

		<-release

		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "capture", Arguments: ConversationIntentArguments{Text: input.Message}})
	})
	telegramControlRunReceiver(t, f.receiver, func() { once.Do(func() { close(release) }) })

	telegramWorkQueueMessage(t, f, "capture before cancellation")

	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("model did not enter")
	}

	telegramWorkQueueMessage(t, f, "/status")
	telegramWaitNotice(t, f, "Receiver:", 3*time.Second)
	telegramWorkQueueMessage(t, f, "/cancel "+c.ID)
	telegramWaitNotice(t, f, "Canceled", 3*time.Second)

	if current := f.session(t, c.ID); current.Phase != "canceled" || current.Capture != "" || calls.Load() != 1 {
		t.Fatal("stop waited for or duplicated model work", current, calls.Load())
	}

	once.Do(func() { close(release) })
	telegramWaitNotice(t, f, "target or arguments", 3*time.Second)

	if current := f.session(t, c.ID); current.Phase != "canceled" || current.Capture != "" {
		t.Fatal("stale interpretation changed cancelled context", current)
	}
}

func telegramControlRunReceiver(t *testing.T, receiver telegramReceiver, release func()) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())

	finished := make(chan error, 1)
	go func() { finished <- receiver.run(ctx) }()

	t.Cleanup(func() {
		release()
		cancel()

		select {
		case err := <-finished:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(10 * time.Second):
			t.Error("receiver did not join owned calls")
		}
	})
}

func telegramWaitNotice(t *testing.T, f *telegramConversationAppFixture, text string, timeout time.Duration) telegramConversationNotice {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case notice := <-f.notices:
			if strings.Contains(notice.Text, text) {
				return notice
			}
		case <-timer.C:
			t.Fatal("notice unavailable", text)

			return telegramConversationNotice{}
		}
	}
}
