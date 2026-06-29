package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/core"
	"forge.adrianpk.com/chc/internal/tui"
	"forge.adrianpk.com/chc/internal/webui"
)

const (
	webArtifactFamilyCases   = "cases"
	webArtifactFamilyDumps   = "dumps"
	webArtifactFamilyQuizzes = "quizzes"
	webArtifactFamilySpecs   = "specs"
	webArtifactFamilyPlans   = "plans"
)

func (p webViewProvider) CatalogView(ctx context.Context, family string) (webui.ArtifactCatalogView, error) {
	operations := newService(ctx, p.env, p.projectContext, p.assetsFS)

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

	return webArtifactCatalog(webArtifactFamilyDumps, webDumpCards(dumps)), nil
}

func quizCatalog(operations service) (webui.ArtifactCatalogView, error) {
	quizzes, err := loadTUIQuizzes(operations)
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	return webArtifactCatalog(webArtifactFamilyQuizzes, webQuizCards(quizzes)), nil
}

func specCatalog(operations service) (webui.ArtifactCatalogView, error) {
	specs, err := loadTUISpecs(operations)
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	return webArtifactCatalog(webArtifactFamilySpecs, webSpecCards(specs)), nil
}

func planCatalog(operations service) (webui.ArtifactCatalogView, error) {
	plans, err := loadTUIPlans(operations)
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	return webArtifactCatalog(webArtifactFamilyPlans, webPlanCards(plans)), nil
}

func (p webViewProvider) ArtifactDetailView(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	catalog, err := p.CatalogView(ctx, request.Family)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	card, ok := findWebArtifactCard(catalog.Items, request.ID)
	if !ok {
		return webui.ArtifactDetailView{
			Family:  request.Family,
			Title:   request.ID,
			Path:    request.ID,
			Missing: true,
		}, nil
	}

	content := card.Preview
	if card.Path != "" {
		readContent, err := p.readProjectArtifact(card.Path)
		if err == nil {
			content = readContent
		} else if !errors.Is(err, os.ErrNotExist) {
			return webui.ArtifactDetailView{}, err
		}
	}

	return webui.ArtifactDetailView{
		Family:     card.Family,
		Title:      card.Title,
		Status:     card.Status,
		Scope:      card.Scope,
		Path:       card.Path,
		SourcePath: card.SourcePath,
		UpdatedAt:  card.UpdatedAt,
		Content:    content,
	}, nil
}

func webArtifactCatalog(family string, items []webui.ArtifactCardView) webui.ArtifactCatalogView {
	return webui.ArtifactCatalogView{
		Family:       family,
		Title:        webui.CatalogTitle(family),
		EmptyMessage: "No " + family + " found.",
		Items:        items,
	}
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
		}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	return cards
}

func webDumpCards(dumps []tui.DumpView) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(dumps))
	for _, dump := range dumps {
		card := webui.ArtifactCardView{
			Family:       webArtifactFamilyDumps,
			ID:           dump.RelativePath,
			Title:        dump.Title,
			Status:       dump.Status,
			Scope:        webScopeLabel(dump.Scope.CaseSlug, dump.UseCaseTitle),
			Path:         dump.RelativePath,
			SourcePath:   dump.SourcePath,
			UpdatedAt:    webArtifactTime(dump.UpdatedAt),
			Preview:      dump.Preview,
			ContentState: webContentState(dump.RelativePath != ""),
		}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	return cards
}

func webQuizCards(quizzes []tui.QuizView) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(quizzes))
	for _, quiz := range quizzes {
		path := firstNonEmptyString(quiz.CuratedPath, quiz.RuntimePath, quiz.RawPath)
		card := webui.ArtifactCardView{
			Family:       webArtifactFamilyQuizzes,
			ID:           webQuizID(quiz, path),
			Title:        quiz.Title,
			Status:       quiz.Status,
			Scope:        webScopeLabel(quiz.Scope.CaseSlug, quiz.Title),
			Path:         path,
			UpdatedAt:    webArtifactTime(quiz.UpdatedAt),
			Preview:      firstNonEmptyString(quiz.CuratedPreview, quiz.ActiveQuestionSummary, quiz.TrailSummary),
			ContentState: webContentState(path != ""),
		}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	return cards
}

func webSpecCards(specs []tui.SpecView) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(specs))
	for _, spec := range specs {
		card := webDocumentCard(webArtifactFamilySpecs, spec.Title, spec.Status, spec.UseCaseTitle, spec.RelativePath, spec.UpdatedAt, spec.Content)
		cards = append(cards, card)
	}

	return cards
}

func webPlanCards(plans []tui.PlanView) []webui.ArtifactCardView {
	cards := make([]webui.ArtifactCardView, 0, len(plans))
	for _, plan := range plans {
		card := webDocumentCard(webArtifactFamilyPlans, plan.Title, plan.Status, plan.UseCaseTitle, plan.RelativePath, plan.UpdatedAt, plan.Content)
		cards = append(cards, card)
	}

	return cards
}

func webDocumentCard(family, title, status, scope, relativePath string, updatedAt time.Time, content string) webui.ArtifactCardView {
	card := webui.ArtifactCardView{
		Family:       family,
		ID:           firstNonEmptyString(relativePath, title),
		Title:        title,
		Status:       status,
		Scope:        scope,
		Path:         relativePath,
		UpdatedAt:    webArtifactTime(updatedAt),
		Preview:      content,
		ContentState: webContentState(relativePath != ""),
	}
	card.DetailURL = webArtifactDetailURL(card)

	return card
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

func webQuizID(quiz tui.QuizView, path string) string {
	if strings.TrimSpace(path) != "" {
		return path
	}

	if strings.TrimSpace(quiz.Scope.CaseSlug) != "" {
		return "case:" + quiz.Scope.CaseSlug
	}

	return "project"
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
