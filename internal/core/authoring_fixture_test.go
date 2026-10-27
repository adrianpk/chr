package core

import (
	"fmt"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func orderedPlanCandidate() generationcore.GeneratedPlan {
	plan := planPublicationCandidate()
	plan.RequireOrderMap = true
	plan.Order = []generationcore.PlanOrderSlice{{ID: "slice-1", Order: 1, Title: "Foundation", Tasks: []generationcore.PlanOrderTask{{ID: "task-t1.1", Order: 1, Title: "Implement foundation"}}}}

	return plan
}

func readTestPlanBinding(t *testing.T, project string, scope quizcore.Scope) workflowstate.PlanRevisionBinding {
	t.Helper()

	binding, err := authoringcore.ObservePlanRevision(t.Context(), project, scope)
	if err != nil {
		t.Fatal(err)
	}

	return binding
}

func planPublicationCandidate() generationcore.GeneratedPlan {
	return generationcore.GeneratedPlan{Scope: quizcore.Scope{ProjectID: "example", CaseSlug: "chosen"}, Title: "Original plan", Body: "Original bounded work.", SourceObservationHash: "observed", SourceInputs: []generationcore.PlanSourceBinding{{Kind: "dump", Reference: "ops/default/dump/00000-context/artifact.md", Revision: "observed-source"}}, CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
}

// workflowPlanFixture represents two integration units with three task contracts
// each. Display labels are intentionally separate from durable input identities.
func workflowPlanFixture() generationcore.GeneratedWorkflowPlan {
	gate := workflowstate.Validation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}

	plan := generationcore.GeneratedWorkflowPlan{MaterializationID: "generation-a", Title: "Delivery", PlanRef: "plan/approved.md", PlanHash: contenthash.Sum([]byte("approved\n")), ClosingValidation: gate}
	for i := 1; i <= 2; i++ {
		slice := generationcore.GeneratedWorkflowSlice{ID: fmt.Sprintf("slice-%d", i), Label: fmt.Sprintf("Slice %d", i), Order: i, Title: "Integration", Contract: "Deliver one reviewed integration unit", Validation: gate}
		for j := 1; j <= 3; j++ {
			slice.Tasks = append(slice.Tasks, generationcore.GeneratedWorkflowTask{ID: fmt.Sprintf("task-%d-%d", i, j), Label: fmt.Sprintf("T%d.%d", i, j), Order: j, Title: "Task", Body: "Original task instructions", Outcome: "Deliver behavior", Scope: "Bounded task", Acceptance: []string{"Observed result"}, Validation: gate, Priority: "normal", Severity: "normal", Type: "task"})
		}

		plan.Slices = append(plan.Slices, slice)
	}

	return plan
}
