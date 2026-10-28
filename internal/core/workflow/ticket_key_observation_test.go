package workflow

import (
	"path/filepath"
	"testing"
)

// Key recovery includes native envelopes and other namespaces after a cache loss.
func TestTicketKeyObservation(t *testing.T) {
	store := workflowTestStore(t)
	if _, err := store.Create(t.Context(), "Reference CHR-1", workflowTestTask("CHR-91")); err != nil {
		t.Fatal(err)
	}

	next, err := NextTicketKeyNumber(t.Context(), store.layout.ProjectDir(), "CHR")
	if err != nil || next != 92 {
		t.Fatal("native key was hidden behind display text", next, err)
	}

	writeTicketFile(t, filepath.Join(store.layout.ProjectDir(), "ops", "team", "ticket", "CHR-110.md"), "---\nkey: CHR-110\ntitle: Other namespace\nstatus: backlog\n---\n")

	next, err = NextTicketKeyNumber(t.Context(), store.layout.ProjectDir(), "CHR")
	if err != nil || next != 111 {
		t.Fatal("namespace switch reused an existing ticket key", next, err)
	}
}
