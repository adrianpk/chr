package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramGuidedAppendAck struct {
	service
	lost error
}

func applyGuidedTelegramControls(t *testing.T, f *telegramConversationAppFixture, commands ...string) {
	t.Helper()

	for index, text := range commands {
		update := retainTelegramHelpInput(t, f.receiver, int64(101+index), text)
		if err := f.receiver.respond(t.Context(), update); err != nil {
			t.Fatal(err)
		}
	}
}

func (s telegramGuidedAppendAck) ApplyTelegramGuided(ctx context.Context, target core.TelegramOperationTarget, input telegramGuidedInput) (core.AuthoringSession, error) {
	session, err := s.service.ApplyTelegramGuided(ctx, target, input)
	if err == nil && input.Command.Action == "append" {
		return core.AuthoringSession{}, s.lost
	}

	return session, err
}

// TestGuidedUseCaseTelegramOvertakenAck confirms an original canonical effect
// after a newer inspection and pause without replacing their current reply fence.
func TestGuidedUseCaseTelegramOvertakenAck(t *testing.T) {
	f, operations := guidedTelegramFixture(t)
	start := f.send(t, "Develop a new use case", 0)
	f.send(t, "Acknowledgement recovery", start.ID)

	lost := errors.New("lost append acknowledgement")
	f.receiver.operations = telegramGuidedAppendAck{service: operations, lost: lost}

	update := retainTelegramHelpInput(t, f.receiver, 100, "Preserve this capture once.")
	if err := f.receiver.respond(t.Context(), update); !errors.Is(err, lost) {
		t.Fatal("fault did not follow the canonical append", err)
	}

	f.receiver.operations = operations
	applyGuidedTelegramControls(t, f, "Show current step", "Pause use case")

	current := f.sessions(t)[0]

	before := readGuidedTelegramSession(t, operations, current)
	if before.Guided.Status != "paused" || strings.Count(before.Guided.Capture, "Preserve this capture once.") != 1 {
		t.Fatal("newer control or original append was not committed")
	}

	telegramAuthoringReopen(t, f)

	if err := f.receiver.respond(t.Context(), update); err != nil {
		t.Fatal("original acknowledgement cannot recover after a newer control", err)
	}

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if after.Revision != before.Revision || f.sessions(t)[0] != current {
		t.Fatal("acknowledgement replay replaced canonical state or the newer display")
	}

	request, found, err := f.receiver.store.TelegramOperation(t.Context(), update, 0)
	if err != nil || !found || request.State != "confirmed" {
		t.Fatal("original effect lacks its receiver receipt", err)
	}
}

// TestGuidedUseCaseTelegramOvertakenButton preserves a later resume when the
// original pause button's owner acknowledgement is recovered after restart.
func TestGuidedUseCaseTelegramOvertakenButton(t *testing.T) {
	f, operations := guidedTelegramFixture(t)
	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Button recovery", start.ID)
	controls := f.send(t, "Show controls", name.ID)
	request, action := guidedTelegramAction(t, f, controls)
	lost := errors.New("lost pause acknowledgement")
	f.receiver.operations = &telegramGuidedLostAck{service: operations, lost: lost}
	queueGuidedTelegramCallback(t, f, action, "pause", 100001)

	if _, err := f.receiver.step(t.Context()); !errors.Is(err, lost) {
		t.Fatal("pause acknowledgement was not lost", err)
	}

	f.receiver.operations = operations
	applyGuidedTelegramControls(t, f, "Show current step", "Resume use case")

	current := f.sessions(t)[0]

	before := readGuidedTelegramSession(t, operations, current)
	if before.Guided.Status == "paused" {
		t.Fatal("newer resume did not apply")
	}

	telegramAuthoringReopen(t, f)

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal("button acknowledgement cannot recover", err)
	}

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	retained, found, err := f.receiver.store.TelegramOperation(t.Context(), request.Source, 0)
	if err != nil || !found || retained.State != "confirmed" || after.Revision != before.Revision || f.sessions(t)[0] != current {
		t.Fatal("button replay lost its receipt or replaced newer state", err)
	}
}
