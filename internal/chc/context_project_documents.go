package chc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var projectDocumentReferencePattern = regexp.MustCompile(`(?:^|[\s("'])((?:\.prj|docs|\.chc)/(?:[A-Za-z0-9._@=+,:;~/-]+)\.(?:md|markdown|txt|yml|yaml))(?:$|[\s)"',.])`)

type ProjectDocumentContextOptions struct {
	ProjectID  string
	ProjectDir string
	TicketPath string
	Plan       RuntimePlan
	IndexedAt  time.Time
}

type ProjectDocumentContext struct {
	ProjectDir        string
	WorkspaceDir      string
	Ticket            IndexedTicket
	TicketDocument    ContextDocument
	ReferencedDocs    []ContextDocument
	ValidationCommand string
	Sources           []ContextSourceInventoryEntry
}

func LoadProjectDocumentContext(opts ProjectDocumentContextOptions) (ProjectDocumentContext, error) {
	projectDir, ticketPath, err := normalizeProjectDocumentContextOptions(opts)
	if err != nil {
		return ProjectDocumentContext{}, err
	}

	ticket, ticketDocument, err := loadContextTicket(opts, projectDir, ticketPath)
	if err != nil {
		return ProjectDocumentContext{}, err
	}

	referencedDocs, documentSources, err := loadReferencedProjectDocuments(projectDir, ticketDocument.Content)
	if err != nil {
		return ProjectDocumentContext{}, err
	}

	sources := projectDocumentSources(projectDir, ticketDocument, documentSources)

	return ProjectDocumentContext{
		ProjectDir:        projectDir,
		WorkspaceDir:      filepath.Join(projectDir, DirectoryName),
		Ticket:            ticket,
		TicketDocument:    ticketDocument,
		ReferencedDocs:    referencedDocs,
		ValidationCommand: projectValidationCommand(opts.Plan),
		Sources:           sources,
	}, nil
}

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

func normalizeProjectDocumentContextOptions(opts ProjectDocumentContextOptions) (string, string, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return "", "", err
	}

	ticketPath := strings.TrimSpace(opts.TicketPath)
	if ticketPath == "" {
		return "", "", fmt.Errorf("project document ticket path is required")
	}

	if !filepath.IsAbs(ticketPath) {
		ticketPath = filepath.Join(projectDir, ticketPath)
	}

	ticketPath, err = filepath.Abs(ticketPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve ticket path: %w", err)
	}

	ticketPath = filepath.Clean(ticketPath)
	if !pathIsInsideProject(projectDir, ticketPath) {
		return "", "", fmt.Errorf("ticket path is outside project: %s", ticketPath)
	}

	return projectDir, ticketPath, nil
}

func loadContextTicket(
	opts ProjectDocumentContextOptions,
	projectDir string,
	ticketPath string,
) (IndexedTicket, ContextDocument, error) {
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
		return IndexedTicket{}, ContextDocument{}, err
	}

	document, err := readContextDocument(projectDir, "ticket:"+ticket.RelativePath, ticketPath)
	if err != nil {
		return IndexedTicket{}, ContextDocument{}, err
	}

	return ticket, document, nil
}

func loadReferencedProjectDocuments(
	projectDir string,
	content string,
) ([]ContextDocument, []ContextSourceInventoryEntry, error) {
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

			return nil, nil, err
		}

		documents = append(documents, document)
		sources = append(sources, includedDocumentSource(id, ref))
	}

	return documents, sources, nil
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

		if _, ok := seen[ref]; ok {
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
