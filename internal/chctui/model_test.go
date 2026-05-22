package chctui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

func TestModelOpensAndFiltersCommandPalette(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	palette, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if !palette.commandPalette.Open {
		t.Fatal("command palette is closed, want open")
	}

	updated, _ = palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("runtime")})

	palette, ok = updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	commands := palette.filteredCommands()
	if len(commands) == 0 {
		t.Fatal("filteredCommands returned no commands")
	}

	if commands[0].View != ViewRuntime {
		t.Fatalf("first filtered command view = %q, want %q", commands[0].View, ViewRuntime)
	}

	view := palette.View()
	for _, want := range []string{
		"Command Palette",
		"Filter: runtime",
		"Go to Runtime",
		"enter runs command",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelExecutesNavigationCommand(t *testing.T) {
	model := NewModel(Options{})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "runtime"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	navigated, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.commandPalette.Open {
		t.Fatal("command palette is open after command execution")
	}

	if navigated.SelectedView() != ViewRuntime {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewRuntime)
	}

	if navigated.statusMessage != "Navigated to Runtime." {
		t.Fatalf("statusMessage = %q, want navigation status", navigated.statusMessage)
	}
}

func TestModelReportsSelectedSourcePathCommand(t *testing.T) {
	model := NewModel(Options{
		Documents: []DocumentView{
			{Index: 0, Title: "Spec", RelativePath: ".prj/specs/specs/tui.md"},
		},
	})
	model.selected = 3
	model.commandPalette.Open = true
	model.commandPalette.Filter = "source"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	navigated, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.statusMessage != "Selected source path: .prj/specs/specs/tui.md" {
		t.Fatalf("statusMessage = %q, want selected source path", navigated.statusMessage)
	}
}

