package app

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// Real PTY keys choose Fast, answer several questions, keep visible history,
// reject a concurrent stale answer, reload explicitly, cancel, resume and finish.
func TestTUIQuizTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)
	terminal := startTUITerminal(t, operations, "TestTUIQuizProcess")
	terminal.see(t, "Alpha navigation")
	contextTerminalQuizUnavailable(t, terminal, operations)
	terminal.keys(t, "\x1bOR")
	terminal.see(t, "Mode: Deep")
	terminal.keys(t, "m")
	terminal.see(t, "Mode: Fast")
	terminal.keys(t, "s")
	terminal.see(t, "TERMINAL_QUESTION_1")
	terminal.keys(t, "TERMINAL_ANSWER_1\r")
	terminal.see(t, "TERMINAL_QUESTION_2")
	terminal.see(t, "Answer: TERMINAL_ANSWER_1")
	terminal.keys(t, "TERMINAL_ANSWER_2\r")
	terminal.see(t, "TERMINAL_QUESTION_3")
	terminal.see(t, "Answer: TERMINAL_ANSWER_2")
	staleTerminalQuiz(t, terminal, operations)
	finishTerminalQuiz(t, terminal)
	terminal.stop(t)

	observed, err := operations.quizRepository().Observe(t.Context(), operations.quizScope(quizcore.Scope{CaseSlug: "alpha-navigation"}))
	if err != nil || observed.State.Status != quizcore.RuntimeStatusDone || observed.State.Mode != quizcore.ModeFast || observed.State.AnswersRecorded != 3 || observed.State.SkipsRecorded != 1 {
		t.Fatal(observed.State, err)
	}

	if len(quizAnswerHistory(observed.Entries)) != 3 {
		t.Fatal("terminal controls lost saved history", observed.Entries)
	}

	checkTerminalQuizScope(t, operations)
}

func contextTerminalQuizUnavailable(t *testing.T, terminal *guidedTerminal, operations service) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, ":")
	terminal.see(t, "Command Palette")
	terminal.keys(t, "context-aware\r")
	terminal.see(t, "Operation: blocked")
	terminal.see(t, "chr author quiz new")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Alpha navigation")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("unavailable command wrote canonical files")
	}
}

func checkTerminalQuizScope(t *testing.T, operations service) {
	t.Helper()

	for _, slug := range []string{"", "beta-navigation"} {
		observed, err := operations.quizRepository().Observe(t.Context(), operations.quizScope(quizcore.Scope{CaseSlug: slug}))
		if err != nil || observed.Exists {
			t.Fatal("terminal mutated another quiz scope", slug, err)
		}
	}
}

func staleTerminalQuiz(t *testing.T, terminal *guidedTerminal, operations service) {
	t.Helper()
	terminal.keys(t, "UNSENT_STALE_TERMINAL_ANSWER")

	observed, err := operations.quizRepository().Observe(t.Context(), operations.quizScope(quizcore.Scope{CaseSlug: "alpha-navigation"}))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.AnswerQuiz(QuizAnswerRequest{Scope: quizcore.Scope{CaseSlug: "alpha-navigation"}, Binding: quizStateBinding(observed.State, "concurrent-terminal-answer"), Answer: "CONCURRENT_TERMINAL_ANSWER"}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, "\r")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "r")
	terminal.see(t, "Reload the current quiz and discard the unsent answer?")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "\x1b")
	terminal.see(t, "UNSENT_STALE_TERMINAL_ANSWER")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("stale answer or cancelled reload changed canonical files")
	}

	terminal.keys(t, "\x1bo")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "r")
	terminal.see(t, "Reload the current quiz and discard the unsent answer?")
	terminal.keys(t, "\r")
	terminal.see(t, "TERMINAL_QUESTION_4")
	terminal.see(t, "Answer: CONCURRENT_TERMINAL_ANSWER")
}

func finishTerminalQuiz(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "\x1bx")
	terminal.see(t, "Cancel this quiz? Saved answers are retained")
	terminal.keys(t, "\x1b")
	terminal.see(t, "TERMINAL_QUESTION_4")
	terminal.keys(t, "\x1bx")
	terminal.see(t, "Cancel this quiz? Saved answers are retained")
	terminal.keys(t, "\r")
	terminal.see(t, "Quiz cancelled.")
	terminal.keys(t, "u")
	terminal.see(t, "TERMINAL_QUESTION_5")
	terminal.keys(t, "\x1bs")
	terminal.see(t, "TERMINAL_QUESTION_6")
	terminal.keys(t, "\x1bd")
	terminal.see(t, "Quiz completed.")
}

// The owned child uses the production root loader and managed host. Only model
// output/runtime binding is deterministic; terminal, revision and document IO is real.
func TestTUIQuizProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned quiz terminal fixture")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_GUIDED_TUI_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	projectContext := commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_GUIDED_TUI_ID")}
	factory := commandFactory{deps: testAppDependencies(), io: IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}}

	options, err := factory.loadTUIOptions(t.Context(), env, projectContext)
	if err != nil {
		t.Fatal(err)
	}

	operations := newService(t.Context(), env, projectContext, testAssetsFS())
	operations.quizModel, operations.now, operations.generationBinder = terminalQuizModel(), serviceQuizClock(), testGenerationBinder

	options.QuizExecutor = newTUIQuizActionExecutor(operations)
	if err := runManagedTUI(t.Context(), options, operations); err != nil {
		t.Fatal(err)
	}
}

func terminalQuizModel() *fakeServiceQuizModel {
	now := serviceQuizClock()()

	candidates := []quizcore.Candidate{}
	for index := 2; index <= 6; index++ {
		candidates = append(candidates, serviceQuizCandidate(fmt.Sprintf("terminal-q-%d", index), fmt.Sprintf("TERMINAL_QUESTION_%d", index), now))
	}

	question := serviceQuizQuestion(quizcore.Question{ID: "terminal-q-1", Prompt: "TERMINAL_QUESTION_1", Topic: "workflow", Depth: 1, TrunkID: "terminal-q-1", CreatedAt: now})
	response := serviceQuizModelResponseWithQuestion("terminal-fast-batch", question)
	response.Candidates = candidates

	return &fakeServiceQuizModel{responses: []quizmodel.Response{response}}
}
