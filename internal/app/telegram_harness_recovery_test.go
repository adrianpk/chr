package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

var errTelegramHarnessLostAck = errors.New("owned harness acknowledgement lost")

type telegramHarnessApplyProbe struct {
	service
	calls int
	lose  bool
}

func (s *telegramHarnessApplyProbe) ChangeTelegramHarnessSelections(ctx context.Context, target core.TelegramOperationTarget, request core.HarnessSelectionChangeRequest, authorize func(context.Context) error) (core.HarnessSelectionSnapshot, error) {
	s.calls++

	result, err := s.service.ChangeTelegramHarnessSelections(ctx, target, request, authorize)
	if err == nil && s.lose {
		s.lose = false

		return core.HarnessSelectionSnapshot{}, errTelegramHarnessLostAck
	}

	return result, err
}

// Restart after the shared owner commits but before the action ledger finishes
// recovers the original receipt, even when local profile setup becomes unreadable.
func TestTelegramHarnessSelectionLostAck(t *testing.T) {
	f, operations, _ := telegramHarnessFixture(t)
	f.send(t, "/project example-project", 0)

	request, action, _ := telegramHarnessOffer(t, f, "/harness worker=beta generation=pi-writing")
	probe := &telegramHarnessApplyProbe{service: telegramOperationConcreteService(t, f.receiver), lose: true}
	f.receiver.operations = probe
	queueTelegramOperationCallback(t, f, action)

	if _, err := f.receiver.step(t.Context()); !errors.Is(err, errTelegramHarnessLostAck) {
		t.Fatal("publication acknowledgement was not interrupted", err)
	}

	saved := telegramHarnessCheckSaved(t, operations, request.ID)
	telegramHarnessBreakProfiles(t, f, operations)
	reopenTelegramOperationReceiver(t, f)

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal("retained confirmation did not recover", err)
	}

	if notice := telegramOperationNotice(t, f); !strings.Contains(notice.Text, "Harness defaults saved") || probe.calls != 2 || webHarnessSnapshot(t, operations) != saved {
		t.Fatal("receipt recovery changed the published image", notice, probe.calls)
	}

	if notice := f.send(t, "/answer "+action.Token+" confirm", 0); !strings.Contains(notice.Text, "already confirmed") || probe.calls != 2 {
		t.Fatal("consumed replay reached the configuration owner", notice)
	}
}

// A lost Telegram send acknowledgement retries only the retained labelled
// notice. An unacknowledged prompt keeps its token and cannot duplicate writes.
func TestTelegramHarnessSelectionSendRetry(t *testing.T) {
	for _, stage := range []string{"prompt", "result"} {
		t.Run(stage, func(t *testing.T) {
			f, operations, _ := telegramHarnessFixture(t)
			f.send(t, "/project example-project", 0)

			probe := &telegramHarnessApplyProbe{service: telegramOperationConcreteService(t, f.receiver)}
			f.receiver.operations = probe

			client := &telegramDropSendAck{client: f.receiver.api.Client}
			if stage == "prompt" {
				f.receiver.api.Client = client
			}

			request, action, _ := telegramHarnessOffer(t, f, "/harness worker=beta generation=pi-writing")
			setTelegramOperationRetryAt(t, f.receiver.env, time.Now().Add(time.Hour))

			f.receiver.api.Client = client
			queueTelegramOperationCallback(t, f, action)

			if _, err := f.receiver.step(t.Context()); err != nil {
				t.Fatal(err)
			}

			telegramOperationNotice(t, f)

			saved := telegramHarnessCheckSaved(t, operations, request.ID)
			reopenTelegramOperationReceiver(t, f)
			setTelegramOperationRetryAt(t, f.receiver.env, time.Now().Add(-time.Minute))

			if err := f.receiver.drain(t.Context()); err != nil {
				t.Fatal(err)
			}

			telegramHarnessCheckSendRetry(t, f, operations, probe.calls, saved, stage, action.Token)
		})
	}
}

// Exact transport redelivery creates one prompt and one source-owned request;
// changing the answer after consumption cannot replace its confirmed choices.
func TestTelegramHarnessSelectionRedelivery(t *testing.T) {
	f, operations, _ := telegramHarnessFixture(t)
	f.send(t, "/project example-project", 0)

	prompt := f.send(t, "/harness worker=beta generation=pi-writing", 0)

	_, after, found := strings.Cut(prompt.Text, "/answer ")
	if !found {
		t.Fatal("prompt omitted its exact answer token", prompt)
	}

	token := strings.Fields(after)[0]
	before := webHarnessSnapshot(t, operations)

	f.updates <- f.last

	if _, err := f.receiver.step(t.Context()); err != nil || len(f.notices) != 0 || len(f.sessions(t)) != 1 || webHarnessSnapshot(t, operations) != before {
		t.Fatal("redelivery duplicated a prompt or applied selections", err)
	}

	if notice := f.send(t, "/answer "+token+" confirm", 0); !strings.Contains(notice.Text, "Harness defaults saved") {
		t.Fatal("retained prompt could not be confirmed", notice)
	}

	saved := webHarnessSnapshot(t, operations)

	f.updates <- f.last

	if _, err := f.receiver.step(t.Context()); err != nil || len(f.notices) != 0 || webHarnessSnapshot(t, operations) != saved {
		t.Fatal("answer redelivery repeated publication", err)
	}
}

func telegramHarnessCheckSendRetry(t *testing.T, f *telegramConversationAppFixture, operations service, calls int, saved core.HarnessSelectionSnapshot, stage, token string) {
	t.Helper()

	retry := telegramOperationNotice(t, f)
	if !strings.Contains(retry.Text, "Delivery confirmation was lost") || saved.RequestID == "" || calls != 1 || webHarnessSnapshot(t, operations) != saved {
		t.Fatal("uncertain send repeated a selection mutation", retry, calls)
	}

	if stage == "prompt" && !strings.Contains(retry.Text, token) {
		t.Fatal("prompt retry replaced its confirmation token")
	}
}
