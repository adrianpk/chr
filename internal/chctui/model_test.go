package chctui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestModelViewRendersProjectAndNavigation(t *testing.T) {
	model := NewModel(Options{
		ProjectDir: "/workspace/project",
		ProjectID:  "project-1",
	})

	view := model.View()

	for _, want := range []string{
		"ChrysoCoord",
		"Project",
		"project-1",
		"/workspace/project",
		"> Dashboard",
		"Board",
		"Tickets",
		"Runtime",
		"Keys:",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelNavigatesViews(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.SelectedView() != ViewBoard {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewBoard)
	}

	updated, _ = navigated.Update(tea.KeyMsg{Type: tea.KeyLeft})
	navigated, ok = updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.SelectedView() != ViewDashboard {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewDashboard)
	}
}

func TestModelRendersTicketBoardAndDetail(t *testing.T) {
	projectDir := t.TempDir()
	firstPath := writeTicketViewFile(t, projectDir, "open.md", "- [x] Done\n- [ ] Pending\n\n# Comments\n\n## 2026-05-22T10:00:00Z chc\n")
	secondPath := writeTicketViewFile(t, projectDir, "ready.md", "- [ ] Ready\n")

	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{
				Path:         firstPath,
				RelativePath: "open.md",
				Key:          "CHC-1",
				Title:        "Fix publishing",
				Status:       "open",
				Priority:     "high",
				Severity:     "critical",
				Type:         "bug",
				UseCase:      "cli-blog",
				ProjectArea:  "indexing",
			},
			{
				Path:         secondPath,
				RelativePath: "ready.md",
				Key:          "CHC-2",
				Title:        "Add completion",
				Status:       "ready",
				Priority:     "normal",
				Severity:     "normal",
				Type:         "feature",
				UseCase:      "cli-blog",
				ProjectArea:  "tasks",
			},
		},
	})
	model.selected = 1

	view := model.View()

	for _, want := range []string{
		"Board",
		"open",
		"ready",
		"CHC-1",
		"Fix publishing",
		"Selected Ticket",
		"Checklist: 1 done / 2 total",
		"Comments: 1",
		"Path: open.md",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelSelectsTickets(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
			{Key: "CHC-2", Title: "Second", Status: "ready"},
		},
	})
	model.selected = 1

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	ticket, ok := navigated.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	if ticket.Key != "CHC-2" {
		t.Fatalf("selected ticket = %q, want CHC-2", ticket.Key)
	}
}

func TestTicketFilterMatchesMetadata(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "Fix publishing", Status: "open"},
			{Key: "CHC-2", Title: "Add completion", Status: "ready"},
		},
	})
	model.ticketFilter = "completion"

	tickets := model.filteredTickets()
	if len(tickets) != 1 {
		t.Fatalf("len(filteredTickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHC-2" {
		t.Fatalf("filtered ticket = %q, want CHC-2", tickets[0].Key)
	}
}

func TestModelRendersDocumentsAndDetail(t *testing.T) {
	model := NewModel(Options{
		Documents: []DocumentView{
			{
				Index:        0,
				Family:       "spec",
				Status:       "accepted",
				Title:        "TUI Navigation",
				RelativePath: ".prj/specs/specs/tui-navigation.md",
				Preview:      "Navigation details.",
			},
		},
	})
	model.selected = 3

	view := model.View()

	for _, want := range []string{
		"Documents",
		"spec",
		"accepted",
		"TUI Navigation",
		"Selected Document",
		"Path: .prj/specs/specs/tui-navigation.md",
		"Navigation details.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestLoadDocumentViewsDiscoversProjectMarkdown(t *testing.T) {
	projectDir := t.TempDir()
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "specs", "specs"), "feature.md", "# Feature Spec\n\nStatus: accepted\n\nDetails.")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "trackers"), "feature-tracker.md", "# Feature Tracker\n\nStatus: planned\n")

	documents, err := LoadDocumentViews(t.Context(), projectDir)
	if err != nil {
		t.Fatalf("LoadDocumentViews returned error: %v", err)
	}

	if len(documents) != 2 {
		t.Fatalf("len(documents) = %d, want 2", len(documents))
	}

	if documents[0].Title != "Feature Spec" {
		t.Fatalf("first document title = %q, want Feature Spec", documents[0].Title)
	}

	if documents[0].Status != "accepted" {
		t.Fatalf("first document status = %q, want accepted", documents[0].Status)
	}
}

func TestModelRendersAuthoringSessionsAndDetail(t *testing.T) {
	updated := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	model := NewModel(Options{
		AuthoringSessions: []chc.AuthoringSession{
			{
				ID:        "auth-1",
				ProjectID: "project-1",
				Status:    chc.AuthoringSessionStatusActive,
				Mode:      chc.AuthoringSessionModeAssisted,
				Target:    chc.AuthoringTargetSpec,
				Topic:     "Billing sync",
				UpdatedAt: updated,
				Inputs: []chc.AuthoringArtifactRef{
					{Path: ".prj/dumps/billing.md"},
				},
				Outputs: []chc.AuthoringArtifactRef{
					{Path: ".chc/specs/auth-1-billing-sync.md"},
				},
			},
		},
	})
	model.selected = 4

	view := model.View()

	for _, want := range []string{
		"Authoring",
		"active",
		"spec",
		"Billing sync",
		"Selected Authoring Session",
		"ID: auth-1",
		"Transcript: .chc/authoring/sessions/auth-1/transcript.md",
		"Inputs: .prj/dumps/billing.md",
		"Outputs: .chc/specs/auth-1-billing-sync.md",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelSelectsDocuments(t *testing.T) {
	model := NewModel(Options{
		Documents: []DocumentView{
			{Index: 0, Title: "First"},
			{Index: 1, Title: "Second"},
		},
	})
	model.selected = 3

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	document, ok := navigated.selectedDocumentView()
	if !ok {
		t.Fatal("selectedDocumentView returned no document")
	}

	if document.Title != "Second" {
		t.Fatalf("selected document = %q, want Second", document.Title)
	}
}

func TestModelRecordsWindowSize(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	resized, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if resized.width != 120 {
		t.Fatalf("width = %d, want 120", resized.width)
	}

	if resized.height != 40 {
		t.Fatalf("height = %d, want 40", resized.height)
	}
}

func writeTicketViewFile(t *testing.T, projectDir, name, body string) string {
	t.Helper()

	err := os.MkdirAll(projectDir, 0o755)
	if err != nil {
		t.Fatalf("create ticket view directory: %v", err)
	}

	path := filepath.Join(projectDir, name)

	err = os.WriteFile(path, []byte(body), 0o644)
	if err != nil {
		t.Fatalf("write ticket view file: %v", err)
	}

	return path
}
