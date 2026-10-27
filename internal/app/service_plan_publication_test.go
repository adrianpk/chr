package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type planPublicationTestModel func(context.Context, quizmodel.Request) (quizmodel.Response, error)

func (f planPublicationTestModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	return f(ctx, request)
}

// Exercise the shared operation with real source owners. Only semantic/model
// boundaries are substituted; length and requiring both sources are not gates.
func TestPlanGenerationDiscovery(t *testing.T) {
	for _, item := range []struct {
		name, dump, answer, status string
		sufficient                 bool
	}{
		{name: "no discovery", status: "absent"},
		{name: "placeholder", dump: "# Context\n\nTODO", status: "absent"},
		{name: "unrelated general dump", dump: "The cafeteria serves lunch.", status: "insufficient"},
		{name: "insufficient", dump: "Improve publishing.", status: "insufficient"},
		{name: "complete dump only", dump: "Publish local Markdown after explicit editorial approval. Reject duplicate slugs; retain drafts and render readable HTML. Verify approval prevents premature publication and exported HTML matches approved text.", sufficient: true},
		{name: "short narrow dump", dump: "Print hello once. Verify stdout is hello.", sufficient: true},
		{name: "quiz only", answer: "Print hello once. Verify stdout is hello.", sufficient: true},
	} {
		t.Run(item.name, func(t *testing.T) {
			operations, scope, _ := planPublicationFixture(t, item.dump)
			if item.answer != "" {
				appendPlanTestAnswer(t, operations, scope, item.answer, quizcore.EntryStatusAnswered)
			}

			calls := 0
			operations.conversationModel = planReadinessTestProvider(func(ctx context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
				calls++

				if item.sufficient {
					return sufficientPlanReadinessProvider()(ctx, request)
				}

				return insufficientPlanTestAssessment(ctx, t, request), nil
			})
			_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{Scope: scope})
			checkPlanDiscoveryResult(t, err, item.status, item.sufficient, calls)
		})
	}
}

func checkPlanDiscoveryResult(t *testing.T, err error, status string, sufficient bool, calls int) {
	t.Helper()

	if sufficient {
		if err != nil || calls != 1 {
			t.Fatal("usable discovery rejected", err, calls)
		}

		return
	}

	var rejected *PlanReadinessError
	if !errors.As(err, &rejected) || rejected.Decision.Status != status {
		t.Fatal("missing readiness failure", status, err)
	}

	if status == "absent" && calls != 0 {
		t.Fatal("absent discovery invoked semantic assessment")
	}
}

// Failed preparation and intervening edits never publish an archive or replace
// the last usable plan. Source writes happen during the substituted model call,
// proving the model runs outside the shared mutation lock.
func TestPlanPublicationFailures(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	for _, failure := range []string{"insufficient", "invalid assessment", "assessment failure", "model failure", "invalid plan", "cancelled", "source changed", "target changed"} {
		t.Run(failure, func(t *testing.T) {
			operations, scope, dump := planPublicationFixture(t, "Print hello once. Verify stdout is hello.")
			paths := seedPlanPublicationTarget(t, operations, scope)
			before := readServiceWorkspaceFile(t, paths.Path)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			configurePlanPublicationFailure(t, &operations, failure, dump, paths.Path, cancel)

			_, err := operations.RegeneratePlan(ctx, PlanRegenerateRequest{Scope: scope})
			if err == nil {
				t.Fatal("failed/stale operation succeeded")
			}

			after := readServiceWorkspaceFile(t, paths.Path)

			if failure == "target changed" {
				before += "\nIntervening manual edit.\n"
			}

			if after != before {
				t.Fatal("failed regeneration replaced the current plan")
			}

			archives, globErr := filepath.Glob(filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "plan", "00000-blog-plan-*"))
			if globErr != nil || len(archives) != 0 {
				t.Fatal("failed regeneration published an archive", archives, globErr)
			}
		})
	}
}

