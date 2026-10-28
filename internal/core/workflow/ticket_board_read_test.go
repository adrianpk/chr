package workflow

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Native authority wins over its derived Markdown; every bounded page remains reachable.
func TestTicketBoardNativePages(t *testing.T) {
	store := nativeBoardStore(t)
	writeTicketFile(t, workspacecore.UserPath(store.layout.ProjectDir(), "tickets", "projection.md"), "---\nkey: CHR-1\ntitle: Outdated projection\nstatus: done\n---\n")

	before := nativeQueueImages(t, store.layout.Root())
	opts := TicketBoardReadOptions{Layout: store.layout, ProjectID: store.projectID, NativeWorkflow: true}

	first := checkNativeBoardPage(t, opts, 128, 128)
	opts.Offset = first.NextOffset

	second := checkNativeBoardPage(t, opts, 2, -1)
	checkNativeBoardTickets(t, append(first.Tickets, second.Tickets...))

	if first.Revision != second.Revision || !maps.Equal(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("Board read changed canonical documents or the source revision")
	}
}

func nativeBoardStore(t *testing.T) *Store {
	t.Helper()

	store := workflowTestStore(t)

	for i := 1; i <= workflowstate.PageLimit+2; i++ {
		record := workflowTestTask(fmt.Sprintf("CHR-%d", i))
		record.Metadata = map[string]json.RawMessage{"title": json.RawMessage(fmt.Sprintf("\"Native %d\"", i))}

		data, err := encodeWorkflowRecord(record)
		if err != nil {
			t.Fatal(err)
		}

		writeTicketFile(t, filepath.Join(store.layout.Root(), record.Path), string(data))
		writeTicketFile(t, filepath.Join(store.layout.Root(), filepath.Dir(record.Path), "artifact.md"), string(record.Body))
	}

	return store
}

func checkNativeBoardPage(t *testing.T, opts TicketBoardReadOptions, loaded, next int) ticketcore.Board {
	t.Helper()

	page, err := ReadTicketBoard(t.Context(), opts)
	if err != nil || page.TotalCount != 130 || page.LoadedCount != loaded || page.Columns[1].Count != 130 || page.NextOffset != next {
		t.Fatalf("page = %#v, %v", page, err)
	}

	return page
}

func checkNativeBoardTickets(t *testing.T, tickets []ticketcore.Indexed) {
	t.Helper()

	seen := map[string]bool{}
	for _, ticket := range tickets {
		if seen[ticket.Key] || !ticket.NativeWorkflow || ticket.Status != ticketcore.StatusReadyForDev || !strings.HasPrefix(ticket.Title, "Native ") {
			t.Fatal("native identity duplicated or replaced by its projection", ticket)
		}

		seen[ticket.Key] = true
	}

	if len(seen) != 130 {
		t.Fatal("pagination lost records")
	}
}
