package core

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

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

	board, err := workflowcore.ReadTicketBoard(t.Context(), workflowcore.TicketBoardReadOptions{Layout: layout, ProjectID: "project-1"})
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

			if _, err := workflowcore.ReadTicketBoard(t.Context(), workflowcore.TicketBoardReadOptions{Layout: layout, ProjectID: "project-1"}); err == nil {
				t.Fatal("invalid store became a successful empty Board")
			}
		})
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: t.TempDir(), Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	for _, offset := range []int{-1, workflowstate.RecordLimit + 1} {
		if _, err := workflowcore.ReadTicketBoard(context.Background(), workflowcore.TicketBoardReadOptions{Layout: layout, ProjectID: "project-1", Offset: offset}); err == nil {
			t.Fatal("out-of-bounds page accepted")
		}
	}
}
