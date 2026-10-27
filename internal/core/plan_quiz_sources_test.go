package core

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Planning receives every answered question without replaying routing metadata.
// Observing the compact input must leave the original history and images intact.
func TestPlanQuizSourceContext(t *testing.T) {
	project := t.TempDir()
	state := planQuizSourceFixture(t, project, 30, "Preserve saved tasks.")
	before := projectDumpTestImage(t, project)
	scope := quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}

	observed, err := ObservePlanSources(t.Context(), project, scope)
	if err != nil {
		t.Fatal("internal metadata consumed the planning context budget", err)
	}

	if len(observed.Quiz.Entries) != 30 || !strings.Contains(observed.Quiz.Entries[0].PromptIntent, "INTERNAL_ROUTING") {
		t.Fatal("the retained quiz observation lost history")
	}

	checkPlanQuizSourceAnswers(t, observed)

	if !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("planning rewrote saved discovery")
	}

	checkPlanQuizSourceRevision(t, project, state, observed.Hash)
}

func planQuizSourceFixture(t *testing.T, project string, count int, answer string) quizcore.RuntimeState {
	t.Helper()

	writeTestFile(t, ProjectMetadataPath(project), RenderProjectMetadata(ProjectMetadata{Name: "Tasks", Description: "Manage local tasks."}))
	writeTestFile(t, CaseMetadataPath(project, "tasks"), RenderCaseMetadata(CaseMetadata{Name: "Task management", Description: "Save and list tasks."}))

	state := validQuizRuntimeState()
	state.CaseSlug = "tasks"
	repository := NewQuizArtifactRepository(project)

	for i := range count {
		entry := validQuizEntry()
		entry.ID, entry.QuestionID = fmt.Sprintf("entry-%d", i), fmt.Sprintf("question-%d", i)
		entry.Question, entry.Answer = fmt.Sprintf("What is requirement %d?", i), answer
		entry.PromptIntent = `{"kind":"quiz_question","metadata":{"routing_state":"` + strings.Repeat("INTERNAL_ROUTING", 300) + `"}}`

		if _, err := repository.AppendRawCapture(t.Context(), state, entry); err != nil {
			t.Fatal(err)
		}
	}

	return state
}

func checkPlanQuizSourceAnswers(t *testing.T, observed PlanSourceObservation) {
	t.Helper()

	input := planQuizSourceInput(t, observed)

	var answers []map[string]string
	if err := json.Unmarshal([]byte(input.Content), &answers); err != nil {
		t.Fatal(err)
	}

	if len(answers) != 30 || strings.Contains(input.Content, "INTERNAL_ROUTING") || input.Reference == "" || input.Revision == "" {
		t.Fatal("planning lost answers or leaked internal routing")
	}

	for i, answer := range answers {
		if answer["question"] != fmt.Sprintf("What is requirement %d?", i) || answer["answer"] != "Preserve saved tasks." {
			t.Fatal("planning changed question/answer order or content", i)
		}
	}
}

func planQuizSourceInput(t *testing.T, observed PlanSourceObservation) PlanSourceInput {
	t.Helper()

	for _, input := range observed.Inputs {
		if input.Kind == "quiz" {
			return input
		}
	}

	t.Fatal("planning omitted quiz discovery")

	return PlanSourceInput{}
}

func checkPlanQuizSourceRevision(t *testing.T, project string, state quizcore.RuntimeState, previous string) {
	t.Helper()

	entry := validQuizEntry()
	entry.ID, entry.QuestionID, entry.Status, entry.Answer = "skip", "skip", quizcore.EntryStatusSkipped, ""

	if _, err := NewQuizArtifactRepository(project).AppendRawCapture(t.Context(), state, entry); err != nil {
		t.Fatal(err)
	}

	observed, err := ObservePlanSources(t.Context(), project, quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug})
	if err != nil || observed.Hash == previous {
		t.Fatal("the compact payload lost full-source revision binding", err)
	}

	checkPlanQuizSourceAnswers(t, observed)
}

// Real discovery size remains bounded; compacting metadata cannot silently
// truncate an oversized user answer or bypass the planning context limit.
func TestPlanQuizSourceLimit(t *testing.T) {
	project := t.TempDir()
	state := planQuizSourceFixture(t, project, 1, strings.Repeat("x", 96*1024))

	_, err := ObservePlanSources(t.Context(), project, quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug})
	if err == nil || !strings.Contains(err.Error(), "plan source context exceeds 96 KiB") {
		t.Fatal("oversized user content was truncated or accepted", err)
	}
}
