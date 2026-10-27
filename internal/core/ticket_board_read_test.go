package core

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Native authority wins over its derived Markdown; every bounded page remains reachable.
func TestTicketBoardNativePages(t *testing.T) {
	store := nativeBoardStore(t)
	writeTicketFile(t, UserPath(store.layout.ProjectDir(), "tickets", "projection.md"), "---\nkey: CHR-1\ntitle: Outdated projection\nstatus: done\n---\n")

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

func nativeBoardStore(t *testing.T) *WorkflowStore {
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

func checkNativeBoardPage(t *testing.T, opts TicketBoardReadOptions, loaded, next int) Board {
	t.Helper()

	page, err := ReadTicketBoard(t.Context(), opts)
	if err != nil || page.TotalCount != 130 || page.LoadedCount != loaded || page.Columns[1].Count != 130 || page.NextOffset != next {
		t.Fatalf("page = %#v, %v", page, err)
	}

	return page
}

func checkNativeBoardTickets(t *testing.T, tickets []IndexedTicket) {
	t.Helper()

	seen := map[string]bool{}
	for _, ticket := range tickets {
		if seen[ticket.Key] || !ticket.NativeWorkflow || ticket.Status != TicketStatusReadyForDev || !strings.HasPrefix(ticket.Title, "Native ") {
			t.Fatal("native identity duplicated or replaced by its projection", ticket)
		}

		seen[ticket.Key] = true
	}

	if len(seen) != 130 {
		t.Fatal("pagination lost records")
	}
}

// Repository status aliases differ from runtime aliases, and invalid data stays visible.
func TestTicketBoardAttention(t *testing.T) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: t.TempDir(), Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct{ path, content string }{
		{"solved/task.md", "---\nid: TKT-20261018000000\ntitle: Resolved repository task\nstatus: solved\nkind: task\nseverity: medium\n---\nDescription.\n"},
		{"legacy.md", "---\nkey: CHR-1\ntitle: Runtime awaiting merge\nstatus: solved\n---\nDescription.\n"},
		{"blocked.md", "---\nkey: CHR-2\ntitle: Blocked\nstatus: blocked\n---\nDescription.\n"},
		{"invalid.md", "Unreadable frontmatter but retained body.\n"},
	} {
		writeTicketFile(t, filepath.Join(layout.Root(), "ticket", tc.path), tc.content)
	}

	board, err := ReadTicketBoard(t.Context(), TicketBoardReadOptions{Layout: layout, ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("attention and alias projection = %#v, %v", board, err)
	}

	if !slices.Equal([]int{board.TotalCount, board.NormalCount, board.AttentionCount, board.Columns[4].Count, board.Columns[5].Count}, []int{4, 2, 2, 1, 1}) {
		t.Fatal("incorrect attention, legacy merge or repository resolution counts", board)
	}

	for _, attention := range board.Attention {
		if len(attention.Reasons) == 0 || attention.Ticket.Path == "" {
			t.Fatal("attention record lost its reason or document identity", attention)
		}
	}
}

// Failed storage reads, ambiguous identities and invalid page bounds cannot look empty.
func TestTicketBoardFailures(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*testing.T, opscore.Layout)
	}{
		{"duplicate", func(t *testing.T, layout opscore.Layout) {
			for _, name := range []string{"first.md", "second.md"} {
				writeTicketFile(t, filepath.Join(layout.Root(), "ticket", name), "---\nkey: CHR-1\ntitle: Duplicate\nstatus: backlog\n---\n")
			}
		}},
		{"symlink", func(t *testing.T, layout opscore.Layout) {
			writeTicketFile(t, filepath.Join(layout.Root(), "ticket", "target.txt"), "body")

			if err := os.Symlink("target.txt", filepath.Join(layout.Root(), "ticket", "linked.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{"storage", func(t *testing.T, layout opscore.Layout) {
			writeTicketFile(t, filepath.Join(layout.Root(), "ticket"), "not a directory")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: t.TempDir(), Namespace: "default"})
			if err != nil {
				t.Fatal(err)
			}

			tc.setup(t, layout)

			if _, err := ReadTicketBoard(t.Context(), TicketBoardReadOptions{Layout: layout, ProjectID: "project-1"}); err == nil {
				t.Fatal("invalid store became a successful empty Board")
			}
		})
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: t.TempDir(), Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	for _, offset := range []int{-1, workflowstate.RecordLimit + 1} {
		if _, err := ReadTicketBoard(context.Background(), TicketBoardReadOptions{Layout: layout, ProjectID: "project-1", Offset: offset}); err == nil {
			t.Fatal("out-of-bounds page accepted")
		}
	}
}
