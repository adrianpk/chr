package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

var projectDocumentReferencePattern = regexp.MustCompile(`(?:^|[\s("'])((?:\.prj|docs|\.chr)/(?:[A-Za-z0-9._@=+,:;~/-]+)\.(?:md|markdown|txt|yml|yaml))(?:$|[\s)"',.])`)

// ProjectDocumentContextOptions describes a package boundary value.
type ProjectDocumentContextOptions struct {
	ProjectID  string
	ProjectDir string
	TicketPath string
	Plan       RuntimePlan
	IndexedAt  time.Time
}

// ProjectDocumentContext describes a package boundary value.
type ProjectDocumentContext struct {
	ProjectDir        string
	WorkspaceDir      string
	Ticket            IndexedTicket
	TicketDocument    ContextDocument
	ReferencedDocs    []ContextDocument
	ValidationCommand string
	Sources           []ContextSourceInventoryEntry
}

type projectDocumentPaths struct {
	ProjectDir string
	TicketPath string
}

type contextTicketLoad struct {
	Ticket   IndexedTicket
	Document ContextDocument
}

type contextDocumentLoad struct {
	Documents []ContextDocument
	Sources   []ContextSourceInventoryEntry
}

// LoadProjectDocumentContext loads the corresponding package data.
func LoadProjectDocumentContext(opts ProjectDocumentContextOptions) (ProjectDocumentContext, error) {
	paths, err := normalizeProjectDocumentContextOptions(opts)
	if err != nil {
		return ProjectDocumentContext{}, err
	}

	ticket, err := loadContextTicket(opts, paths.ProjectDir, paths.TicketPath)
	if err != nil {
		return ProjectDocumentContext{}, err
	}

	referenced, err := loadReferencedProjectDocuments(paths.ProjectDir, ticket.Document.Content)
	if err != nil {
		return ProjectDocumentContext{}, err
	}

	sources := projectDocumentSources(paths.ProjectDir, ticket.Document, referenced.Sources)

	return ProjectDocumentContext{
		ProjectDir:        paths.ProjectDir,
		WorkspaceDir:      filepath.Join(paths.ProjectDir, workspacecore.DirectoryName),
		Ticket:            ticket.Ticket,
		TicketDocument:    ticket.Document,
		ReferencedDocs:    referenced.Documents,
		ValidationCommand: projectValidationCommand(opts.Plan),
		Sources:           sources,
	}, nil
}

// RenderProjectContext renders the corresponding package data.
func RenderProjectContext(ctx ProjectDocumentContext) string {
	var output strings.Builder

	output.WriteString("# Project Context\n\n")
	writeMarkdownField(&output, "Project Directory", ctx.ProjectDir)
	writeMarkdownField(&output, "Workspace Directory", ctx.WorkspaceDir)
	writeMarkdownField(&output, "Ticket", ctx.Ticket.Key+" - "+ctx.Ticket.Title)
	writeMarkdownField(&output, "Ticket Path", ctx.TicketDocument.RelativePath)
	writeMarkdownList(&output, "Validation Commands", []string{ctx.ValidationCommand})

	return output.String()
}

// RenderDocumentsContext renders the corresponding package data.
func RenderDocumentsContext(ctx ProjectDocumentContext) string {
	var output strings.Builder

	output.WriteString("# Document Context\n\n")
	output.WriteString("## Ticket\n\n")
	writeContextDocuments(&output, []ContextDocument{ctx.TicketDocument}, "Ticket document was not loaded.")

	output.WriteString("## Referenced Documents\n\n")
	writeContextDocuments(&output, ctx.ReferencedDocs, "No referenced project documents were found.")

	output.WriteString("## Conversation History\n\n")
	output.WriteString("Implicit conversation history is not included. Relevant prior discussion must be captured in project documents before execution.\n")

	return output.String()
}

// RenderSourceInventory renders the corresponding package data.
func RenderSourceInventory(sources []ContextSourceInventoryEntry) string {
	var output strings.Builder

	output.WriteString("# Context Source Inventory\n\n")
	output.WriteString("| ID | Kind | Class | Path | Included | Required | Reason |\n")
	output.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")

	for _, source := range sources {
		output.WriteString("| ")
		output.WriteString(markdownTableCell(source.ID))
		output.WriteString(" | ")
		output.WriteString(markdownTableCell(string(source.Kind)))
		output.WriteString(" | ")
		output.WriteString(markdownTableCell(string(source.Class)))
		output.WriteString(" | ")
		output.WriteString(markdownTableCell(source.Path))
		output.WriteString(" | ")
		output.WriteString(markdownBool(source.Included))
		output.WriteString(" | ")
		output.WriteString(markdownBool(source.Required))
		output.WriteString(" | ")
		output.WriteString(markdownTableCell(source.Reason))
		output.WriteString(" |\n")
	}

	return output.String()
}

