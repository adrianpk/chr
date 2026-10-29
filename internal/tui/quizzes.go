package tui

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// QuizScope identifies the project or use-case scope for a quiz.
type QuizScope struct {
	CaseSlug string
}

// QuizView is the terminal projection of a quiz session or available scope.
type QuizView struct {
	History               []QuizAnswerRecord
	Binding               QuizActionBinding
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
	CandidateCount        int
	BranchReturnReason    string
	UpdatedAt             time.Time
	ActiveQuestionSummary string
	TrailSummary          string
	CuratedPreview        string
}

func (m *Model) renderQuizFocusList(limit int) []string {
	if len(m.visibleQuizzes()) == 0 {
		return []string{"No quiz use cases found."}
	}

	lines := []string{shellMutedStyle.Render("Use Case")}

	for _, quiz := range m.visibleQuizzes() {
		prefix := " "
		if quiz.Index == m.selectedQuiz {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %s", prefix, quizListLabel(quiz)))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m *Model) renderSelectedQuizDetailLines(width int) []string {
	quiz, ok := m.selectedQuizView()
	if !ok {
		return []string{"No quiz selected."}
	}

	if m.quizAnswer.Active && quizScopeLabel(m.quizAnswer.Scope) == quizScopeLabel(quiz.Scope) {
		return m.renderSelectedQuizAnswerLines(quiz, width)
	}

	lines := []string{
		shellMutedStyle.Render(m.quizActionHint(quiz)),
		"",
		quiz.Title,
		quizHumanStatus(quiz),
		quizModeLine(quiz),
	}

	if strings.TrimSpace(quiz.ActiveQuestionSummary) != "" {
		lines = append(lines, "")
		lines = append(lines, wrapPlainText(quiz.ActiveQuestionSummary, maxInt(width, 1))...)
	}

	if m.quizBusy {
		lines = append(lines, "", "Working...")
		lines = append(lines, "  "+m.statusMessage)
	}

	if strings.TrimSpace(quiz.ActiveQuestionSummary) == "" && strings.TrimSpace(quiz.CuratedPreview) != "" {
		lines = append(lines, "", "Notes:")
		lines = append(lines, indentLines(renderContentReadLines(quiz.CuratedPreview, maxInt(width-2, 1)), "  ")...)
	}

	return append(lines, m.renderQuizHistory(quiz.History, width)...)
}

func (m *Model) renderSelectedQuizAnswerLines(quiz QuizView, width int) []string {
	lines := []string{shellMutedStyle.Render(m.quizActionHint(quiz)), quizModeLine(quiz), ""}
	lines = append(lines, wrapPlainText(m.quizAnswer.Question, maxInt(width, 1))...)
	lines = append(lines, strings.Split(m.renderQuizAnswerInput(maxInt(width, 1), 6), "\n")...)
	lines = append(lines, "alt+x cancels quiz  alt+m changes mode")

	return append(lines, m.renderQuizHistory(m.quizAnswer.History, width)...)
}

func (m *Model) renderQuizzes() string {
	if len(m.visibleQuizzes()) == 0 {
		return "Quizzes\n  No quiz use cases found."
	}

	var output strings.Builder

	output.WriteString("Quizzes\n")

	for _, quiz := range m.visibleQuizzes() {
		prefix := "  "
		if quiz.Index == m.selectedQuiz {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(quizListLabel(quiz))
		output.WriteString("  ")
		output.WriteString(shellMutedStyle.Render(quizHumanStatus(quiz)))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(strings.Join(m.renderSelectedQuizDetailLines(72), "\n"))

	return strings.TrimRight(output.String(), "\n")
}

func quizHumanStatus(quiz QuizView) string {
	switch quiz.Status {
	case "not_started":
		return "Not started"
	case "active":
		return fmt.Sprintf("Question %d", maxInt(quiz.QuestionsAsked, 1))
	case "done":
		return "Done"
	case "cancelled":
		return "Cancelled"
	case "blocked":
		return "Blocked"
	default:
		if strings.TrimSpace(quiz.Status) == "" {
			return "Not started"
		}

		return quiz.Status
	}
}

func (m *Model) quizActionHint(quiz QuizView) string {
	if m.wizardQuizVisible() {
		return m.wizardQuizHint(quiz)
	}

	return m.manualQuizActionHint(quiz)
}

func (m *Model) manualQuizActionHint(quiz QuizView) string {
	if m.quizBusy {
		return "Working"
	}

	if m.quizAnswer.Active && quizScopeLabel(m.quizAnswer.Scope) == quizScopeLabel(quiz.Scope) {
		return "enter submits  shift+enter inserts line  alt+s skip  alt+d done  esc cancel"
	}

	if strings.TrimSpace(quiz.ActiveQuestionSummary) != "" {
		return "enter answers  m mode  x cancel  r refresh"
	}

	if quiz.Status == "done" && strings.TrimSpace(quiz.CuratedPath) != "" {
		return "s restart  m Deep/Fast  r refresh"
	}

	if quiz.Status == "cancelled" || quiz.Status == "blocked" || quiz.Status == "idle" {
		return "u resume  s restart  m Deep/Fast  x cancel  r refresh"
	}

	return "s start  m Deep/Fast  r refresh"
}

func (m *Model) selectedQuizView() (QuizView, bool) {
	if m.wizardQuizVisible() {
		return *m.guided.View.Quiz, true
	}

	for _, quiz := range m.visibleQuizzes() {
		if quiz.Index == m.selectedQuiz {
			return quiz, true
		}
	}

	if len(m.visibleQuizzes()) > 0 {
		return m.visibleQuizzes()[0], true
	}

	return QuizView{}, false
}

func (m *Model) selectPreviousQuiz() {
	if len(m.visibleQuizzes()) == 0 {
		return
	}

	m.quizHistoryScroll = 0

	position := m.selectedQuizPosition()
	if position == 0 {
		m.selectedQuiz = m.visibleQuizzes()[len(m.visibleQuizzes())-1].Index

		return
	}

	m.selectedQuiz = m.visibleQuizzes()[position-1].Index
}

func (m *Model) selectNextQuiz() {
	if len(m.visibleQuizzes()) == 0 {
		return
	}

	m.quizHistoryScroll = 0

	position := m.selectedQuizPosition()
	if position >= len(m.visibleQuizzes())-1 {
		m.selectedQuiz = m.visibleQuizzes()[0].Index

		return
	}

	m.selectedQuiz = m.visibleQuizzes()[position+1].Index
}

func (m *Model) selectQuizByScope(scope QuizScope) {
	for _, quiz := range m.visibleQuizzes() {
		if quiz.Scope.CaseSlug == scope.CaseSlug {
			m.selectedQuiz = quiz.Index

			return
		}
	}

	if len(m.visibleQuizzes()) > 0 {
		m.selectedQuiz = m.visibleQuizzes()[0].Index

		return
	}

	m.selectedQuiz = 0
}

func (m *Model) selectedQuizScope() QuizScope {
	quiz, ok := m.selectedQuizView()
	if !ok {
		return QuizScope{CaseSlug: m.caseContext.CaseSlug}
	}

	return quiz.Scope
}

func (m *Model) updateQuizzesKey(msg tea.KeyMsg) tea.Cmd {
	if m.scrollQuizHistoryKey(msg.String()) {
		return nil
	}

	switch msg.String() {
	case "s":
		cmd := m.startQuizAction(m.newQuizActionState(QuizActionStart).Request)

		return cmd
	case "m":
		return m.cycleQuizMode()
	case "u":
		return m.startQuizAction(m.newQuizActionState(QuizActionResume).Request)
	case "x":
		m.planQuizAction(QuizActionCancel)
	case "enter":
		m.planQuizAction(QuizActionAnswer)
	case "r":
		cmd := m.startQuizAction(QuizActionRequest{Action: QuizActionRefresh, Scope: m.selectedQuizScope()})

		return cmd
	case "esc":
		m.selectPrimaryPane()
	default:
		return m.updateNavigationKey(msg)
	}

	return nil
}

func (m *Model) newQuizActionState(action QuizActionID) QuizActionState {
	scope := m.selectedQuizScope()
	title := quizActionTitle(action)

	var binding QuizActionBinding
	if quiz, ok := m.selectedQuizView(); ok {
		binding = quiz.Binding
	}

	if binding.ExpectedRevision != "" {
		binding.OperationID = rand.Text()
	}

	mode := quizActionMode(action)
	if action == QuizActionStart {
		if quiz, ok := m.selectedQuizView(); ok {
			mode = quiz.Mode
		}
	}

	return QuizActionState{
		Pending: true,
		Request: QuizActionRequest{
			Binding: binding,
			Guided:  m.capturedWizardQuiz(),
			Action:  action,
			Scope:   scope,
			Mode:    mode,
		},
		Title: title,
		Details: []string{
			"Use Case: " + quizUseCaseLabel(scope),
		},
	}
}

func (m *Model) selectedQuizPosition() int {
	for index, quiz := range m.visibleQuizzes() {
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
	return quizUseCaseLabel(scope)
}

func quizUseCaseLabel(scope QuizScope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "Project-wide"
	}

	return scope.CaseSlug
}

func quizActionTitle(action QuizActionID) string {
	titles := map[QuizActionID]string{
		QuizActionStart:    "Start quiz",
		QuizActionResume:   "Resume quiz",
		QuizActionSkip:     "Skip active question",
		QuizActionComplete: "Complete quiz",
		QuizActionCancel:   "Cancel quiz",
		QuizActionRefresh:  "Refresh quizzes",
		QuizActionDeepMode: "Switch quiz to deep mode",
		QuizActionFastMode: "Switch quiz to fast mode",
	}

	title, ok := titles[action]
	if !ok {
		return "Quiz action"
	}

	return title
}

func quizActionMode(action QuizActionID) string {
	switch action {
	case QuizActionDeepMode:
		return "deep"
	case QuizActionFastMode:
		return "fast"
	default:
		return ""
	}
}
