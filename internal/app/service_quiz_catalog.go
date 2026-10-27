package app

import (
	"strings"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizCatalogItem describes a known scope using confirmed retained answers.
// Reference opens the same scope without starting or replacing its session.
type QuizCatalogItem struct {
	Scope                       quizcore.Scope
	Title, Reference, Preview   string
	SessionID, Revision, Status string
	AnswerCount                 int
}

// QuizCatalog separates answered cards from eligible unanswered destinations.
// The project card is present in Items even when it is also pending.
type QuizCatalog struct {
	Items, Pending []QuizCatalogItem
}

// ReadQuizCatalog observes existing project/case scopes without creating data.
// Native runtime counters and curated document existence cannot imply answers.
func (s service) ReadQuizCatalog() (QuizCatalog, error) {
	scopes, err := s.quizReadScopes()
	if err != nil {
		return QuizCatalog{}, err
	}

	var catalog QuizCatalog

	for _, scope := range scopes {
		item, err := s.quizCatalogItem(scope)
		if err != nil {
			return QuizCatalog{}, err
		}

		if item.Scope.CaseSlug == "" || item.AnswerCount > 0 {
			catalog.Items = append(catalog.Items, item)
		}

		if item.AnswerCount == 0 {
			catalog.Pending = append(catalog.Pending, item)
		}
	}

	return catalog, nil
}

func (s service) quizCatalogItem(scope quizReadScope) (QuizCatalogItem, error) {
	item := QuizCatalogItem{Scope: scope.scope, Title: scope.title}
	if scope.scope.CaseSlug == "" {
		item.Title = "Project Quiz"
	}

	paths, err := s.quizRepository().Paths(scope.scope)
	if err != nil {
		return item, err
	}

	observed, err := s.quizRepository().Observe(s.ctx, scope.scope)
	if err != nil {
		return item, err
	}

	item.Reference, item.Revision = paths.CuratedRelativePath, observed.Revision
	item.SessionID, item.Status = observed.State.SessionID, string(observed.State.Status)

	for _, entry := range observed.Entries {
		if entry.Status != quizcore.EntryStatusAnswered {
			continue
		}

		item.AnswerCount++
		if item.AnswerCount == 1 {
			// Preview only opening answer content; the complete history stays in
			// the focused view. Bound rendered bytes before sending the catalog.
			item.Preview = quizCatalogPreview(entry)
		}
	}

	return item, nil
}

func quizCatalogPreview(entry quizcore.Entry) string {
	const previewLimit = 1024

	text := []rune(strings.TrimSpace(entry.Question + "\n" + entry.Answer))
	if len(text) > previewLimit {
		text = text[:previewLimit]
	}

	return string(text)
}
