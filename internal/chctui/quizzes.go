package chctui

import (
	"fmt"
	"strings"
	"time"
)

type QuizScope struct {
	CaseSlug string
}

type QuizView struct {
	Index                 int
	Title                 string
	Scope                 QuizScope
	Status                string
	Mode                  string
	RawPath               string
	CuratedPath           string
	RuntimePath           string
	QuestionBudget        int
	QuestionsAsked        int
	AnswersRecorded       int
	SkipsRecorded         int
	UpdatedAt             time.Time
	ActiveQuestionSummary string
	TrailSummary          string
	CuratedPreview        string
}

func (m Model) renderQuizFocusList(limit int) []string {
	if len(m.quizzes) == 0 {
		return []string{"No quiz scopes found."}
	}

	lines := []string{shellMutedStyle.Render("Scope        Status       Mode   Asked  Title")}

	for _, quiz := range m.quizzes {
		prefix := " "
		if quiz.Index == m.selectedQuiz {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf(
			"%s %-12s %-12s %-6s %-6d %s",
			prefix,
			quizScopeLabel(quiz.Scope),
			quiz.Status,
			quiz.Mode,
			quiz.QuestionsAsked,
			quiz.Title,
		))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m Model) renderSelectedQuizDetailLines(width int) []string {
	quiz, ok := m.selectedQuizView()
	if !ok {
		return []string{"No quiz selected."}
	}

	lines := []string{
		fmt.Sprintf("Scope: %s", quizScopeLabel(quiz.Scope)),
		fmt.Sprintf("Status: %s", quiz.Status),
		fmt.Sprintf("Mode: %s", quiz.Mode),
		fmt.Sprintf("Title: %s", quiz.Title),
		fmt.Sprintf("Questions: %d/%d", quiz.QuestionsAsked, quiz.QuestionBudget),
		fmt.Sprintf("Answers: %d", quiz.AnswersRecorded),
		fmt.Sprintf("Skips: %d", quiz.SkipsRecorded),
		fmt.Sprintf("Updated: %s", renderTime(quiz.UpdatedAt)),
		fmt.Sprintf("Raw: %s", quiz.RawPath),
		fmt.Sprintf("Curated: %s", quiz.CuratedPath),
		fmt.Sprintf("Runtime: %s", quiz.RuntimePath),
	}

	if strings.TrimSpace(quiz.ActiveQuestionSummary) != "" {
		lines = append(lines, "", "Active question:")
		lines = append(lines, indentLines(wrapPlainText(quiz.ActiveQuestionSummary, maxInt(width-2, 1)), "  ")...)
	}

	if strings.TrimSpace(quiz.TrailSummary) != "" {
		lines = append(lines, "", "Trail:")
		lines = append(lines, indentLines(wrapPlainText(quiz.TrailSummary, maxInt(width-2, 1)), "  ")...)
	}

	lines = append(lines, "", "Curated preview:")
	lines = append(lines, indentLines(wrapPlainText(emptyValue(quiz.CuratedPreview), maxInt(width-2, 1)), "  ")...)

	return lines
}

func (m Model) renderQuizzes() string {
	if len(m.quizzes) == 0 {
		return "Quizzes\n  No quiz scopes found."
	}

	var output strings.Builder

	output.WriteString("Quizzes\n")
	output.WriteString("  Scope        Status       Mode   Asked  Title\n")

	for _, quiz := range m.quizzes {
		prefix := "  "
		if quiz.Index == m.selectedQuiz {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(fmt.Sprintf(
			"%-12s %-12s %-6s %-6d %s",
			quizScopeLabel(quiz.Scope),
			quiz.Status,
			quiz.Mode,
			quiz.QuestionsAsked,
			quiz.Title,
		))
		output.WriteString("\n")
	}

	output.WriteString("\nSelected Quiz\n")
	output.WriteString(strings.Join(m.renderSelectedQuizDetailLines(72), "\n"))

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) selectedQuizView() (QuizView, bool) {
	for _, quiz := range m.quizzes {
		if quiz.Index == m.selectedQuiz {
			return quiz, true
		}
	}

	if len(m.quizzes) > 0 {
		return m.quizzes[0], true
	}

	return QuizView{}, false
}

func (m *Model) selectPreviousQuiz() {
	if len(m.quizzes) == 0 {
		return
	}

	position := m.selectedQuizPosition()
	if position == 0 {
		m.selectedQuiz = m.quizzes[len(m.quizzes)-1].Index

		return
	}

	m.selectedQuiz = m.quizzes[position-1].Index
}

func (m *Model) selectNextQuiz() {
	if len(m.quizzes) == 0 {
		return
	}

	position := m.selectedQuizPosition()
	if position >= len(m.quizzes)-1 {
		m.selectedQuiz = m.quizzes[0].Index

		return
	}

	m.selectedQuiz = m.quizzes[position+1].Index
}

func (m Model) selectedQuizPosition() int {
	for index, quiz := range m.quizzes {
		if quiz.Index == m.selectedQuiz {
			return index
		}
	}

	return 0
}

func normalizeQuizViews(quizzes []QuizView) []QuizView {
	normalized := append([]QuizView(nil), quizzes...)
	for index := range normalized {
		normalized[index].Index = index
		if strings.TrimSpace(normalized[index].Status) == "" {
			normalized[index].Status = "unknown"
		}

		if strings.TrimSpace(normalized[index].Mode) == "" {
			normalized[index].Mode = "deep"
		}
	}

	return normalized
}

func quizScopeLabel(scope QuizScope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "project"
	}

	return scope.CaseSlug
}
