package app

import "forge.adrianpk.com/chr/internal/tui"

func tuiQuizHistory(records []QuizAnswerRecord) []tui.QuizAnswerRecord {
	history := make([]tui.QuizAnswerRecord, 0, len(records))
	for _, record := range records {
		history = append(history, tui.QuizAnswerRecord{Question: record.Question, Answer: record.Answer, AnsweredAt: record.AnsweredAt})
	}

	return history
}
