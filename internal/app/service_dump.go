package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Title   string
	Content string
	Scope   DumpScope
}

// DumpAppendRequest appends content to an existing dump.
type DumpAppendRequest struct {
	RelativePath string
	Content      string
}

// DumpSaveRequest replaces the content of an existing dump.
type DumpSaveRequest struct {
	RelativePath string
	Content      string
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
	Message       string
	Dumps         []DumpView
	Content       string
	EditorCommand *DumpEditorCommand
}

func (s service) ListDumps() (DumpListResult, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return DumpListResult{}, err
	}

	return DumpListResult{
		Message: "Dumps loaded.",
		Dumps:   s.dumpViewsFromIndex(index),
	}, nil
}

func (s service) CreateDump(request DumpCreateRequest) (DumpOperationResult, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return DumpOperationResult{}, fmt.Errorf("dump title is required")
	}

	curation := s.curateDump(request)

	path, err := s.createDumpPairFiles(title, request.Scope, request.Content, curation)
	if err != nil {
		return DumpOperationResult{}, err
	}

	result, err := s.ListDumps()
	if err != nil {
		return DumpOperationResult{}, err
	}

	relativePath, err := filepath.Rel(s.projectContext.ProjectDir, path)
	if err != nil {
		relativePath = path
	}

	return DumpOperationResult{
		Message: "Created dump " + filepath.ToSlash(relativePath) + ".",
		Dumps:   result.Dumps,
	}, nil
}

func (s service) AppendDump(request DumpAppendRequest) (DumpOperationResult, error) {
	path, err := s.resolveDumpPath(request.RelativePath)
	if err != nil {
		return DumpOperationResult{}, err
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		content = "Appended dump note from the TUI."
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return DumpOperationResult{}, fmt.Errorf("open dump for append: %w", err)
	}

	_, writeErr := file.WriteString("\n\n" + content + "\n")
	closeErr := file.Close()

	if writeErr != nil {
		return DumpOperationResult{}, fmt.Errorf("append dump: %w", writeErr)
	}

	if closeErr != nil {
		return DumpOperationResult{}, fmt.Errorf("close dump: %w", closeErr)
	}

	result, err := s.ListDumps()
	if err != nil {
		return DumpOperationResult{}, err
	}

	return DumpOperationResult{
		Message: "Appended dump " + filepath.ToSlash(request.RelativePath) + ".",
		Dumps:   result.Dumps,
	}, nil
}

func (s service) SaveDump(request DumpSaveRequest) (DumpOperationResult, error) {
	path, err := s.resolveDumpPath(request.RelativePath)
	if err != nil {
		return DumpOperationResult{}, err
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		return DumpOperationResult{}, fmt.Errorf("dump content is required")
	}

	err = os.WriteFile(path, []byte(content+"\n"), 0o644)
	if err != nil {
		return DumpOperationResult{}, fmt.Errorf("save dump: %w", err)
	}

	result, err := s.ListDumps()
	if err != nil {
		return DumpOperationResult{}, err
	}

	return DumpOperationResult{
		Message: "Saved dump " + filepath.ToSlash(request.RelativePath) + ".",
		Dumps:   result.Dumps,
	}, nil
}

func (s service) ReadDump(request DumpReadRequest) (DumpOperationResult, error) {
	path, err := s.resolveDumpPath(request.RelativePath)
	if err != nil {
		return DumpOperationResult{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return DumpOperationResult{}, fmt.Errorf("read dump: %w", err)
	}

	return DumpOperationResult{
		Message: "Editing dump " + filepath.ToSlash(request.RelativePath) + ".",
		Content: string(content),
	}, nil
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
	return core.DumpDirectory(s.projectContext.ProjectDir)
}

func (s service) resolveDumpPath(relativePath string) (string, error) {
	relativePath = filepath.ToSlash(strings.TrimSpace(relativePath))
	if relativePath == "" {
		return "", fmt.Errorf("dump path is required")
	}

	cleanPath := filepath.Clean(filepath.FromSlash(relativePath))
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || cleanPath == ".." {
		return "", fmt.Errorf("dump path must be project-relative")
	}

	if filepath.ToSlash(filepath.Dir(cleanPath)) != core.UserRelativePath("dumps") {
		return "", fmt.Errorf("dump path must be under %s", core.UserRelativePath("dumps"))
	}

	path := filepath.Join(s.projectContext.ProjectDir, cleanPath)

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat dump: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("dump path is a directory")
	}

	return path, nil
}

type dumpPathPair struct {
	CuratedPath string
	RawPath     string
}

