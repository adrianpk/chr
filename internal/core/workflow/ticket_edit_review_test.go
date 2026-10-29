package workflow

import (
	"testing"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// Verified delivery references remain immutable even when callers bypass UI
// field restrictions and submit a separately captured reference operation.
func TestTicketReviewReceiptGuard(t *testing.T) {
	document := TicketDocument{ReadOnly: map[string]string{"review_refs": "The verified delivery receipt owns this value."}}

	request := TicketEditRequest{ReviewRef: &ticketmetadata.ReviewRef{Type: "pull_request", Provider: "forgejo", URL: "https://forge.example.org/team/project/pulls/2"}}
	if _, err := prepareTicketReviewReference("retained", document, request); err == nil {
		t.Fatal("verified receipt reference was mutable")
	}
}
