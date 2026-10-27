package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Finalization reuses its prepared source and confirmed result after later dump
// writes; an unchanged retry never repeats curation or appends context twice.
func TestAuthorAssistedDumpRetry(t *testing.T) {
	root := newTestWorkspace(t)
	sessionID := createAuthorAssistedSessionForTest(t, root)
	args := []string{"author", "assisted", "finalize", "--dir", root, "--session", sessionID, "--fake-summary", "Retain the retry ceiling.", "--target-kind", "dump", "--target-name", "  Project context  ", "--operation-id", "finalize-fixture"}

	var first, replay bytes.Buffer
	if err := run(t.Context(), args, &first, &bytes.Buffer{}); err != nil {
		t.Fatal("first finalization failed", err)
	}

	if _, err := core.MutateDump(t.Context(), root, core.DumpMutationRequest{OperationID: "ordinary", Action: "capture", Title: "Later context", Content: "Ordinary captured material."}, nil); err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), args, &replay, &bytes.Buffer{}); err != nil || replay.String() != first.String() {
		t.Fatal("finalization retry did not return original result", replay.String(), err)
	}

	checkAuthorAssistedDumpCapture(t, root)

	changed := append([]string(nil), args...)
	changed[8] = "Changed requirements."

	if err := run(t.Context(), changed, &bytes.Buffer{}, &bytes.Buffer{}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed finalization arguments accepted", err)
	}
}

func checkAuthorAssistedDumpCapture(t *testing.T, root string) {
	t.Helper()

	observed, err := core.ObserveProjectDump(t.Context(), root)
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 2 {
		t.Fatal("finalization duplicated captures", observed, err)
	}

	document, err := core.ReadDumpDocument(t.Context(), root, observed.Destination)
	if err != nil || strings.Count(document.Content, "Retain the retry ceiling.") != 1 || !strings.Contains(document.Content, "Ordinary captured material.") {
		t.Fatal("retry overwrote later material or duplicated its contribution", document, err)
	}
}
