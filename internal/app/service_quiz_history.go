package app

import (
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

const (
	quizHistoryEntryLimit = 100
	quizHistoryByteLimit  = 128 * 1024
)

func (s service) quizHistoryContext(scope quizcore.Scope, entries []quizcore.Entry) (core.QuizModelContextSection, error) {
	if scope.CaseSlug != "" {
		general, err := s.quizRepository().LoadRawCaptureEntries(s.ctx, quizcore.Scope{ProjectID: scope.ProjectID})
		if err != nil {
			return core.QuizModelContextSection{}, err
		}

		entries = append(general, entries...)
	}

	if len(entries) > quizHistoryEntryLimit {
		return core.QuizModelContextSection{}, fmt.Errorf("quiz discovery history exceeds %d entries; narrow the selected scope", quizHistoryEntryLimit)
	}

	content, err := json.Marshal(quizModelRecentEntryPayloads(entries, len(entries)))
	if err != nil {
		return core.QuizModelContextSection{}, err
	}

	if len(content) > quizHistoryByteLimit {
		return core.QuizModelContextSection{}, fmt.Errorf("quiz discovery history exceeds %d bytes; narrow the selected scope", quizHistoryByteLimit)
	}

	return core.QuizModelContextSection{
		Name: "quiz_history", Layer: "quiz", Authority: "quiz_context", Content: string(content),
	}, nil
}
