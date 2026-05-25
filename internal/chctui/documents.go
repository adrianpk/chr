package chctui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DocumentView struct {
	Index        int
	Family       string
	Status       string
	Title        string
	RelativePath string
	UpdatedAt    time.Time
	Preview      string
}

func LoadDocumentViews(ctx context.Context, projectDir string) ([]DocumentView, error) {
	documents := []DocumentView{}

	for _, root := range documentRoots(projectDir) {
		rootDocuments, err := loadDocumentRoot(ctx, projectDir, root.family, root.path)
		if err != nil {
			return nil, err
		}

		documents = append(documents, rootDocuments...)
	}

	sort.SliceStable(documents, func(i, j int) bool {
		if documents[i].Family == documents[j].Family {
			return documents[i].RelativePath < documents[j].RelativePath
		}

		return documents[i].Family < documents[j].Family
	})

	for index := range documents {
		documents[index].Index = index
	}

	return documents, nil
}

func loadDocumentRoot(ctx context.Context, projectDir, family, rootPath string) ([]DocumentView, error) {
	documents := []DocumentView{}

	err := filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, walkErr error) error {
		return appendDocumentView(ctx, projectDir, family, path, entry, walkErr, &documents)
	})
	if errors.Is(err, os.ErrNotExist) {
		return documents, nil
	}

	if err != nil {
		return nil, fmt.Errorf("load %s documents: %w", family, err)
	}

	return documents, nil
}

func appendDocumentView(
	ctx context.Context,
	projectDir string,
	family string,
	path string,
	entry os.DirEntry,
	walkErr error,
	documents *[]DocumentView,
) error {
	if walkErr != nil {
		return walkErr
	}

	err := ctx.Err()
	if err != nil {
		return err
	}

	if entry.IsDir() || !isMarkdownPath(path) {
		return nil
	}

	if isArchivedGeneratedDocumentPath(projectDir, path) {
		return nil
	}

	document, err := documentViewFromPath(projectDir, family, path)
	if err != nil {
		return err
	}

	*documents = append(*documents, document)

	return nil
}

