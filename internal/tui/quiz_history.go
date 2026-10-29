package tui

import (
	"time"
)

// QuizAnswerRecord contains a saved question and its complete answer.
type QuizAnswerRecord struct {
	Question, Answer string
	AnsweredAt       time.Time
}

func (m *Model) renderQuizHistory(history []QuizAnswerRecord, width int) []string {
	if len(history) == 0 {
		return []string{"", "Answer history: no saved answers."}
	}

	lines := []string{}
	for _, record := range history {
		lines = append(lines, record.AnsweredAt.UTC().Format(time.RFC3339))
		lines = append(lines, wrapPlainText("Question: "+record.Question, max(1, width))...)
		lines = append(lines, wrapPlainText("Answer: "+record.Answer, max(1, width))...)
		lines = append(lines, "")
	}

	lines = lines[:len(lines)-1]
	limit := m.quizHistoryPageSize()
	start := min(m.quizHistoryScroll, max(0, len(lines)-limit))

	return append([]string{"", "Answer history (PgUp/PgDn):"}, lines[start:min(start+limit, len(lines))]...)
}

func (m *Model) quizHistoryPageSize() int {
	if m.height == 0 {
		return 8
	}

	return max(1, min(8, m.height-23))
}

func (m *Model) scrollQuizHistoryKey(key string) bool {
	switch key {
	case "pgup":
		m.quizHistoryScroll = max(0, m.quizHistoryScroll-m.quizHistoryPageSize())
	case "pgdown":
		m.quizHistoryScroll += m.quizHistoryPageSize()
	default:
		return false
	}

	return true
}