func planPublicationFixture(t *testing.T, dump string) (service, quizcore.Scope, string) {
	t.Helper()

	project := t.TempDir()
	writeServiceWorkspaceFile(t, authoringcore.ProjectMetadataPath(project), authoringcore.RenderProjectMetadata(authoringcore.ProjectMetadata{Name: "Greeting", Description: "Deliver a bounded greeting command."}))
	writeServiceWorkspaceFile(t, authoringcore.CaseMetadataPath(project, "blog"), authoringcore.RenderCaseMetadata(authoringcore.CaseMetadata{Name: "Greeting command", Description: "Print one greeting to stdout."}))

	reference := "ops/default/dump/00000-context/artifact.md"
	if dump != "" {
		writeServiceWorkspaceFile(t, filepath.Join(project, reference), "---\ntitle: Context\n---\n\n"+dump+"\n")
	}

	operations := newTestService(t, project)
	operations.now, operations.quizModel = serviceQuizClock(), &fakeServiceQuizModel{}
	operations.conversationModel = sufficientPlanReadinessProvider()

	return operations, quizcore.Scope{ProjectID: operations.projectContext.ProjectID, CaseSlug: "blog"}, reference
}

func seedPlanPublicationTarget(t *testing.T, operations service, scope quizcore.Scope) authoringcore.GeneratedPlanPaths {
	t.Helper()

	paths, err := authoringcore.WriteGeneratedPlan(t.Context(), operations.projectContext.ProjectDir, generationcore.GeneratedPlan{Scope: scope, Title: "Retained plan", Body: "Keep this usable until a complete replacement is ready.", CreatedAt: operations.utcNow()})
	if err != nil {
		t.Fatal(err)
	}

	return paths
}

func appendPlanTestAnswer(t *testing.T, operations service, scope quizcore.Scope, answer string, status quizcore.EntryStatus) {
	t.Helper()

	state, err := operations.newQuizRuntimeState(scope, quizcore.ModeDeep)
	if err != nil {
		t.Fatal(err)
	}

	entry := quizcore.Entry{ID: "entry-1", QuestionID: "question-1", Question: "What should this case deliver and how is it verified?", Answer: answer, Status: status, Mode: quizcore.ModeDeep, Topic: "scope", Depth: 1, CreatedAt: operations.utcNow(), AnsweredAt: operations.utcNow()}
	if _, err := operations.quizRepository().AppendRawCapture(t.Context(), state, entry); err != nil {
		t.Fatal(err)
	}
}

func insufficientPlanTestAssessment(ctx context.Context, t *testing.T, request core.StructuredModelRequest) json.RawMessage {
	t.Helper()

	output, err := sufficientPlanReadinessProvider()(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	return json.RawMessage(strings.Replace(strings.Replace(string(output), `"sufficient":true`, `"sufficient":false`, 1), `"missing":[]`, `"missing":["Define the intended outcome and its verification"]`, 1))
}

func configurePlanPublicationFailure(t *testing.T, operations *service, failure, dump, plan string, cancel context.CancelFunc) {
	t.Helper()

	switch failure {
	case "insufficient":
		operations.conversationModel = planReadinessTestProvider(func(ctx context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
			return insufficientPlanTestAssessment(ctx, t, request), nil
		})
	case "invalid assessment":
		operations.conversationModel = planReadinessTestProvider(func(context.Context, core.StructuredModelRequest) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		})
	case "assessment failure":
		operations.conversationModel = planReadinessTestProvider(func(context.Context, core.StructuredModelRequest) (json.RawMessage, error) {
			return nil, errors.New("provider unavailable")
		})
	default:
		operations.quizModel = planPublicationTestModel(func(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
			applyPlanTestInterruption(ctx, t, *operations, failure, dump, plan, cancel)

			if failure == "model failure" {
				return quizmodel.Response{}, errors.New("model unavailable")
			}

			response, err := (&fakeServiceQuizModel{}).Invoke(ctx, request)
			if failure == "invalid plan" {
				response.Plan.Body, response.PlanOutline.Title = "", ""
			}

			return response, err
		})
	}
}

func applyPlanTestInterruption(ctx context.Context, t *testing.T, operations service, failure, dump, plan string, cancel context.CancelFunc) {
	t.Helper()

	switch failure {
	case "cancelled":
		cancel()
	case "source changed":
		observed, err := authoringcore.ReadDumpDocument(ctx, operations.projectContext.ProjectDir, dump)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := authoringcore.MutateDump(ctx, operations.projectContext.ProjectDir, authoringcore.DumpMutationRequest{Action: "save", OperationID: "source-change", Reference: dump, ExpectedRevision: observed.Revision, Title: "Context", Content: "Changed the result to goodbye."}, nil); err != nil {
			t.Fatal(err)
		}
	case "target changed":
		file, err := os.OpenFile(plan, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}

		_, writeErr := file.WriteString("\nIntervening manual edit.\n")
		if err := errors.Join(writeErr, file.Close()); err != nil {
			t.Fatal(err)
		}
	}
}
