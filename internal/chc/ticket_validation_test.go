package chc

import (
	"strings"
	"testing"
	"time"
)

func TestValidateIndexedTicketAcceptsBoardMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusBacklog
	ticket.Synopsis = "Implement the Board ticket creation path."
	ticket.SourceKind = "manual"

	err := ValidateIndexedTicket(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketAcceptsLegacyTicketWithoutSynopsis(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusOpenLegacy
	ticket.Synopsis = ""

	err := ValidateIndexedTicket(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = "unknown"

	err := ValidateIndexedTicket(ticket)
	if err == nil {
		t.Fatal("ValidateIndexedTicket returned nil error")
	}

	if !strings.Contains(err.Error(), "invalid ticket status: unknown") {
		t.Fatalf("error = %q, want invalid status", err.Error())
	}
}

func validIndexedTicket(now time.Time) IndexedTicket {
	return IndexedTicket{
		ProjectID:    "project-1",
		Path:         "/tmp/project/.chc/user/tickets/CHC-1-Test.md",
		RelativePath: "CHC-1-Test.md",
		Key:          "CHC-1",
		Title:        "Test ticket",
		Status:       TicketStatusBacklog,
		Priority:     "normal",
		Severity:     "normal",
		Type:         "task",
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
}
