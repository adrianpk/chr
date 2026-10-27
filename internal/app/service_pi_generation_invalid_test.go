package app

import (
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Real Quiz, plan and conversational parsers reject selected Pi output without fallback.
func TestPiGenerationInvalidCallers(t *testing.T) {
	for _, family := range []string{"quiz", "plan", "conversation"} {
		t.Run(family, func(t *testing.T) {
			operations, driver := piGenerationFixture(t, "not-json")
			project := operations.projectContext.ProjectDir
			writeServiceQuizFixture(t, project)

			var err error

			switch family {
			case "quiz":
				_, err = operations.StartQuiz(QuizStartRequest{Mode: quizcore.QuizModeDeep})
			case "plan":
				harnessSelectionWrite(t, filepath.Join(driver, "plan-readiness"), "enabled")

				_, err = operations.GeneratePlan(t.Context(), PlanGenerateRequest{OperationID: "invalid-plan", Scope: quizcore.QuizScope{CaseSlug: "blog"}})
			case "conversation":
				bound, ctx, bindErr := operations.bindGeneration(t.Context(), "invalid-conversation", "conversation")
				if bindErr != nil {
					t.Fatal(bindErr)
				}

				_, err = bound.conversationModel.InvokeStructuredModel(ctx, piCapacityRequest(project, "invalid"))
			}

			if err == nil {
				t.Fatal("invalid selected Pi output accepted")
			}

			valid := 0
			if family == "plan" {
				valid = 1
			}

			checkPiGenerationMetadata(t, project, valid, 1)
		})
	}
}

// An invalid tracking result cannot create tracking material or tickets.
func TestPiGenerationTrackingInvalid(t *testing.T) {
	operations, _ := piGenerationFixture(t, "not-json")
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)

	path := filepath.Join(project, ".chr", "default", "plans", "blog-plan.md")
	writeServiceWorkspaceFile(t, path, "---\ntitle: Blog Plan\nstatus: planned\ncase_slug: blog\n---\n\n# Blog Plan\n\n## Slice 1\n\nImplement planning foundation.\n")

	path = canonicalServiceWorkspaceFixturePath(path)

	relative, err := filepath.Rel(project, path)
	if err != nil {
		t.Fatal(err)
	}

	if err := core.EnsureOpsArtifactBodyMetadata(project, filepath.ToSlash(relative)); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{OperationID: "invalid-tracking", Scope: quizcore.QuizScope{CaseSlug: "blog"}}); err == nil {
		t.Fatal("invalid Pi tracking accepted")
	}

	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) {
		t.Fatal("invalid tracking changed the canonical plan", err)
	}

	checkPiGenerationMetadata(t, project, 0, 1)
}
