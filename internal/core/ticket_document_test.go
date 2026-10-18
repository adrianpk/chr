package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Editing the full description preserves unknown metadata, discussion and identity.
func TestTicketDocumentRoundTrip(t *testing.T) {
	for _, kind := range []string{"canonical", "legacy", "repository"} {
		t.Run(kind, func(t *testing.T) {
			opts, original := ticketEditorFixture(t, kind)
			body := "# Description\n\nFull replacement text with details beyond the synopsis.\n\n# Validation\n\nKeep evidence.\n"
			fields := map[string]string{
				"title": "Edited title", "priority": "high", "severity": "high", "type": "bug",
				"project_area": "ui", "source_kind": "implementation", "owner": "Maintainer", "assignee": "Developer",
				"branch": "feat/ticket", "pr": "https://forge.example.test/team/project/pulls/7", "commits": "123456789abcdef\nabcdef123456789",
			}
			request := TicketEditRequest{Key: original.Ticket.Key, Revision: original.Revision, Fields: fields, Body: &body, Binding: ticketEditorBinding("save")}

			saved, err := UpdateTicketDocument(t.Context(), opts, request)
			if err != nil {
				t.Fatal(err)
			}

			checkTicketEditorSaved(t, original, saved, body, fields)

			before := ticketEditorBytes(t, saved.Ticket.Path)
			if _, err := UpdateTicketDocument(t.Context(), opts, request); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("stale editor overwrote newer content", err)
			}

			if ticketEditorBytes(t, saved.Ticket.Path) != before {
				t.Fatal("stale edit changed the file")
			}

			checkTicketEditorComment(t, opts, saved)
		})
	}
}

func ticketEditorFixture(t *testing.T, kind string) (TicketBoardReadOptions, TicketDocument) {
	t.Helper()

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: t.TempDir(), Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	key, identity := "CHR-7", "key: CHR-7\n"

	path := filepath.Join(layout.Root(), "ticket", "CHR-7.md")
	if kind == "legacy" {
		path = UserPath(layout.ProjectDir(), "tickets", "CHR-7.md")
	}

	if kind == "repository" {
		key, identity = "TKT-20261018000007", "id: TKT-20261018000007\nkind: task\nscope: ui\nsource: chat\n"
		path = filepath.Join(layout.Root(), "ticket", "in_progress", "20261018000007-ticket.md")
	}

	content := "---\n" + identity + "title: Original title\nstatus: in_progress\npriority: normal\nseverity: high\ntype: task\ncreated_at: 2026-01-02T03:04:05Z\nupdated_at: 2026-01-02T03:04:05Z\ncustom:\n  title: Keep nested title\n  flag: true\n---\n\n# Description\n\nOriginal complete description.\n\n# Comments\n\n## 2026-01-02T04:00:00Z Reviewer\n\nExisting comment.\n\n# Validation\n\nKeep evidence.\n"
	writeTicketFile(t, path, content)

	opts := TicketBoardReadOptions{Layout: layout, ProjectID: "project-1"}

	document, err := ReadTicketDocument(t.Context(), opts, key)
	if err != nil || document.Fields["title"] != "Original title" {
		t.Fatal("document read changed its identity or nested metadata shadowed its title", err, document)
	}

	return opts, document
}

func checkTicketEditorSaved(t *testing.T, original, saved TicketDocument, body string, fields map[string]string) {
	t.Helper()

	if saved.Ticket.Path != original.Ticket.Path || saved.Description != strings.TrimSpace(body) || saved.Revision == original.Revision {
		t.Fatal("save lost document identity, full body or revision", saved)
	}

	for name, value := range fields {
		if saved.Fields[name] != value {
			t.Fatalf("field %s = %q; want %q", name, saved.Fields[name], value)
		}
	}

	content := ticketEditorBytes(t, saved.Ticket.Path)
	for _, retained := range []string{"custom:\n  title: Keep nested title\n  flag: true", "Existing comment.", "Keep evidence."} {
		if !strings.Contains(content, retained) {
			t.Fatal("save lost retained data", retained)
		}
	}
}

func checkTicketEditorComment(t *testing.T, opts TicketBoardReadOptions, saved TicketDocument) {
	t.Helper()

	commented, err := UpdateTicketDocument(t.Context(), opts, TicketEditRequest{
		Key: saved.Ticket.Key, Revision: saved.Revision, Comment: "Independent comment.", Binding: ticketEditorBinding("comment"),
	})
	if err != nil || commented.Description != saved.Description || commented.Fields["title"] != saved.Fields["title"] || !strings.Contains(commented.Discussion, "Independent comment.") || !strings.Contains(commented.Discussion, "Existing comment.") {
		t.Fatal("comment changed metadata, description or existing discussion", commented, err)
	}
}

// The owner rejects invalid fields and transitions before replacing any bytes.
func TestTicketDocumentRejections(t *testing.T) {
	opts, original := ticketEditorFixture(t, "canonical")
	for _, fields := range []map[string]string{
		{"priority": "urgent"},
		{"type": "imaginary"},
		{"status": TicketStatusDone},
		{"claim_generation": "99"},
		{"pr": "file:///secret"},
		{"commits": "not-a-commit"},
	} {
		before := ticketEditorBytes(t, original.Ticket.Path)

		_, err := UpdateTicketDocument(t.Context(), opts, TicketEditRequest{Key: original.Ticket.Key, Revision: original.Revision, Fields: fields, Binding: ticketEditorBinding("rejected")})
		if err == nil || before != ticketEditorBytes(t, original.Ticket.Path) {
			t.Fatal("invalid mutation changed bytes", fields, err)
		}
	}
}

// Competing non-Git saves compare the same observation after process-safe exclusion.
func TestTicketDocumentConcurrent(t *testing.T) {
	opts, original := ticketEditorFixture(t, "canonical")
	start := make(chan struct{})
	results := make(chan error, 2)

	for i := range 2 {
		go func() {
			<-start

			body := fmt.Sprintf("Concurrent body %d.", i)

			_, err := UpdateTicketDocument(t.Context(), opts, TicketEditRequest{Key: original.Ticket.Key, Revision: original.Revision, Body: &body, Binding: ticketEditorBinding(fmt.Sprintf("concurrent-%d", i))})

			results <- err
		}()
	}

	close(start)

	success, conflict := 0, 0

	for range 2 {
		err := <-results
		if err == nil {
			success++
		} else if errors.Is(err, ErrWorkflowConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}

	if success != 1 || conflict != 1 {
		t.Fatal("competing saves did not preserve one winner", success, conflict)
	}
}

func ticketEditorBinding(id string) WorkflowMutationBinding {
	return WorkflowMutationBinding{RequestID: id, Operator: "editor"}
}

func ticketEditorBytes(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
