package chcapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type DumpScope struct {
	CaseSlug string
}

type DumpView struct {
	NodeID       string
	Title        string
	Status       string
	Scope        DumpScope
	RelativePath string
	UpdatedAt    time.Time
}

type DumpListResult struct {
	Message string
	Dumps   []DumpView
}

type DumpCreateRequest struct {
	Title   string
	Content string
	Scope   DumpScope
}

type DumpAppendRequest struct {
	RelativePath string
	Content      string
}

type DumpSaveRequest struct {
	RelativePath string
	Content      string
}

type DumpOpenRequest struct {
	RelativePath string
}

type DumpReadRequest struct {
	RelativePath string
}

type DumpEditorCommand struct {
	Program     string
	Args        []string
	Dir         string
	Description string
}

type DumpEditorLauncher interface {
	EditorCommand(ctx context.Context, path string) (DumpEditorCommand, error)
}

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
		Dumps:   dumpViewsFromIndex(index),
	}, nil
}

func (s service) CreateDump(request DumpCreateRequest) (DumpOperationResult, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return DumpOperationResult{}, fmt.Errorf("dump title is required")
	}

	content := renderDumpMarkdown(title, request.Scope, request.Content)

	path, err := s.createDumpFile(title, content)
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
	return chc.DumpDirectory(s.projectContext.ProjectDir)
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

	if filepath.ToSlash(filepath.Dir(cleanPath)) != ".chc/dumps" {
		return "", fmt.Errorf("dump path must be under .chc/dumps")
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

func (s service) createDumpFile(title, content string) (string, error) {
	err := os.MkdirAll(s.dumpDirectory(), 0o755)
	if err != nil {
		return "", fmt.Errorf("create dump directory: %w", err)
	}

	for _, path := range candidateDumpPaths(s.dumpDirectory(), title) {
		file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if os.IsExist(openErr) {
			continue
		}

		if openErr != nil {
			return "", fmt.Errorf("create dump file: %w", openErr)
		}

		_, writeErr := file.WriteString(content)
		closeErr := file.Close()

		if writeErr != nil {
			return "", fmt.Errorf("write dump file: %w", writeErr)
		}

		if closeErr != nil {
			return "", fmt.Errorf("close dump file: %w", closeErr)
		}

		return path, nil
	}

	return "", fmt.Errorf("create dump file: exhausted unique names for %q", title)
}

func dumpViewsFromIndex(index chc.WorkspaceIndex) []DumpView {
	nodeIDs := append([]string(nil), index.ByKind[chc.WorkspaceNodeDump]...)
	sort.Strings(nodeIDs)

	dumps := make([]DumpView, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		node, ok := index.NodesByID[nodeID]
		if !ok {
			continue
		}

		dumps = append(dumps, dumpViewFromNode(nodeID, node))
	}

	return dumps
}

func dumpViewFromNode(nodeID string, node chc.WorkspaceNode) DumpView {
	return DumpView{
		NodeID:       nodeID,
		Title:        node.Title,
		Status:       node.Status,
		Scope:        DumpScope{CaseSlug: node.CaseSlug},
		RelativePath: node.RelativePath,
		UpdatedAt:    node.ModifiedAt,
	}
}

func candidateDumpPaths(directory, title string) []string {
	slug := dumpSlug(title)
	if slug == "" {
		slug = "dump"
	}

	paths := []string{filepath.Join(directory, slug+".md")}
	for index := 2; index <= 100; index++ {
		paths = append(paths, filepath.Join(directory, fmt.Sprintf("%s-%d.md", slug, index)))
	}

	return paths
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

func renderDumpMarkdown(title string, scope DumpScope, content string) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("title: ")
	output.WriteString(quoteDumpFrontmatter(title))
	output.WriteString("\n")
	output.WriteString("status: \"raw\"\n")

	if strings.TrimSpace(scope.CaseSlug) != "" {
		output.WriteString("case: ")
		output.WriteString(quoteDumpFrontmatter(scope.CaseSlug))
		output.WriteString("\n")
	}

	output.WriteString("---\n\n")
	output.WriteString("# ")
	output.WriteString(title)
	output.WriteString("\n\n")

	body := strings.TrimSpace(content)
	if body == "" {
		body = "Initial dump created from the TUI."
	}

	output.WriteString(body)
	output.WriteString("\n")

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
