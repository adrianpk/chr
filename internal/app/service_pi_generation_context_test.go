package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// Explanations use their canonical context and reject malformed selected-provider prose.
func TestPiGenerationExplanation(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"explanation":"Retain the configured retry ceiling."}`)
	project := operations.projectContext.ProjectDir
	harnessSelectionWrite(t, filepath.Join(project, "README.md"), "# Prepared project\n\nRetain the configured retry ceiling.\n")

	result, err := operations.ExplainAuthoring(t.Context(), AuthoringExplanationRequest{ID: "pi-explanation", Question: "What retry limit is required?"})
	if err != nil || result.Text == "" || len(result.SourceRefs) != 1 {
		t.Fatal("Pi explanation lost canonical context", err)
	}

	harnessSelectionWrite(t, filepath.Join(driver, "result.json"), `{"explanation":""}`)

	if _, err := operations.ExplainAuthoring(t.Context(), AuthoringExplanationRequest{ID: "pi-explanation-invalid", Question: "What is required?"}); err == nil {
		t.Fatal("invalid Pi explanation accepted")
	}

	checkPiGenerationMetadata(t, project, 1, 1)
}

// The actual contextual Quiz CLI receives Pi through its shared operation binder.
func TestPiGenerationContextualQuiz(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"questions":[{"id":"scope","question":"What behavior must be preserved?","tags":["scope"],"follow_up_depth":0,"completion_recommended":false}]}`)
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)

	devRoot := t.TempDir()

	env, err := newUserEnvironment(devRoot)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range []string{"auth.json", "harnesses.yml"} {
		data, err := os.ReadFile(filepath.Join(operations.env.ConfigDir(), file))
		if err != nil {
			t.Fatal(err)
		}

		harnessSelectionWrite(t, filepath.Join(env.ConfigDir(), file), string(data))
	}

	var stdout bytes.Buffer
	if err := run(t.Context(), []string{"author", "quiz", "new", "--dir", project, "--dev-root", devRoot, "--target", "spec", "--topic", "Prepared scope"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal("contextual Pi Quiz failed", err)
	}

	if len(piGenerationCalls(t, driver)) != 1 {
		t.Fatal("contextual Quiz missed selected Pi runner")
	}

	checkPiGenerationMetadata(t, project, 1, 0)
}

// Review uses the canonical project's selection inside its distinct execution workspace.
func TestPiGenerationReview(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"decision":"needs_changes","summary":"A required check is missing.","blocking_findings":[{"category":"evidence","title":"Missing check","detail":"Run the required focused check."}],"non_blocking_notes":[],"reviewed_artifacts":["result.json"]}`)
	worktree := t.TempDir()
	ticket := testLoopTicketWithPullRequestRef("CHR-1", "Prepared review")
	ticket.Status = core.TicketStatusReadyForReview
	path := writeTicketMarkdownOnly(t, worktree, "CHR-1-Review.md", ticket)

	indexed, err := parseAgenticReviewTicket(t.Context(), operations.projectContext.ProjectID, worktree, path, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	execution := core.ExecutionRecord{
		ID: "pi-review", ProjectID: operations.projectContext.ProjectID, TicketKey: "CHR-1", TicketPath: path,
		Workspace: core.ExecutionWorkspaceRef{Mode: core.WorkWorkspaceModeWorktree, ExecutionWorkspaceDir: worktree},
		Artifacts: core.ExecutionArtifacts{ArtifactID: "00001-review", RunDir: filepath.Join(worktree, "ops", "default", "run", "00001-review"), ResultPath: "result.json"},
	}

	result, err := operations.runAgenticReviewCycle(indexed, execution, core.ResultApplication{TicketPath: path}, core.ReviewWorkflowConfig{})
	if err != nil || result.Status != core.TicketStatusReadyForDev {
		t.Fatal("Pi review failed", err)
	}

	if len(piGenerationCalls(t, driver)) != 1 {
		t.Fatal("review missed selected Pi runner")
	}
}
