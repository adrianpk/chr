package app

import (
	"encoding/json"
	"fmt"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

const (
	quizHistoryEntryLimit = 100
	quizHistoryByteLimit  = 128 * 1024
)

func (s service) quizHistoryContext(scope quizcore.Scope, entries []quizcore.Entry) (quizmodel.ContextSection, error) {
	if scope.CaseSlug != "" {
		general, err := s.quizRepository().LoadRawCaptureEntries(s.ctx, quizcore.Scope{ProjectID: scope.ProjectID})
		if err != nil {
			return quizmodel.ContextSection{}, err
		}

		entries = append(general, entries...)
	}

	if len(entries) > quizHistoryEntryLimit {
		return quizmodel.ContextSection{}, fmt.Errorf("quiz discovery history exceeds %d entries; narrow the selected scope", quizHistoryEntryLimit)
	}

	content, err := json.Marshal(quizModelRecentEntryPayloads(entries, len(entries)))
	if err != nil {
		return quizmodel.ContextSection{}, err
	}

	if len(content) > quizHistoryByteLimit {
		return quizmodel.ContextSection{}, fmt.Errorf("quiz discovery history exceeds %d bytes; narrow the selected scope", quizHistoryByteLimit)
	}

	return quizmodel.ContextSection{
		Name: "quiz_history", Layer: "quiz", Authority: "quiz_context", Content: string(content),
	}, nil
}
