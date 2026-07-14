package core

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadProjectDocumentContextLoadsTicketReferences(t *testing.T) {
	dir := t.TempDir()
	ticketPath := filepath.Join(dir, ".chr", "default", "tickets", "CHR-1-example.md")
	specPath := filepath.Join(dir, ".prj", "specs", "specs", "feature.md")
	docsPath := filepath.Join(dir, "docs", "reference", "feature.md")

	writeTestFile(t, specPath, "# Feature Spec\n")
	writeTestFile(t, docsPath, "# Feature Reference\n")
	writeTestFile(t, ticketPath, renderProjectDocumentTestTicket(t, Ticket{
		Key:         "CHR-1",
		Title:       "Example",
		Status:      "open",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "feature",
		UseCase:     "context",
		ProjectArea: "runtime",
	}, `
# Description

Use .prj/specs/specs/feature.md and docs/reference/feature.md.
`))

	ctx, err := LoadProjectDocumentContext(ProjectDocumentContextOptions{
		ProjectID:  "project-1",
		ProjectDir: dir,
		TicketPath: ticketPath,
		Plan: RuntimePlan{
			ValidationCommand: "go test ./...",
		},
		IndexedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("LoadProjectDocumentContext returned error: %v", err)
	}

	if len(ctx.ReferencedDocs) != 2 {
		t.Fatalf("ReferencedDocs = %d, want 2", len(ctx.ReferencedDocs))
	}

	rendered := RenderDocumentsContext(ctx)
	for _, want := range []string{"# Feature Spec", "# Feature Reference", "Implicit conversation history is not included"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered documents missing %q:\n%s", want, rendered)
		}
	}

	project := RenderProjectContext(ctx)
	if !strings.Contains(project, "go test ./...") {
		t.Fatalf("project context missing validation command:\n%s", project)
	}
}

func TestLoadProjectDocumentContextRecordsMissingReferences(t *testing.T) {
	dir := t.TempDir()
	ticketPath := filepath.Join(dir, ".chr", "default", "tickets", "CHR-1-example.md")

	writeTestFile(t, ticketPath, renderProjectDocumentTestTicket(t, Ticket{
		Key:       "CHR-1",
		Title:     "Example",
		Status:    "open",
		Priority:  "normal",
		Severity:  "normal",
		Type:      "feature",
		CreatedAt: time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 20, 11, 30, 0, 0, time.UTC),
	}, `
# Description

Missing reference: .prj/drafts/missing.md.
`))

	ctx, err := LoadProjectDocumentContext(ProjectDocumentContextOptions{
		ProjectDir: dir,
		TicketPath: ticketPath,
		IndexedAt:  time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("LoadProjectDocumentContext returned error: %v", err)
	}

	var found bool

	for _, source := range ctx.Sources {
		if source.Path == ".prj/drafts/missing.md" {
			found = true

			if source.Included {
				t.Fatal("missing reference source was marked included")
			}

			if source.Reason != "file was not found" {
				t.Fatalf("missing reference reason = %q", source.Reason)
			}
		}
	}

	if !found {
		t.Fatal("missing reference source was not recorded")
	}
}

func TestRenderSourceInventoryIncludesDocumentSources(t *testing.T) {
	rendered := RenderSourceInventory([]ContextSourceInventoryEntry{
		{
			ID:       "document:docs/reference/example.md",
			Kind:     ContextSourceKindDocument,
			Class:    ContextSourceClassFresh,
			Path:     "docs/reference/example.md",
			Included: true,
			Required: false,
		},
	})

	for _, want := range []string{"document:docs/reference/example.md", "document", "fresh", "yes"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("source inventory missing %q:\n%s", want, rendered)
		}
	}
}

func renderProjectDocumentTestTicket(t *testing.T, ticket Ticket, body string) string {
	t.Helper()

	if ticket.CreatedAt.IsZero() {
		ticket.CreatedAt = time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	}

	if ticket.UpdatedAt.IsZero() {
		ticket.UpdatedAt = time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	}

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	return content + "\n" + body
}
