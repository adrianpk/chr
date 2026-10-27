package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// PlanView is the TUI-facing projection of a generated plan document.
type PlanView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        quizcore.Scope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

// PlanListResult contains the current plan list and a status message.
type PlanListResult struct {
	Message string
	Plans   []PlanView
}

func (s service) ListPlans() (PlanListResult, error) {
	scopes, err := s.quizReadScopes()
	if err != nil {
		return PlanListResult{}, err
	}

	index, err := s.workspaceIndex()
	if err != nil {
		return PlanListResult{}, err
	}

	plans := make([]PlanView, 0, len(scopes))
	for _, scope := range scopes {
		plans = append(plans, s.planViewFromScope(scope, index))
	}

	for index := range plans {
		plans[index].Index = index
	}

	return PlanListResult{Message: "Plans loaded.", Plans: plans}, nil
}

func (s service) planViewFromScope(scope quizReadScope, index core.WorkspaceIndex) PlanView {
	node, ok := s.latestPlanNodeForScope(scope.scope, index)
	if !ok {
		return PlanView{
			UseCaseTitle: scope.title,
			Title:        scope.title,
			Status:       "missing",
			Scope:        scope.scope,
		}
	}

	return s.planViewFromNode(scope.scope, scope.title, node)
}

func (s service) latestPlanNodeForScope(scope quizcore.Scope, index core.WorkspaceIndex) (core.WorkspaceNode, bool) {
	return latestDocumentNodeForScope(scope, index, core.WorkspaceNodePlan)
}

func (s service) planViewFromNode(scope quizcore.Scope, fallbackTitle string, node core.WorkspaceNode) PlanView {
	title := node.Title
	if strings.TrimSpace(title) == "" {
		title = fallbackTitle
	}

	return s.planViewFromPath(scope, node.RelativePath, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(node.RelativePath))).withMetadata(
		fallbackTitle,
		title,
		node.Status,
		node.ModifiedAt,
	)
}

func (s service) planViewFromPath(scope quizcore.Scope, relativePath, path string) PlanView {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content = nil
	} else if err != nil {
		content = []byte("Unable to read plan: " + err.Error())
	}

	title, status, _ := summarizeWorkspaceMarkdownDocument(string(content), filepath.Base(path))

	return PlanView{
		UseCaseTitle: specUseCaseFallbackTitle(scope),
		Title:        title,
		Status:       status,
		Scope:        scope,
		RelativePath: filepath.ToSlash(relativePath),
		UpdatedAt:    fileModifiedAt(path),
		Content:      string(content),
	}
}

func (v PlanView) withMetadata(useCaseTitle, title, status string, updatedAt time.Time) PlanView {
	if strings.TrimSpace(useCaseTitle) != "" {
		v.UseCaseTitle = useCaseTitle
	}

	if strings.TrimSpace(title) != "" {
		v.Title = title
	}

	if strings.TrimSpace(status) != "" {
		v.Status = status
	}

	if !updatedAt.IsZero() {
		v.UpdatedAt = updatedAt
	}

	return v
}
