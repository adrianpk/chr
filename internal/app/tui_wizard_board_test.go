package app

import (
	"strings"
	"testing"
)

// Execution uses the actual case-filtered Board and retained native counters;
// opening the Board does not approve, continue or replace the invocation.
func TestTUIWizardBoard(t *testing.T) {
	operations, session, _, _ := guidedResultFixture(t)
	before := assertGuidedNativeAdmission(t, operations, session)
	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)

	if !strings.Contains(model.View(), "Tickets: 1") || !strings.Contains(model.View(), before.Members[0].Admission.Key) {
		t.Fatal("wizard replaced the shared Board", model.View())
	}

	if !strings.Contains(model.View(), "1 of 1 validated; 0 confirmed; 1 remaining") || !strings.Contains(model.View(), "Native review requires approval") {
		t.Fatal("Board hid retained progress or review", model.View())
	}

	after := assertGuidedNativeAdmission(t, operations, session)
	if after.Revision != before.Revision || serviceWorkTaskAttempts(after) != serviceWorkTaskAttempts(before) {
		t.Fatal("Board observation dispatched native work")
	}
}
