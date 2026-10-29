package app

import (
	"sort"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizAnswerRecord projects retained answers without transport-owned markup.
type QuizAnswerRecord struct {
	ID, Question, Answer string
	AnsweredAt           time.Time
}

func quizAnswerHistory(entries []quizcore.Entry) []QuizAnswerRecord {
	history := make([]QuizAnswerRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.Status == quizcore.EntryStatusAnswered {
			history = append(history, QuizAnswerRecord{ID: entry.ID, Question: entry.Question, Answer: entry.Answer, AnsweredAt: entry.AnsweredAt})
		}
	}

	sort.Slice(history, func(i, j int) bool {
		if history[i].AnsweredAt.Equal(history[j].AnsweredAt) {
			return history[i].ID > history[j].ID
		}

		return history[i].AnsweredAt.After(history[j].AnsweredAt)
	})

	return history
}
