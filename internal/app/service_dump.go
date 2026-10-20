package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// DumpScope narrows a dump operation to an optional use-case scope.
type DumpScope struct {
	CaseSlug string
}

// DumpView is the TUI-facing projection of a workspace dump document.
type DumpView struct {
	Revision     string
	NodeID       string
	UseCaseTitle string
	Title        string
	Status       string
	DocumentType string
	SourcePath   string
	Scope        DumpScope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

// DumpListResult contains the current dump list and a status message.
type DumpListResult struct {
	Message string
	Dumps   []DumpView
}

// DumpCreateRequest describes a new dump capture.
type DumpCreateRequest struct {
	OperationID, ExpectedRevision, Namespace string
	Title                                    string
	Content                                  string
	Scope                                    DumpScope
}

// DumpAppendRequest appends content to an existing dump.
type DumpAppendRequest struct {
	OperationID, ExpectedRevision string
	RelativePath                  string
	Content                       string
}

// DumpSaveRequest replaces the content of an existing dump.
type DumpSaveRequest struct {
	OperationID, ExpectedRevision string
	RelativePath                  string
	Content                       string
}

// DumpOpenRequest selects a dump for editing.
type DumpOpenRequest struct {
	RelativePath string
}

// DumpReadRequest selects a dump for reading.
type DumpReadRequest struct {
	RelativePath string
}

// DumpEditorCommand describes an external editor invocation.
type DumpEditorCommand struct {
	Program     string
	Args        []string
	Dir         string
	Description string
}

// DumpEditorLauncher resolves the editor command for a dump path.
type DumpEditorLauncher interface {
	EditorCommand(ctx context.Context, path string) (DumpEditorCommand, error)
}

// DumpOperationResult is the TUI-facing result for dump operations.
type DumpOperationResult struct {
	Committed, ProjectionStale bool
	Mutation                   core.ProjectDumpMutationResult
	Revision, RelativePath     string
	SourceHash, MetadataHash   string
	Message                    string
	Dumps                      []DumpView
	Content                    string
	EditorCommand              *DumpEditorCommand
}

func (s service) ReadDump(request DumpReadRequest) (DumpOperationResult, error) {
	observed, err := core.ReadDumpDocument(s.ctx, s.projectContext.ProjectDir, request.RelativePath)
	if err != nil {
		return DumpOperationResult{}, err
	}

	return DumpOperationResult{Message: "Editing dump " + observed.Artifact.Reference + ".", Content: observed.Content, Revision: observed.Revision, RelativePath: observed.Artifact.Reference, SourceHash: observed.Artifact.SourceHash, MetadataHash: observed.Artifact.MetadataHash}, nil
}

func (s service) OpenDump(request DumpOpenRequest) (DumpOperationResult, error) {
	path, err := s.resolveDumpPath(request.RelativePath)
	if err != nil {
		return DumpOperationResult{}, err
	}

	command, err := s.dumpEditor.EditorCommand(s.ctx, path)
	if err != nil {
		return DumpOperationResult{}, err
	}

	return DumpOperationResult{
		Message:       "Editing dump " + filepath.ToSlash(request.RelativePath) + ".",
		EditorCommand: &command,
	}, nil
}

func (s service) dumpDirectory() string {
	return core.DefaultOpsArtifactRoot(s.projectContext.ProjectDir, "dump")
}

func (s service) resolveDumpPath(relativePath string) (string, error) {
	observed, err := core.ResolveProjectDumpReference(s.ctx, s.projectContext.ProjectDir, relativePath)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(observed.Reference)), nil
}

func (s service) curateDump(request DumpCreateRequest) (DumpCurationResult, error) {
	provider := s.dumpCurator
	if provider == nil {
		provider = fallbackDumpCurationProvider{Reason: "dump curation provider is not configured"}
	}

	result, err := provider.CurateDump(s.ctx, DumpCurationRequest{
		ProjectDir: s.projectContext.ProjectDir,
		ProjectID:  s.projectContext.ProjectID,
		Title:      request.Title,
		Content:    request.Content,
		Scope:      request.Scope,
	})
	if err == nil {
		return normalizeDumpCurationResult(result, request.Content), nil
	}

	if _, binding, bound := core.HarnessGenerationSelection(s.ctx); bound && binding.Selection.Source == "project" {
		return DumpCurationResult{}, err
	}

	fallback, fallbackErr := (fallbackDumpCurationProvider{
		Reason: "llm_unavailable: " + err.Error(),
	}).CurateDump(s.ctx, DumpCurationRequest{
		ProjectDir: s.projectContext.ProjectDir,
		ProjectID:  s.projectContext.ProjectID,
		Title:      request.Title,
		Content:    request.Content,
		Scope:      request.Scope,
	})

	return fallback, fallbackErr
}