func (m Model) renderDocuments() string {
	if len(m.documents) == 0 {
		return "Documents\n  No project documents were found."
	}

	var output strings.Builder

	output.WriteString("Documents\n")
	output.WriteString("  Type       Status      Title\n")

	for _, document := range m.documents {
		prefix := "  "
		if document.Index == m.selectedDocument {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(renderDocumentRow(document))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedDocumentDetail())

	return strings.TrimRight(output.String(), "\n")
}

//nolint:unused // Kept for legacy/non-shell document family renderers.
func (m Model) renderDocumentsByFamily(title string, families map[string]bool) string {
	documents := m.documentsByFamily(families)
	if len(documents) == 0 {
		return title + "\n  No documents found."
	}

	var output strings.Builder

	output.WriteString(title)
	output.WriteString("\n")
	output.WriteString("  Type       Status      Title\n")

	for _, document := range documents {
		prefix := "  "
		if document.Index == m.selectedDocument {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(renderDocumentRow(document))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedDocumentDetailFor(documents))

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) documentsByFamily(families map[string]bool) []DocumentView {
	documents := []DocumentView{}

	for _, document := range m.documents {
		if families[document.Family] {
			documents = append(documents, document)
		}
	}

	return documents
}

func renderDocumentRow(document DocumentView) string {
	return fmt.Sprintf(
		"%-10s %-11s %s",
		document.Family,
		document.Status,
		document.Title,
	)
}

func (m Model) renderSelectedDocumentDetail() string {
	return m.renderSelectedDocumentDetailFor(m.documents)
}

func (m Model) renderSelectedDocumentDetailFor(documents []DocumentView) string {
	document, ok := m.selectedDocumentViewFor(documents)
	if !ok {
		return "Selected Document\n  No document selected."
	}

	var output strings.Builder

	output.WriteString("Selected Document\n")
	output.WriteString(fmt.Sprintf("  Type: %s\n", document.Family))
	output.WriteString(fmt.Sprintf("  Status: %s\n", document.Status))
	output.WriteString(fmt.Sprintf("  Title: %s\n", document.Title))
	output.WriteString(fmt.Sprintf("  Updated: %s\n", renderTime(document.UpdatedAt)))
	output.WriteString(fmt.Sprintf("  Path: %s\n", document.RelativePath))
	output.WriteString("  Preview:\n")
	output.WriteString(indentText(document.Preview, "    "))

	return strings.TrimRight(output.String(), "\n")
}

//nolint:unused // Kept for legacy/non-shell document view paths.
func (m Model) selectedDocumentView() (DocumentView, bool) {
	return m.selectedDocumentViewFor(m.documents)
}

func (m Model) selectedDocumentViewFor(documents []DocumentView) (DocumentView, bool) {
	for _, document := range documents {
		if document.Index == m.selectedDocument {
			return document, true
		}
	}

	if len(documents) > 0 {
		return documents[0], true
	}

	return DocumentView{}, false
}

func (m *Model) selectPreviousDocument() {
	documents := m.currentDocumentViews()
	if len(documents) == 0 {
		return
	}

	position := m.selectedDocumentPosition(documents)
	if position == 0 {
		m.selectedDocument = documents[len(documents)-1].Index

		return
	}

	m.selectedDocument = documents[position-1].Index
}

func (m *Model) selectNextDocument() {
	documents := m.currentDocumentViews()
	if len(documents) == 0 {
		return
	}

	position := m.selectedDocumentPosition(documents)
	if position >= len(documents)-1 {
		m.selectedDocument = documents[0].Index

		return
	}

	m.selectedDocument = documents[position+1].Index
}

func (m Model) currentDocumentViews() []DocumentView {
	switch m.SelectedView() {
	case ViewDumps:
		return m.documentsByFamily(map[string]bool{"dump": true})
	case ViewQuizzes:
		return m.documentsByFamily(map[string]bool{"quiz": true})
	case ViewSpecs:
		return m.documentsByFamily(map[string]bool{"spec": true})
	default:
		return append([]DocumentView(nil), m.documents...)
	}
}

func (m Model) selectedDocumentPosition(documents []DocumentView) int {
	for index, document := range documents {
		if document.Index == m.selectedDocument {
			return index
		}
	}

	return 0
}

func documentRoots(projectDir string) []struct {
	family string
	path   string
} {
	return []struct {
		family string
		path   string
	}{
		{family: "dump", path: filepath.Join(projectDir, ".chc", "dumps")},
		{family: "dump", path: filepath.Join(projectDir, ".prj", "dumps")},
		{family: "quiz", path: filepath.Join(projectDir, ".prj", "quizzes")},
		{family: "spec", path: filepath.Join(projectDir, ".prj", "specs")},
		{family: "plan", path: filepath.Join(projectDir, ".prj", "plans")},
		{family: "tracker", path: filepath.Join(projectDir, ".prj", "trackers")},
		{family: "report", path: filepath.Join(projectDir, ".prj", "reports")},
		{family: "handoff", path: filepath.Join(projectDir, ".prj", "handovers")},
		{family: "quiz", path: filepath.Join(projectDir, ".chc", "quizzes")},
		{family: "spec", path: filepath.Join(projectDir, ".chc", "specs")},
		{family: "plan", path: filepath.Join(projectDir, ".chc", "plans")},
		{family: "tracker", path: filepath.Join(projectDir, ".chc", "trackers")},
	}
}

func isArchivedGeneratedDocumentPath(projectDir, path string) bool {
	relativePath, err := filepath.Rel(filepath.Join(projectDir, ".chc"), path)
	if err != nil {
		return false
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")

	return len(parts) >= 3 &&
		(parts[0] == "plans" || parts[0] == "specs") &&
		parts[1] == "archive"
}

func documentViewFromPath(projectDir, family, path string) (DocumentView, error) {
	info, err := os.Stat(path)
	if err != nil {
		return DocumentView{}, fmt.Errorf("stat document: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return DocumentView{}, fmt.Errorf("read document: %w", err)
	}

	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return DocumentView{}, fmt.Errorf("relativize document path: %w", err)
	}

	title, status, preview := summarizeMarkdownDocument(string(content), filepath.Base(path))

	return DocumentView{
		Family:       family,
		Status:       status,
		Title:        title,
		RelativePath: filepath.ToSlash(relativePath),
		UpdatedAt:    info.ModTime(),
		Preview:      preview,
	}, nil
}

func summarizeMarkdownDocument(content, fallbackName string) (string, string, string) {
	title := strings.TrimSuffix(fallbackName, filepath.Ext(fallbackName))
	status := ""
	previewLines := []string{}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "# ") && title == strings.TrimSuffix(fallbackName, filepath.Ext(fallbackName)) {
			title = strings.TrimSpace(strings.TrimPrefix(line, "# "))

			continue
		}

		nextStatus, ok := documentStatusLine(line)
		if ok && status == "" {
			status = nextStatus

			continue
		}

		if len(previewLines) < 3 {
			previewLines = append(previewLines, line)
		}
	}

	if status == "" {
		status = "unknown"
	}

	if len(previewLines) == 0 {
		previewLines = append(previewLines, "No preview available.")
	}

	return title, status, strings.Join(previewLines, "\n")
}

func documentStatusLine(line string) (string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok || !strings.EqualFold(strings.TrimSpace(key), "status") {
		return "", false
	}

	return strings.TrimSpace(value), true
}

func isMarkdownPath(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))

	return extension == ".md" || extension == ".markdown"
}

func indentText(value, prefix string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = prefix + line
	}

	return strings.Join(lines, "\n")
}

func renderTime(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}

	return value.UTC().Format(time.RFC3339)
}
