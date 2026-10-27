package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// An answer commits even when its continuation fails. Resume after restart uses
// that answer's binding; the next independent answer reads the current role.
func TestHarnessGenerationBindingQuiz(t *testing.T) {
	generationBindingQuizCase(t, harnessGenerationFixture)
}

func generationBindingQuizCase(t *testing.T, fixture func(*testing.T, string) (service, string)) {
	operations, driverDir := fixture(t, generationQuestionFixture("Which behavior must remain?"))
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.QuizModeDeep})
	if err != nil || started.State.ActiveQuestion == nil {
		t.Fatal("selected driver could not generate a quiz question", err)
	}

	switchGenerationFixture(t, operations, "other")
	harnessSelectionWrite(t, filepath.Join(driverDir, "result.json"), generationQuestionFixture("How should that behavior be checked?"))
	harnessSelectionWrite(t, filepath.Join(driverDir, "fail-next"), "owned failure")

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Preserve existing ownership and retry limits."})
	if err != nil || !answered.State.ContinuationPending || answered.State.AnswersRecorded != 1 {
		t.Fatal("provider failure discarded the answer", err)
	}

	next := resumeGenerationFixture(t, operations, driverDir, answered.State)
	checkGenerationFixtureModels(t, project, map[string]string{started.State.GenerationOperationID: "fixture-model", answered.State.GenerationOperationID: "other-model", next: "fixture-model"})
}

func resumeGenerationFixture(t *testing.T, operations service, driverDir string, pending quizcore.QuizRuntimeState) string {
	t.Helper()
	switchGenerationFixture(t, operations, "writing")

	profiles := filepath.Join(operations.env.ConfigDir(), "harnesses.yml")

	data, err := os.ReadFile(profiles)
	if err != nil {
		t.Fatal(err)
	}

	harnessName := "codex"
	if strings.Contains(string(data), "harness: pi") {
		harnessName = "pi"
	}

	original := "  other:\n    harness: " + harnessName + "\n    command: " + filepath.Join(driverDir, "driver")
	harnessSelectionWrite(t, profiles, strings.Replace(string(data), original, "  other:\n    harness: codex\n    command: /unavailable-fixture", 1))

	restarted := newService(t.Context(), operations.env, operations.projectContext, testAssetsFS())

	resumed, err := restarted.ResumeQuiz(QuizResumeRequest{})
	if err != nil || resumed.State.ContinuationPending || resumed.State.ActiveQuestion == nil || resumed.State.GenerationOperationID != pending.GenerationOperationID {
		t.Fatal("resumed continuation rebound its profile", err)
	}

	return independentQuizGenerationFixture(t, restarted, driverDir, pending)
}

func independentQuizGenerationFixture(t *testing.T, restarted service, driverDir string, pending quizcore.QuizRuntimeState) string {
	t.Helper()
	harnessSelectionWrite(t, filepath.Join(driverDir, "result.json"), generationQuestionFixture("Which acceptance check is still missing?"))

	next, err := restarted.AnswerQuiz(QuizAnswerRequest{Answer: "Use the focused regression boundary."})
	if err != nil || next.State.GenerationOperationID == pending.GenerationOperationID || next.State.ContinuationPending {
		t.Fatal("new independent answer retained old selection", err)
	}

	records := generationFixtureMetadata(t, restarted.projectContext.ProjectDir)
	for _, record := range records {
		if record.OperationID == next.State.GenerationOperationID && record.Selection.Selection.Profile == "writing" {
			return next.State.GenerationOperationID
		}
	}

	t.Fatal("new answer did not use the current profile")

	return ""
}

func generationQuestionFixture(question string) string {
	return `{"question":"` + question + `","tags":["scope"],"follow_up_depth":1,"completion_recommended":false}`
}
