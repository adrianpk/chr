//nolint:wsl_v5 // Artifact normalization keeps closely related transformation steps together.
package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
	"forge.adrianpk.com/chr/internal/webui"
)

const (
	webArtifactFamilyCases   = "cases"
	webArtifactFamilyDumps   = "dumps"
	webArtifactFamilyQuizzes = "quizzes"
	webArtifactFamilySpecs   = "specs"
	webArtifactFamilyPlans   = "plans"
)

func (p webViewProvider) CatalogView(ctx context.Context, family string) (webui.ArtifactCatalogView, error) {
	operations := p.service(ctx)

	switch family {
	case webArtifactFamilyCases:
		return p.caseCatalog()
	case webArtifactFamilyDumps:
		return dumpCatalog(operations)
	case webArtifactFamilyQuizzes:
		return quizCatalog(operations)
	case webArtifactFamilySpecs:
		return specCatalog(operations)
	case webArtifactFamilyPlans:
		return planCatalog(operations)
	default:
		return webui.ArtifactCatalogView{}, fmt.Errorf("unknown artifact family %q", family)
	}
}

func (p webViewProvider) caseCatalog() (webui.ArtifactCatalogView, error) {
	cases, err := core.NewCaseMetadataService(p.projectContext.ProjectDir).LoadAll()
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	return webArtifactCatalog(webArtifactFamilyCases, webCaseCards(p.projectContext.ProjectDir, cases)), nil
}

func dumpCatalog(operations service) (webui.ArtifactCatalogView, error) {
	dumps, err := loadTUIDumps(operations)
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	cases, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadBounded(operations.ctx, core.WorkflowRecordLimit)
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	view := webArtifactCatalog(webArtifactFamilyDumps, webDumpCards(dumps))
	for _, item := range cases {
		view.Cases = append(view.Cases, webui.TicketCaseView{Slug: item.Slug, Name: item.Name})
	}
	view.Revision = webStableRevision(view)

	return view, nil
}

func quizCatalog(operations service) (webui.ArtifactCatalogView, error) {
	catalog, err := operations.ReadQuizCatalog()
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}
	view := webArtifactCatalog(webArtifactFamilyQuizzes, webQuizCards(catalog.Items))
	for _, pending := range catalog.Pending {
		view.PendingQuizzes = append(view.PendingQuizzes, webui.QuizScopeChoice{Title: pending.Title, URL: quizFocusURL(pending.Reference)})
	}
	view.Revision = webStableRevision(view)

	return view, nil
}

func specCatalog(operations service) (webui.ArtifactCatalogView, error) {
	documents, err := operations.ListSpecDocuments()
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}
	cards := make([]webui.ArtifactCardView, 0, len(documents))
	for _, document := range documents {
		preview := []rune(document.Body)
		if len(preview) > 1024 {
			preview = preview[:1024]
		}
		card := webui.ArtifactCardView{Family: "specs", ID: document.Reference, Title: document.Title, Preview: string(preview), Revision: document.Revision}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	return webArtifactCatalog("specs", cards), nil
}

func (p webViewProvider) ArtifactDetailView(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	switch request.Family {
	case webArtifactFamilySpecs:
		return p.specArtifactDetail(ctx, request)
	case webArtifactFamilyQuizzes:
		return p.quizArtifactDetail(ctx, request)
	case webArtifactFamilyDumps:
		return p.dumpArtifactDetail(ctx, request)
	case webArtifactFamilyPlans:
		return p.planArtifactDetail(ctx, request)
	default:
		return p.readArtifactDetail(ctx, request)
	}
}

func (p webViewProvider) readArtifactDetail(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	catalog, err := p.CatalogView(ctx, request.Family)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	card, ok := findWebArtifactCard(catalog.Items, request.ID)
	if !ok {
		view := webui.ArtifactDetailView{
			ID:      request.ID,
			Family:  request.Family,
			Title:   request.ID,
			Path:    request.ID,
			Missing: true,
		}
		view.Revision = webStableRevision(view)

		return view, nil
	}

	if request.Family == webArtifactFamilyCases {
		return p.caseDetailView(ctx, card)
	}

	content := card.Preview
	revision := card.Revision

	if card.Path != "" {
		readContent, err := p.readProjectArtifact(card.Path)
		if err == nil {
			content = readContent
			revision = webContentRevision(readContent)
		} else if !errors.Is(err, os.ErrNotExist) {
			return webui.ArtifactDetailView{}, err
		}
	}

	view := webui.ArtifactDetailView{
		Family:     card.Family,
		ID:         card.ID,
		Title:      card.Title,
		Status:     card.Status,
		Scope:      card.Scope,
		Path:       card.Path,
		SourcePath: card.SourcePath,
		UpdatedAt:  card.UpdatedAt,
		Content:    content,
		Revision:   revision,
	}
	if view.Revision == "" {
		view.Revision = webStableRevision(view)
	}

	return view, nil
}

