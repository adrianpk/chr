package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestGuidedUseCaseQuizCommitted binds clarification to retained bytes and seals sufficient answers.
func TestGuidedUseCaseQuizCommitted(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "research")

	dump := session.Guided.Checkpoints[0]
	replaceGuidedWorkingDump(t, operations.projectContext.ProjectDir, dump)

	command := guidedQuizCommand(session, "start-quiz", "quiz", "")

	current, err := operations.ConductGuidedQuiz(t.Context(), command)
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedQuizSource(t, model)

	quiz := current.Guided.Quiz
	if quiz == nil || quiz.QuestionID != "question-1" || current.Guided.Status != "waiting_input" {
		t.Fatalf("current question missing: %#v", current.Guided)
	}

	answer := guidedQuizCommand(current, "answer-first", "answer", "Use the existing bounded workflow.")

	completed, err := operations.ConductGuidedQuiz(t.Context(), answer)
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedQuizCompletion(t, completed, dump)

	replay, err := operations.ConductGuidedQuiz(t.Context(), answer)
	if err != nil || replay.Revision != completed.Revision || len(model.requests) != 2 {
		t.Fatalf("duplicate answer repeated generation: %v", err)
	}
}

func assertGuidedQuizSource(t *testing.T, model *fakeServiceQuizModel) {
	t.Helper()

	found := false

	for _, section := range model.requests[0].ContextSections {
		if strings.Contains(section.Content, "Deliver a bounded workflow.") {
			found = true
		}

		if strings.Contains(section.Content, "Later replacement") {
			t.Fatal("quiz used a working copy")
		}
	}

	if !found {
		t.Fatal("committed dump missing from quiz request")
	}
}

// TestGuidedUseCaseQuizBinding rejects stale, foreign and empty answers before reserving effects.
func TestGuidedUseCaseQuizBinding(t *testing.T) {
	operations, session, _ := guidedQuizFixture(t, "default")

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	valid := guidedQuizCommand(current, "answer", "answer", "A meaningful answer.")

	for _, variant := range []string{"question", "source", "revision", "empty", "run"} {
		t.Run(variant, func(t *testing.T) {
			command := valid

			command = invalidGuidedQuizAnswer(command, variant)
			if _, err := operations.ConductGuidedQuiz(t.Context(), command); err == nil {
				t.Fatal("invalid answer accepted")
			}

			saved, err := core.ReadGuidedUseCase(t.Context(), operations.projectContext.ProjectDir, valid.GuidedUseCaseScope)
			if err != nil || saved.Revision != current.Revision || saved.Guided.PendingOperation != nil {
				t.Fatalf("invalid answer reserved effects: %v", err)
			}
		})
	}
}

// TestGuidedUseCaseQuizLostAck reconciles the quiz owner's receipt before repeating generation.
func TestGuidedUseCaseQuizLostAck(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")
	lost := errors.New("lost quiz acknowledgement")
	operations.afterGuidedQuiz = func() error { return lost }

	command := guidedQuizCommand(session, "start", "quiz", "")
	if _, err := operations.ConductGuidedQuiz(t.Context(), command); !errors.Is(err, lost) {
		t.Fatalf("fault not reached: %v", err)
	}

	operations.afterGuidedQuiz = nil

	recovered, err := operations.ConductGuidedQuiz(t.Context(), command)
	if err != nil || recovered.Guided.Quiz == nil || len(model.requests) != 1 {
		t.Fatalf("quiz operation repeated: %v", err)
	}
}

// TestGuidedUseCaseQuizInvalid does not turn empty or failed generation into clarification.
func TestGuidedUseCaseQuizInvalid(t *testing.T) {
	for _, variant := range []string{"empty", "premature"} {
		t.Run(variant, func(t *testing.T) {
			operations, session, model := guidedQuizFixture(t, "default")
			if variant == "empty" {
				model.responses[0].Question.Prompt = ""
			} else {
				model.responses[0].Question.Required = false
			}

			blocked, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
			if err == nil || blocked.Guided.Stage != "quiz" || len(blocked.Guided.Checkpoints) != 1 {
				t.Fatalf("invalid generation advanced: %v", err)
			}
		})
	}
}