func TestModelPlansArtifactActionCommand(t *testing.T) {
	model := NewModel(Options{})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "create-ticket"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if planned.commandPalette.Open {
		t.Fatal("command palette is open after artifact action planning")
	}

	if !planned.artifactAction.Pending {
		t.Fatal("artifact action is not pending")
	}

	if planned.artifactAction.Action != ArtifactActionCreateTicket {
		t.Fatalf("artifact action = %q, want %q", planned.artifactAction.Action, ArtifactActionCreateTicket)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Action",
		"Create ticket",
		"Title: TUI created ticket",
		"enter confirms action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelCancelsArtifactAction(t *testing.T) {
	model := NewModel(Options{})
	model.artifactAction = NewArtifactActionState(ArtifactActionCreateTicket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.artifactAction.Pending {
		t.Fatal("artifact action is pending after cancellation")
	}

	if cancelled.statusMessage != "Action cancelled." {
		t.Fatalf("statusMessage = %q, want cancellation status", cancelled.statusMessage)
	}
}

func TestModelConfirmsArtifactAction(t *testing.T) {
	executor := &fakeArtifactActionExecutor{
		result: ArtifactActionResult{
			Message: "Created ticket CHC-1 at .chc/tickets/CHC-1-tui-created-ticket.md",
			Tickets: []chc.IndexedTicket{
				{Key: "CHC-1", Title: "TUI created ticket", Status: "open"},
			},
			Documents: []DocumentView{
				{Index: 0, Title: "TUI Plan"},
			},
			AuthoringSessions: []chc.AuthoringSession{
				{
					ID:        "author-20260522-100000",
					ProjectID: "project-1",
					Status:    chc.AuthoringSessionStatusDraft,
					Mode:      chc.AuthoringSessionModeTUI,
					Target:    chc.AuthoringTargetDraft,
					Topic:     "TUI artifact creation",
				},
			},
		},
	}
	model := NewModel(Options{ActionExecutor: executor})
	model.artifactAction = NewArtifactActionState(ArtifactActionCreateTicket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.action != ArtifactActionCreateTicket {
		t.Fatalf("executor action = %q, want %q", executor.action, ArtifactActionCreateTicket)
	}

	if confirmed.artifactAction.Pending {
		t.Fatal("artifact action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	assertConfirmedArtifactState(t, confirmed)
}

func TestModelReportsArtifactActionError(t *testing.T) {
	executor := &fakeArtifactActionExecutor{err: errors.New("write failed")}
	model := NewModel(Options{ActionExecutor: executor})
	model.artifactAction = NewArtifactActionState(ArtifactActionCreateTicket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if failed.artifactAction.Pending {
		t.Fatal("artifact action is pending after failure")
	}

	if failed.statusMessage != "Action failed: write failed" {
		t.Fatalf("statusMessage = %q, want action failure", failed.statusMessage)
	}
}

func TestModelPlansTicketActionCommand(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "ticket-status-ready"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if planned.commandPalette.Open {
		t.Fatal("command palette is open after ticket action planning")
	}

	if !planned.ticketAction.Pending {
		t.Fatal("ticket action is not pending")
	}

	if planned.ticketAction.Request.Key != "CHC-1" {
		t.Fatalf("ticket action key = %q, want CHC-1", planned.ticketAction.Request.Key)
	}

	if planned.ticketAction.Request.Status != "ready" {
		t.Fatalf("ticket action status = %q, want ready", planned.ticketAction.Request.Status)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Ticket Action",
		"Move ticket status",
		"Ticket: CHC-1",
		"New status: ready",
		"enter confirms ticket action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelCancelsTicketAction(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.ticketAction = NewTicketActionState(TicketActionAddComment, ticket, "")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.ticketAction.Pending {
		t.Fatal("ticket action is pending after cancellation")
	}

	if cancelled.statusMessage != "Ticket action cancelled." {
		t.Fatalf("statusMessage = %q, want ticket cancellation status", cancelled.statusMessage)
	}
}

func TestModelConfirmsTicketAction(t *testing.T) {
	executor := &fakeTicketActionExecutor{
		result: TicketActionResult{
			Message: "Moved ticket CHC-1 to ready.",
			Tickets: []chc.IndexedTicket{
				{Key: "CHC-1", Title: "First", Status: "ready"},
			},
		},
	}
	model := NewModel(Options{
		TicketExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.ticketAction = NewTicketActionState(TicketActionMoveStatus, ticket, "ready")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != TicketActionMoveStatus {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, TicketActionMoveStatus)
	}

	if executor.request.Key != "CHC-1" || executor.request.Status != "ready" {
		t.Fatalf("executor request = %#v, want CHC-1 ready", executor.request)
	}

	if confirmed.ticketAction.Pending {
		t.Fatal("ticket action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	if len(confirmed.tickets) != 1 || confirmed.tickets[0].Status != "ready" {
		t.Fatalf("tickets = %#v, want reloaded ready ticket", confirmed.tickets)
	}
}

func TestModelReportsTicketActionError(t *testing.T) {
	executor := &fakeTicketActionExecutor{err: errors.New("invalid ticket status")}
	model := NewModel(Options{
		TicketExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.ticketAction = NewTicketActionState(TicketActionMoveStatus, ticket, "unknown")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if failed.ticketAction.Pending {
		t.Fatal("ticket action is pending after failure")
	}

	if failed.statusMessage != "Ticket action failed: invalid ticket status" {
		t.Fatalf("statusMessage = %q, want ticket action failure", failed.statusMessage)
	}
}

func TestModelPlansWorkActionCommand(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "work-start-ticket"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if planned.commandPalette.Open {
		t.Fatal("command palette is open after work action planning")
	}

	if !planned.workAction.Pending {
		t.Fatal("work action is not pending")
	}

	if planned.workAction.Request.TicketKey != "CHC-1" {
		t.Fatalf("work action ticket key = %q, want CHC-1", planned.workAction.Request.TicketKey)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Work Action",
		"Start selected ticket",
		"Ticket: CHC-1",
		"enter confirms work action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelCancelsWorkAction(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.workAction.Pending {
		t.Fatal("work action is pending after cancellation")
	}

	if cancelled.statusMessage != "Work action cancelled." {
		t.Fatalf("statusMessage = %q, want work cancellation status", cancelled.statusMessage)
	}
}

func TestModelConfirmsWorkAction(t *testing.T) {
	executor := &fakeWorkActionExecutor{
		result: WorkActionResult{
			Message: "Execution exec-1 for ticket CHC-1 finished with status succeeded.",
			Tickets: []chc.IndexedTicket{
				{Key: "CHC-1", Title: "First", Status: "reviewing"},
			},
			Observability: ObservabilityView{
				LatestExecution: ExecutionView{ID: "exec-1", Status: "succeeded"},
			},
		},
	}
	model := NewModel(Options{
		WorkExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != WorkActionStartTicket || executor.request.TicketKey != "CHC-1" {
		t.Fatalf("executor request = %#v, want start CHC-1", executor.request)
	}

	if confirmed.workAction.Pending {
		t.Fatal("work action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	if len(confirmed.tickets) != 1 || confirmed.tickets[0].Status != "reviewing" {
		t.Fatalf("tickets = %#v, want reloaded reviewing ticket", confirmed.tickets)
	}

	if confirmed.observability.LatestExecution.ID != "exec-1" {
		t.Fatalf("LatestExecution.ID = %q, want exec-1", confirmed.observability.LatestExecution.ID)
	}
}

func TestModelReportsWorkActionError(t *testing.T) {
	executor := &fakeWorkActionExecutor{err: errors.New("runtime validation failed")}
	model := NewModel(Options{
		WorkExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if failed.workAction.Pending {
		t.Fatal("work action is pending after failure")
	}

	if failed.statusMessage != "Work action failed: runtime validation failed" {
		t.Fatalf("statusMessage = %q, want work action failure", failed.statusMessage)
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

func TestModelRendersAgentObservability(t *testing.T) {
	model := NewModel(Options{
		Observability: ObservabilityView{
			DaemonStatus:    "running",
			DaemonPID:       1234,
			DaemonUpdatedAt: "2026-05-22T10:00:00Z",
			DaemonHeartbeat: "2026-05-22T10:00:00Z",
			LatestLoop: LoopRunView{
				ID:             "loop-1",
				Status:         "blocked",
				UpdatedAt:      "2026-05-22T10:01:00Z",
				EligibleCount:  3,
				ProcessedCount: 1,
				BlockedCount:   1,
				FailedCount:    0,
				FailureReason:  "blocked ticket",
			},
			LatestExecution: ExecutionView{
				ID:             "exec-1",
				TicketKey:      "CHC-1",
				Status:         "failed",
				UpdatedAt:      "2026-05-22T10:02:00Z",
				FailureReason:  "validation failed",
				RunDir:         ".chc/runtime/runs/CHC-1/exec-1",
				ResultPath:     ".chc/runtime/runs/CHC-1/exec-1/result.json",
				LogPath:        ".chc/runtime/runs/CHC-1/exec-1/runtime.log",
				ValidationPath: ".chc/runtime/runs/CHC-1/exec-1/validation.log",
			},
		},
	})
	model.selected = 5

	view := model.View()

	for _, want := range []string{
		"Agent",
		"Daemon: running",
		"PID: 1234",
		"Latest Loop",
		"ID: loop-1",
		"Counts: 3 eligible / 1 processed / 1 blocked / 0 failed",
		"Latest Execution",
		"Ticket: CHC-1",
		"Validation: .chc/runtime/runs/CHC-1/exec-1/validation.log",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersRuntimeObservability(t *testing.T) {
	model := NewModel(Options{
		Observability: ObservabilityView{
			RuntimeImage:             "chc-agent-runtime:local",
			RuntimeDockerfile:        ".chc/runtime/agent/Dockerfile",
			RuntimeArtifactDir:       ".chc/runtime/runs",
			RuntimeValidationCommand: "make verify",
			RuntimeChecks: []RuntimeCheckView{
				{Name: "workspace", Status: "ok", Path: ".chc", Message: "ok"},
				{Name: "dockerfile", Status: "fail", Path: ".chc/runtime/agent/Dockerfile", Message: "missing"},
			},
		},
	})
	model.selected = 6

	view := model.View()

	for _, want := range []string{
		"Runtime",
		"Image: chc-agent-runtime:local",
		"Dockerfile: .chc/runtime/agent/Dockerfile",
		"Artifact Dir: .chc/runtime/runs",
		"Validation: make verify",
		"workspace",
		"dockerfile",
		"fail",
		"missing",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestNewObservabilityViewSelectsLatestRecords(t *testing.T) {
	projectDir := t.TempDir()
	older := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)

	view := NewObservabilityView(
		projectDir,
		chc.NewStoppedDaemonState(newer),
		[]chc.ExecutionRecord{
			{ID: "exec-old", Status: chc.ExecutionStatusSucceeded, UpdatedAt: older},
			{ID: "exec-new", Status: chc.ExecutionStatusFailed, UpdatedAt: newer},
		},
		[]chc.LoopRunRecord{
			{ID: "loop-old", Status: chc.LoopRunStatusSucceeded, UpdatedAt: older},
			{ID: "loop-new", Status: chc.LoopRunStatusFailed, UpdatedAt: newer},
		},
		chc.RuntimeValidationResult{
			Plan: chc.RuntimePlan{
				Image:             "image",
				DockerfilePath:    filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
				ArtifactDir:       filepath.Join(projectDir, ".chc", "runtime", "runs"),
				ValidationCommand: "make verify",
			},
			Checks: []chc.RuntimeValidationCheck{
				{Name: "dockerfile", Path: filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"), OK: true, Message: "ok"},
			},
		},
	)

	if view.LatestExecution.ID != "exec-new" {
		t.Fatalf("LatestExecution.ID = %q, want exec-new", view.LatestExecution.ID)
	}

	if view.LatestLoop.ID != "loop-new" {
		t.Fatalf("LatestLoop.ID = %q, want loop-new", view.LatestLoop.ID)
	}

	if view.RuntimeDockerfile != ".chc/runtime/agent/Dockerfile" {
		t.Fatalf("RuntimeDockerfile = %q, want relative runtime path", view.RuntimeDockerfile)
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

func TestModelViewUsesWindowRealEstate(t *testing.T) {
	model := NewModel(Options{
		ProjectDir: "/workspace/project",
		ProjectID:  "project-1",
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	resized, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	view := resized.View()

	lines := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
	if len(lines) != 20 {
		t.Fatalf("View() rendered %d lines, want 20:\n%s", len(lines), view)
	}

	for index, line := range lines {
		if lipgloss.Width(line) != 80 {
			t.Fatalf("line %d width = %d, want 80: %q", index, lipgloss.Width(line), line)
		}
	}

	for _, want := range []string{
		"ChrysoCoord",
		"F1",
		"Project",
		"F10",
		"Quit",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
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

type fakeArtifactActionExecutor struct {
	action ArtifactActionID
	result ArtifactActionResult
	err    error
}

func (e *fakeArtifactActionExecutor) ExecuteArtifactAction(action ArtifactActionID) (ArtifactActionResult, error) {
	e.action = action

	return e.result, e.err
}

type fakeTicketActionExecutor struct {
	request TicketActionRequest
	result  TicketActionResult
	err     error
}

func (e *fakeTicketActionExecutor) ExecuteTicketAction(request TicketActionRequest) (TicketActionResult, error) {
	e.request = request

	return e.result, e.err
}

type fakeWorkActionExecutor struct {
	request WorkActionRequest
	result  WorkActionResult
	err     error
}

func (e *fakeWorkActionExecutor) ExecuteWorkAction(request WorkActionRequest) (WorkActionResult, error) {
	e.request = request

	return e.result, e.err
}

func assertConfirmedArtifactState(t *testing.T, model Model) {
	t.Helper()

	if len(model.tickets) != 1 || model.tickets[0].Key != "CHC-1" {
		t.Fatalf("tickets = %#v, want reloaded CHC-1 ticket", model.tickets)
	}

	if len(model.documents) != 1 || model.documents[0].Title != "TUI Plan" {
		t.Fatalf("documents = %#v, want reloaded TUI Plan document", model.documents)
	}

	if len(model.authoringSessions) != 1 || model.authoringSessions[0].ID != "author-20260522-100000" {
		t.Fatalf("authoringSessions = %#v, want reloaded authoring session", model.authoringSessions)
	}
}
