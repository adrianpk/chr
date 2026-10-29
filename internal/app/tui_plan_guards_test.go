package app

import (
	"os"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Unbound writes are rejected at the terminal boundary before provider work.
func TestTUIPlanBinding(t *testing.T) {
	for _, field := range []string{"operation", "revision"} {
		t.Run(field, func(t *testing.T) {
			provider, generator := webPlanDeliveryFixture(t)
			operations := provider.service(t.Context())

			request := tuiPlanDeliveryRequest(t, operations, "bound-operation")
			if field == "operation" {
				request.OperationID = ""
			} else {
				request.Revision = ""
			}

			if _, err := newTUIPlanActionExecutor(operations).ExecutePlanAction(request); err == nil || len(generator.base.requests) != 0 || generator.calls != 0 {
				t.Fatal("unbound terminal mutation reached generation", err)
			}
		})
	}
}

// Explicit replacement carries its captured target revision to the same
// publication owner. Concurrent source/target changes retain the usable plan.
func TestTUIPlanRegenerate(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	for _, change := range []string{"unchanged", "target changed", "source changed"} {
		t.Run(change, func(t *testing.T) { runTUIPlanRegenerate(t, change) })
	}
}

func runTUIPlanRegenerate(t *testing.T, change string) {
	t.Helper()

	operations, scope, dump := planPublicationFixture(t, "Print hello once. Verify stdout is hello.")
	paths := seedPlanPublicationTarget(t, operations, scope)

	revision, err := authoringcore.PlanTargetRevision(t.Context(), operations.projectContext.ProjectDir, scope)
	if err != nil {
		t.Fatal(err)
	}

	before := readServiceWorkspaceFile(t, paths.Path)
	if change == "target changed" {
		before += "\nConcurrent target.\n"
		if err := os.WriteFile(paths.Path, []byte(before), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if change == "source changed" {
		configurePlanPublicationFailure(t, &operations, change, dump, paths.Path, func() {})
	}

	request := tui.PlanActionRequest{Action: tui.PlanActionRegenerate, Scope: tui.QuizScope{CaseSlug: scope.CaseSlug}, OperationID: "replace-terminal-plan", Revision: revision}

	result, err := newTUIPlanActionExecutor(operations).ExecutePlanAction(request)
	if change == "unchanged" {
		if err != nil || result.DocumentPath == "" {
			t.Fatal(result, err)
		}
	} else if err == nil || readServiceWorkspaceFile(t, paths.Path) != before {
		t.Fatal("stale replacement changed the usable plan", result, err)
	}
}