func (s service) curateDump(request DumpCreateRequest) DumpCurationResult {
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
		return normalizeDumpCurationResult(result, request.Content)
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
	if fallbackErr != nil {
		return normalizeDumpCurationResult(DumpCurationResult{}, request.Content)
	}

	return fallback
}

func (s service) relativeDumpPath(path string) string {
	relativePath, err := filepath.Rel(s.projectContext.ProjectDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(relativePath)
}

func candidateDumpPathPairs(directory, title string) []dumpPathPair {
	slug := dumpSlug(title)
	if slug == "" {
		slug = "dump"
	}

	pairs := []dumpPathPair{{
		CuratedPath: filepath.Join(directory, slug+".md"),
		RawPath:     filepath.Join(directory, slug+"-raw.md"),
	}}
	for index := 2; index <= 100; index++ {
		pairs = append(pairs, dumpPathPair{
			CuratedPath: filepath.Join(directory, fmt.Sprintf("%s-%d.md", slug, index)),
			RawPath:     filepath.Join(directory, fmt.Sprintf("%s-%d-raw.md", slug, index)),
		})
	}

	return pairs
}

func dumpPairPathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func writeDumpPair(pair dumpPathPair, curatedContent, rawContent string) error {
	err := writeNewDumpFile(pair.CuratedPath, curatedContent)
	if err != nil {
		return err
	}

	err = writeNewDumpFile(pair.RawPath, rawContent)
	if err != nil {
		removeErr := os.Remove(pair.CuratedPath)
		if removeErr != nil {
			return fmt.Errorf("create raw dump: %w; remove partial curated dump: %w", err, removeErr)
		}

		return err
	}

	return nil
}

func writeNewDumpFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create dump file: %w", err)
	}

	_, writeErr := file.WriteString(content)
	closeErr := file.Close()

	if writeErr != nil {
		return fmt.Errorf("write dump file: %w", writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close dump file: %w", closeErr)
	}

	return nil
}

func (s service) createDumpPairFiles(
	title string,
	scope DumpScope,
	rawContent string,
	curation DumpCurationResult,
) (string, error) {
	err := os.MkdirAll(s.dumpDirectory(), 0o755)
	if err != nil {
		return "", fmt.Errorf("create dump directory: %w", err)
	}

	for _, pair := range candidateDumpPathPairs(s.dumpDirectory(), title) {
		if dumpPairPathExists(pair.CuratedPath) || dumpPairPathExists(pair.RawPath) {
			continue
		}

		curatedRelativePath := s.relativeDumpPath(pair.CuratedPath)
		rawRelativePath := s.relativeDumpPath(pair.RawPath)
		curatedContent := renderCuratedDumpMarkdown(title, scope, rawRelativePath, curation)
		rawCaptureContent := renderRawDumpMarkdown(title, scope, curatedRelativePath, rawContent, curation)

		err = writeDumpPair(pair, curatedContent, rawCaptureContent)
		if err != nil {
			return "", err
		}

		return pair.CuratedPath, nil
	}

	return "", fmt.Errorf("create dump file: exhausted unique names for %q", title)
}

func (s service) dumpViewsFromIndex(index core.WorkspaceIndex) []DumpView {
	nodeIDs := append([]string(nil), index.ByKind[core.WorkspaceNodeDump]...)
	sort.Strings(nodeIDs)

	dumps := make([]DumpView, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		node, ok := index.NodesByID[nodeID]
		if !ok {
			continue
		}

		if core.IsRawDumpNode(node) {
			continue
		}

		dumps = append(dumps, s.dumpViewFromNode(nodeID, node))
	}

	return dumps
}

func (s service) dumpViewFromNode(nodeID string, node core.WorkspaceNode) DumpView {
	return DumpView{
		NodeID:       nodeID,
		UseCaseTitle: s.dumpUseCaseTitle(node.CaseSlug),
		Title:        node.Title,
		Status:       node.Status,
		DocumentType: node.DocumentType,
		SourcePath:   node.SourcePath,
		Scope:        DumpScope{CaseSlug: node.CaseSlug},
		RelativePath: node.RelativePath,
		UpdatedAt:    node.ModifiedAt,
		Content:      s.dumpContent(node.RelativePath),
	}
}

func (s service) dumpUseCaseTitle(caseSlug string) string {
	if strings.TrimSpace(caseSlug) == "" {
		return "Project-wide"
	}

	cases, err := core.NewCaseMetadataService(s.projectContext.ProjectDir).LoadAll()
	if err == nil {
		for _, metadata := range cases {
			if metadata.Slug == caseSlug && strings.TrimSpace(metadata.Name) != "" {
				return metadata.Name
			}
		}
	}

	return caseSlug
}

func (s service) dumpContent(relativePath string) string {
	path := filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(relativePath))

	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return string(content)
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