func guidedQuizFixture(t *testing.T, namespace string) (service, core.AuthoringSession, *fakeServiceQuizModel) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git guided quiz fixture")
	}

	provider := webCaseCreateProvider(t)
	root := provider.projectContext.ProjectDir
	initializeGuidedWebGit(t, root)

	operations := newTestService(t, root)
	scope := core.GuidedUseCaseScope{RunID: "guided-quiz", ProjectID: operations.projectContext.ProjectID, Namespace: namespace}

	session, err := core.StartGuidedUseCase(t.Context(), root, core.GuidedUseCaseStart{GuidedUseCaseScope: scope, Actor: "operator", ProgressionIntent: true})
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range []struct{ action, content string }{{"name", "Guided clarification"}, {"append", "Deliver a bounded workflow."}} {
		session, err = core.MutateGuidedUseCase(t.Context(), root, core.GuidedUseCaseCommand{GuidedUseCaseScope: scope, OperationID: input.action, ExpectedRevision: session.Revision, Action: input.action, Content: input.content})
		if err != nil {
			t.Fatal(err)
		}
	}

	session, err = core.FinishGuidedUseCaseDump(t.Context(), root, core.GuidedUseCaseCommand{GuidedUseCaseScope: scope, OperationID: "finish", ExpectedRevision: session.Revision, Action: "finish"})
	if err != nil {
		t.Fatal(err)
	}

	now := serviceQuizClock()()
	question := serviceQuizModelResponse("question-response", "question-1", now)
	completion := serviceQuizModelResponse("completion-response", "complete", now)
	completion.Question.Required = false
	model := &fakeServiceQuizModel{responses: []core.QuizModelResponse{question, completion}}
	operations.quizModel = model

	return operations, session, model
}

func guidedQuizCommand(session core.AuthoringSession, id, action, content string) core.GuidedUseCaseCommand {
	command := core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, ExpectedRevision: session.Revision, OperationID: id, Action: action, Content: content}
	if session.Guided.Quiz != nil {
		command.QuestionID, command.InputDigest = session.Guided.Quiz.QuestionID, session.Guided.Quiz.InputDigest
	}

	return command
}

// TestGuidedUseCaseQuizManual rejects adoption of a manually started quiz in this case.
func TestGuidedUseCaseQuizManual(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")

	scope := core.QuizScope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}
	if _, err := operations.StartQuiz(QuizStartRequest{Scope: scope}); err != nil {
		t.Fatal(err)
	}

	blocked, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start-guided", "quiz", ""))
	if err == nil || blocked.Guided.Quiz != nil || len(model.requests) != 1 {
		t.Fatalf("adopted or replaced manual quiz: %v", err)
	}
}

// TestGuidedUseCaseQuizConcurrent accepts one current answer across simultaneous clients.
func TestGuidedUseCaseQuizConcurrent(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)

	var wait sync.WaitGroup
	for _, id := range []string{"client-one", "client-two"} {
		wait.Add(1)
		go func() {
			defer wait.Done()

			_, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, id, "answer", "Retain the agreed bounded scope."))
			results <- err
		}()
	}

	wait.Wait()
	close(results)

	successes := 0

	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, core.ErrWorkflowConflict) {
			t.Fatal(err)
		}
	}

	if successes != 1 || len(model.requests) != 2 {
		t.Fatalf("concurrent clients repeated the quiz: successes=%d requests=%d", successes, len(model.requests))
	}
}

func replaceGuidedWorkingDump(t *testing.T, root string, dump core.GuidedStageCheckpoint) {
	t.Helper()

	for _, artifact := range dump.Artifacts {
		if strings.HasSuffix(artifact.Path, "/artifact.md") {
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(artifact.Path)), []byte("Later replacement"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func invalidGuidedQuizAnswer(command core.GuidedUseCaseCommand, variant string) core.GuidedUseCaseCommand {
	switch variant {
	case "question":
		command.QuestionID = "other"
	case "source":
		command.InputDigest = "other"
	case "revision":
		command.ExpectedRevision--
	case "empty":
		command.Content = " "
	case "run":
		command.RunID = "another-run"
	}

	return command
}

func assertGuidedQuizCompletion(t *testing.T, completed core.AuthoringSession, dump core.GuidedStageCheckpoint) {
	t.Helper()

	if completed.Guided.Stage != "specification" || len(completed.Guided.Checkpoints) != 2 {
		t.Fatalf("quiz was not sealed: %#v", completed.Guided)
	}

	checkpoint := completed.Guided.Checkpoints[1]
	if checkpoint.Stage != "quiz" || len(checkpoint.Sources) == 0 || checkpoint.Sources[0].Commit != dump.Commit {
		t.Fatal("quiz lost its committed dependency")
	}
}