func dumpSlug(value string) string {
	var output strings.Builder

	lastDash := false

	for _, char := range strings.ToLower(value) {
		switch {
		case char >= 'a' && char <= 'z':
			output.WriteRune(char)

			lastDash = false
		case char >= '0' && char <= '9':
			output.WriteRune(char)

			lastDash = false
		case !lastDash:
			output.WriteRune('-')

			lastDash = true
		}
	}

	return strings.Trim(output.String(), "-")
}

func renderCuratedDumpMarkdown(
	title string,
	scope DumpScope,
	rawRelativePath string,
	curation DumpCurationResult,
) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("title: ")
	output.WriteString(quoteDumpFrontmatter(title))
	output.WriteString("\n")
	output.WriteString("status: \"curated\"\n")
	output.WriteString("type: \"dump\"\n")
	output.WriteString("source: ")
	output.WriteString(quoteDumpFrontmatter(rawRelativePath))
	output.WriteString("\n")

	if strings.TrimSpace(scope.CaseSlug) != "" {
		output.WriteString("case: ")
		output.WriteString(quoteDumpFrontmatter(scope.CaseSlug))
		output.WriteString("\n")
	}

	output.WriteString("---\n\n")
	output.WriteString("# ")
	output.WriteString(title)
	output.WriteString("\n\n")
	output.WriteString(strings.TrimSpace(curation.PolishedBody))
	output.WriteString("\n")

	return output.String()
}

func renderRawDumpMarkdown(
	title string,
	scope DumpScope,
	curatedRelativePath string,
	rawContent string,
	curation DumpCurationResult,
) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("title: ")
	output.WriteString(quoteDumpFrontmatter(title + " Raw Capture"))
	output.WriteString("\n")
	output.WriteString("status: \"raw\"\n")
	output.WriteString("type: \"dump_raw\"\n")
	output.WriteString("curated: ")
	output.WriteString(quoteDumpFrontmatter(curatedRelativePath))
	output.WriteString("\n")

	if strings.TrimSpace(scope.CaseSlug) != "" {
		output.WriteString("case: ")
		output.WriteString(quoteDumpFrontmatter(scope.CaseSlug))
		output.WriteString("\n")
	}

	output.WriteString("---\n\n")
	output.WriteString("# ")
	output.WriteString(title)
	output.WriteString(" Raw Capture\n\n")
	output.WriteString("## Highlights\n\n")

	for _, highlight := range curation.Highlights {
		output.WriteString("- ")
		output.WriteString(highlight)
		output.WriteString("\n")
	}

	output.WriteString("\n## Raw Capture\n\n")

	body := strings.TrimSpace(rawContent)
	if body == "" {
		body = "No dump content provided."
	}

	output.WriteString(body)
	output.WriteString("\n")

	if len(curation.Warnings) > 0 {
		output.WriteString("\n## Curation Warnings\n\n")

		for _, warning := range curation.Warnings {
			output.WriteString("- ")
			output.WriteString(warning)
			output.WriteString("\n")
		}
	}

	return output.String()
}

func quoteDumpFrontmatter(value string) string {
	return fmt.Sprintf("%q", strings.TrimSpace(value))
}

type defaultDumpEditorLauncher struct{}

func (defaultDumpEditorLauncher) EditorCommand(_ context.Context, path string) (DumpEditorCommand, error) {
	editor := strings.TrimSpace(os.Getenv("VISUAL"))
	if editor == "" {
		editor = strings.TrimSpace(os.Getenv("EDITOR"))
	}

	if editor == "" {
		editor = "vi"
	}

	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		shell = "/bin/sh"
	}

	dir := filepath.Dir(path)
	file := filepath.Base(path)
	script := "cd " + shellQuote(dir) + " && exec " + editor + " " + shellQuote(file)

	return DumpEditorCommand{
		Program:     shell,
		Args:        []string{"-lc", script},
		Dir:         dir,
		Description: editor + " " + file,
	}, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
