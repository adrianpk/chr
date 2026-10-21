package app

import (
	"strings"
	"testing"
)

// TestTelegramPresentationActionWork checks the actual HTTP preview against the
// native owner: human scope and consequences are visible before any admission.
func TestTelegramPresentationActionWork(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)
	_, prompt := telegramWorkOffer(t, f, "/work ticket "+record.Task.Key)
	for _, part := range []string{record.Task.Key, "Selected tickets:", "Publication:", "Limits:", "validation", "Confirming admits"} {
		if !strings.Contains(prompt.Text, part) {
			t.Fatal("start preview omitted scope or consequence", part, prompt)
		}
	}
	for _, audit := range []string{"Snapshot:", "Policy revision:", "Target commit:"} {
		if strings.Contains(prompt.Text, audit) {
			t.Fatal("routine caption exposed an audit field", audit)
		}
	}
	checkTelegramWorkRunCount(t, operations, 0)
}

// TestTelegramPresentationControlPreview distinguishes accepting a pause request
// from quiescence and verifies that presenting the control does not change work.
func TestTelegramPresentationControlPreview(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)
	before := checkTelegramWorkRunCount(t, operations, 1)
	prompt := f.send(t, "/pause "+ref.Invocation.ID, 0)
	if !strings.Contains(prompt.Text, "Request pause") || !strings.Contains(prompt.Text, "worker may still be stopping") {
		t.Fatal("pause preview claimed a completed stop", prompt)
	}
	after := checkTelegramWorkRunCount(t, operations, 1)
	if after.Revision != before.Revision || after.StopIntent != before.StopIntent {
		t.Fatal("preview changed the native invocation")
	}
}
