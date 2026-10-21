package app

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"forge.adrianpk.com/chr/internal/core"
)

// Progress distinguishes native validation, waiting and confirmed integration;
// long dynamic titles cannot displace counts, truth or the existing inspection route.
func TestTelegramPresentationProgressStates(t *testing.T) {
	cases := []struct {
		name, state, stop, outcome, receipt, hash, want string
		confirmed, units                                int
	}{
		{name: "validated", state: "waiting_review", want: "No confirmed integration receipt"},
		{name: "completed without receipt", state: "completed", want: "No confirmed integration receipt"},
		{name: "receipt without hash", state: "completed", receipt: "receipt/native", want: "No confirmed integration receipt"},
		{name: "partial integration", state: "running", receipt: "receipt/native", hash: "native-hash", want: "Integration confirmed for 1 unit(s)", confirmed: 1, units: 1},
		{name: "integration wait", state: "waiting_integration", want: "waiting for integration"},
		{name: "admission", state: "pending", want: "execution has not started"},
		{name: "capacity", state: "waiting_capacity", want: "waiting for execution capacity"},
		{name: "stop pending", state: "running", stop: "pause", want: "may still be stopping"},
		{name: "recovery", state: "paused", outcome: "recovered", want: "Recovered the same saved work"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			status := WorkflowWorkStatus{WorkflowIdentity: core.WorkflowIdentity{ID: "run-original", Revision: 3}, WorkTitle: strings.Repeat("README `title` 😀 ", 100), Status: item.state, StopIntent: item.stop, Outcome: item.outcome, ValidatedTasks: 1, Total: 2, ConfirmedTickets: item.confirmed, IntegratedUnits: item.units, Remaining: 2 - item.confirmed, DeliveryRef: item.receipt, DeliveryHash: item.hash, NextActionOwner: "operator", CandidateHash: "private-candidate-hash", ApprovalHash: "private-review-hash"}

			text := core.RenderTelegramMessage(telegramWorkProgressMessage(status, "Example project")).PlainFallback
			for _, value := range []string{item.want, "Validated tasks: 1/2", fmt.Sprintf("confirmed tickets: %d", item.confirmed), "Example project", "chr work status --json run-original"} {
				if !strings.Contains(text, value) {
					t.Fatalf("missing %q in %s", value, text)
				}
			}

			if len(utf16.Encode([]rune(text))) > 1800 || !utf8.ValidString(text) || strings.Contains(text, status.CandidateHash) || strings.Contains(text, status.ApprovalHash) {
				t.Fatal("unbounded or audit-heavy progress", text)
			}
		})
	}
}