func (p webViewProvider) caseDetailView(ctx context.Context, card webui.ArtifactCardView) (webui.ArtifactDetailView, error) {
	metadata, err := core.NewCaseMetadataService(p.projectContext.ProjectDir).ReadBounded(ctx, card.ID)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	content := metadata.Description
	if strings.TrimSpace(metadata.Context) != "" {
		content = content + "\n\n" + metadata.Context
	}

	view := webui.ArtifactDetailView{
		Plan:            p.casePlanningView(ctx, card.ID, metadata.Name),
		Family:          card.Family,
		ID:              card.ID,
		Title:           metadata.Name,
		WorkflowEnabled: p.workflowEnabled(),
		Description:     metadata.Description,
		Context:         metadata.Context,
		Status:          card.Status,
		Scope:           card.Scope,
		Path:            card.Path,
		SourcePath:      card.SourcePath,
		UpdatedAt:       card.UpdatedAt,
		Content:         content,
		Revision:        metadata.SourceHash,
	}
	if view.Revision == "" {
		view.Revision = webStableRevision(view)
	}

	return view, nil
}

func webArtifactCatalog(family string, items []webui.ArtifactCardView) webui.ArtifactCatalogView {
	view := webui.ArtifactCatalogView{
		Family:       family,
		Title:        webui.CatalogTitle(family),
		EmptyMessage: "No " + family + " found.",
		Items:        items,
	}
	view.Revision = webStableRevision(view)

	return view
}

func webCaseCards(projectDir string, cases []core.CaseMetadata) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(cases))
	for _, metadata := range cases {
		card := webui.ArtifactCardView{
			Family:       webArtifactFamilyCases,
			ID:           metadata.Slug,
			Title:        metadata.Name,
			Status:       "available",
			Scope:        metadata.Slug,
			Path:         caseRelativePath(projectDir, metadata),
			SourcePath:   metadata.SourcePath,
			Preview:      firstNonEmptyString(metadata.Description, metadata.Context),
			ContentState: webContentState(metadata.Exists),
			Revision:     metadata.SourceHash,
		}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	return cards
}

func webDumpCards(dumps []tui.DumpView) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(dumps))
	for _, dump := range dumps {
		displayTitle := webDumpDisplayTitle(dump.Title)
		scope := webScopeLabel(dump.Scope.CaseSlug, dump.UseCaseTitle)
		card := webui.ArtifactCardView{
			Family:       webArtifactFamilyDumps,
			ID:           dump.RelativePath,
			Title:        dump.Title,
			DisplayTitle: displayTitle,
			Status:       dump.Status,
			Scope:        scope,
			Meta:         webDumpMeta(displayTitle, scope, dump.Status),
			Path:         dump.RelativePath,
			SourcePath:   dump.SourcePath,
			UpdatedAt:    webArtifactTime(dump.UpdatedAt),
			Preview:      webDumpPreview(dump.Title, dump.Preview),
			ContentState: webContentState(dump.RelativePath != ""),
			Revision:     dump.Revision,
		}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	return cards
}

func webDumpDisplayTitle(title string) string {
	title = strings.TrimSpace(title)
	for _, suffix := range []string{" Context", " context"} {
		if strings.HasSuffix(title, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(title, suffix))
		}
	}

	return title
}

func webDumpMeta(displayTitle, scope, status string) string {
	parts := []string{}
	if strings.TrimSpace(scope) != "" && !strings.EqualFold(strings.TrimSpace(displayTitle), strings.TrimSpace(scope)) {
		parts = append(parts, strings.TrimSpace(scope))
	}
	if strings.TrimSpace(status) != "" && !strings.EqualFold(strings.TrimSpace(status), "curated") {
		parts = append(parts, strings.TrimSpace(status))
	}

	return strings.Join(parts, " / ")
}