func normalizeProjectDocumentContextOptions(opts ProjectDocumentContextOptions) (projectDocumentPaths, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return projectDocumentPaths{}, err
	}

	ticketPath := strings.TrimSpace(opts.TicketPath)
	if ticketPath == "" {
		return projectDocumentPaths{}, fmt.Errorf("project document ticket path is required")
	}

	if !filepath.IsAbs(ticketPath) {
		ticketPath = filepath.Join(projectDir, ticketPath)
	}

	ticketPath, err = filepath.Abs(ticketPath)
	if err != nil {
		return projectDocumentPaths{}, fmt.Errorf("resolve ticket path: %w", err)
	}

	ticketPath = filepath.Clean(ticketPath)
	if !projectcore.ContainsPath(projectDir, ticketPath) {
		return projectDocumentPaths{}, fmt.Errorf("ticket path is outside project: %s", ticketPath)
	}

	return projectDocumentPaths{ProjectDir: projectDir, TicketPath: ticketPath}, nil
}

func loadContextTicket(
	opts ProjectDocumentContextOptions,
	projectDir string,
	ticketPath string,
) (contextTicketLoad, error) {
	indexedAt := opts.IndexedAt
	if indexedAt.IsZero() {
		indexedAt = time.Now().UTC()
	}

	ticket, err := NewTicketParser().ParseTicket(context.Background(), ParseTicketOptions{
		ProjectID:  opts.ProjectID,
		ProjectDir: projectDir,
		Path:       ticketPath,
		IndexedAt:  indexedAt,
	})
	if err != nil {
		return contextTicketLoad{}, err
	}

	document, err := readContextDocument(projectDir, "ticket:"+ticket.RelativePath, ticketPath)
	if err != nil {
		return contextTicketLoad{}, err
	}

	return contextTicketLoad{Ticket: ticket, Document: document}, nil
}

func loadReferencedProjectDocuments(
	projectDir string,
	content string,
) (contextDocumentLoad, error) {
	refs := referencedProjectDocumentPaths(content)
	documents := make([]ContextDocument, 0, len(refs))
	sources := make([]ContextSourceInventoryEntry, 0, len(refs))

	for _, ref := range refs {
		path := filepath.Join(projectDir, filepath.FromSlash(ref))
		id := "document:" + ref

		document, err := readContextDocument(projectDir, id, path)
		if err != nil {
			if os.IsNotExist(err) {
				sources = append(sources, missingDocumentSource(id, ref))

				continue
			}

			return contextDocumentLoad{}, err
		}

		documents = append(documents, document)
		sources = append(sources, includedDocumentSource(id, ref))
	}

	return contextDocumentLoad{Documents: documents, Sources: sources}, nil
}

func referencedProjectDocumentPaths(content string) []string {
	matches := projectDocumentReferencePattern.FindAllStringSubmatch(content, -1)
	seen := map[string]struct{}{}
	refs := make([]string, 0, len(matches))

	for _, match := range matches {
		ref := filepath.ToSlash(filepath.Clean(match[1]))
		if ref == "." || strings.HasPrefix(ref, "../") {
			continue
		}

		_, ok := seen[ref]
		if ok {
			continue
		}

		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	sort.Strings(refs)

	return refs
}

func readContextDocument(projectDir, id, path string) (ContextDocument, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ContextDocument{}, err
	}

	relPath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return ContextDocument{}, fmt.Errorf("resolve context document relative path: %w", err)
	}

	return ContextDocument{
		ID:           id,
		Path:         path,
		RelativePath: filepath.ToSlash(relPath),
		Content:      string(content),
	}, nil
}

func projectDocumentSources(
	projectDir string,
	ticketDocument ContextDocument,
	documentSources []ContextSourceInventoryEntry,
) []ContextSourceInventoryEntry {
	sources := []ContextSourceInventoryEntry{
		{
			ID:          "project:" + filepath.Base(projectDir),
			Kind:        ContextSourceKindProject,
			Class:       ContextSourceClassInvariant,
			Path:        ".",
			Description: "Project identity and validation context.",
			Required:    true,
			Included:    true,
		},
		{
			ID:          ticketDocument.ID,
			Kind:        ContextSourceKindDocument,
			Class:       ContextSourceClassFresh,
			Path:        ticketDocument.RelativePath,
			Description: "Selected ticket document.",
			Required:    true,
			Included:    true,
		},
	}

	return append(sources, documentSources...)
}

func includedDocumentSource(id, path string) ContextSourceInventoryEntry {
	return ContextSourceInventoryEntry{
		ID:          id,
		Kind:        ContextSourceKindDocument,
		Class:       ContextSourceClassFresh,
		Path:        path,
		Description: "Ticket-referenced project document.",
		Required:    false,
		Included:    true,
	}
}

func missingDocumentSource(id, path string) ContextSourceInventoryEntry {
	return ContextSourceInventoryEntry{
		ID:          id,
		Kind:        ContextSourceKindDocument,
		Class:       ContextSourceClassFresh,
		Path:        path,
		Description: "Ticket-referenced project document.",
		Required:    false,
		Included:    false,
		Reason:      "file was not found",
	}
}

func projectValidationCommand(plan RuntimePlan) string {
	if strings.TrimSpace(plan.ValidationCommand) != "" {
		return plan.ValidationCommand
	}

	return DefaultRuntimeValidationCommand
}

func markdownBool(value bool) string {
	if value {
		return "yes"
	}

	return "no"
}

func markdownTableCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")

	return strings.TrimSpace(value)
}
