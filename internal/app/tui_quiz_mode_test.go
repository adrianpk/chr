package app

import (
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

// The production adapter must submit the visible choice, rather than silently
// forcing Deep generation when the user selected Fast before Start.
func TestTUIQuizStartMode(t *testing.T) {
	for _, mode := range []quizcore.Mode{quizcore.ModeDeep, quizcore.ModeFast} {
		t.Run(string(mode), func(t *testing.T) {
			operations, _ := newRetainedQuizFixture(t)

			executor := newTUIQuizActionExecutor(operations)
			if _, err := executor.ExecuteQuizAction(tui.QuizActionRequest{Action: tui.QuizActionStart, Mode: string(mode)}); err != nil {
				t.Fatal(err)
			}

			observed, err := operations.quizRepository().Observe(t.Context(), operations.quizScope(quizcore.Scope{}))
			if err != nil || observed.State.Mode != mode {
				t.Fatal("adapter changed the selected mode", observed.State.Mode, err)
			}
		})
	}
}