func webDumpPreview(title, content string) string {
	content = stripMarkdownFrontmatter(strings.TrimSpace(content))
	lines := strings.Split(content, "\n")
	previewLines := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		if line == "" || strings.EqualFold(line, strings.TrimSpace(title)) || webDumpPreviewHeading(line) {
			continue
		}

		previewLines = append(previewLines, line)
	}

	preview := strings.Join(previewLines, " ")
	preview = strings.Join(strings.Fields(preview), " ")
	if preview == "" {
		return "No entries yet"
	}

	const maxPreviewRunes = 220
	runes := []rune(preview)
	if len(runes) > maxPreviewRunes {
		return strings.TrimSpace(string(runes[:maxPreviewRunes])) + "..."
	}

	return preview
}

func webDumpPreviewHeading(line string) bool {
	normalized := strings.Trim(strings.ToLower(strings.TrimSpace(line)), ":")
	switch normalized {
	case "summary", "context", "highlights", "notes", "details", "required behavior", "acceptance criteria":
		return true
	default:
		return false
	}
}

func stripMarkdownFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}

	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			return strings.TrimSpace(strings.Join(lines[index+1:], "\n"))
		}
	}

	return content
}

func webQuizCards(quizzes []QuizCatalogItem) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(quizzes))
	for _, quiz := range quizzes {
		cards = append(cards, webui.ArtifactCardView{
			Family: webArtifactFamilyQuizzes, ID: quiz.Reference, Title: quiz.Title,
			Preview: quiz.Preview, DetailURL: quizFocusURL(quiz.Reference), Revision: quiz.Revision,
		})
	}

	return cards
}

func (p webViewProvider) readProjectArtifact(relativePath string) (string, error) {
	path, err := confinedProjectPath(p.projectContext.ProjectDir, relativePath)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read artifact: %w", err)
	}

	return string(content), nil
}

func confinedProjectPath(projectDir, relativePath string) (string, error) {
	cleanPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relativePath)))
	if cleanPath == "." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("artifact path must stay inside project: %q", relativePath)
	}

	base, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}

	candidate, err := filepath.Abs(filepath.Join(base, cleanPath))
	if err != nil {
		return "", fmt.Errorf("resolve artifact path: %w", err)
	}

	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return "", fmt.Errorf("relativize artifact path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("artifact path must stay inside project: %q", relativePath)
	}

	return candidate, nil
}

func findWebArtifactCard(cards []webui.ArtifactCardView, id string) (webui.ArtifactCardView, bool) {
	for _, card := range cards {
		if card.ID == id {
			return card, true
		}
	}

	return webui.ArtifactCardView{}, false
}

func webArtifactDetailURL(card webui.ArtifactCardView) string {
	return "/" + card.Family + "/detail?id=" + url.QueryEscape(card.ID)
}

func webArtifactTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format("2006-01-02 15:04")
}

func webScopeLabel(slug, title string) string {
	if strings.TrimSpace(slug) == "" {
		return "Project-wide"
	}

	return firstNonEmptyString(title, slug)
}

func webContentState(available bool) string {
	if available {
		return "available"
	}

	return "missing"
}

func caseRelativePath(projectDir string, metadata core.CaseMetadata) string {
	if metadata.SourcePath == "" {
		return ""
	}

	relativePath, err := filepath.Rel(projectDir, metadata.SourcePath)
	if err == nil && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(relativePath)
	}

	return filepath.ToSlash(filepath.Join(core.DirectoryName, core.UserDirectoryName, "cases", metadata.Slug+".md"))
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

func webContentRevision(content string) string {
	sum := sha256.Sum256([]byte(content))

	return hex.EncodeToString(sum[:])
}

func (p webViewProvider) dumpArtifactDetail(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	document, err := core.ReadDumpDocument(ctx, p.projectContext.ProjectDir, request.ID)
	if errors.Is(err, os.ErrNotExist) {
		view := webui.ArtifactDetailView{Family: request.Family, ID: request.ID, Path: request.ID, Missing: true}
		view.Revision = webStableRevision(view)

		return view, nil
	}
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}
	dump := p.service(ctx).dumpView(document)
	dump.Content = document.Body

	return webui.ArtifactDetailView{Family: request.Family, ID: dump.RelativePath, Path: dump.RelativePath, Title: dump.Title, Status: dump.Status, Scope: webScopeLabel(dump.Scope.CaseSlug, dump.UseCaseTitle), SourcePath: dump.SourcePath, UpdatedAt: webArtifactTime(dump.UpdatedAt), Content: dump.Content, Revision: dump.Revision, SaveRequestID: dumpRequestID(""), AppendRequestID: dumpRequestID("")}, nil
}
