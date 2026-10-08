package core

import (
	"encoding/json"
	"fmt"
	"testing"
)

// workflowPlanFixture represents two integration units with three task contracts
// each. Display labels are intentionally separate from durable input identities.
func workflowPlanFixture() GeneratedWorkflowPlan {
	gate := WorkflowValidation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}

	plan := GeneratedWorkflowPlan{MaterializationID: "generation-a", Title: "Delivery", PlanRef: "plan/approved.md", PlanHash: workflowHash([]byte("approved\n")), ClosingValidation: gate}
	for i := 1; i <= 2; i++ {
		slice := GeneratedWorkflowSlice{ID: fmt.Sprintf("slice-%d", i), Label: fmt.Sprintf("Slice %d", i), Order: i, Title: "Integration", Contract: "Deliver one reviewed integration unit", Validation: gate}
		for j := 1; j <= 3; j++ {
			slice.Tasks = append(slice.Tasks, GeneratedWorkflowTask{ID: fmt.Sprintf("task-%d-%d", i, j), Label: fmt.Sprintf("T%d.%d", i, j), Order: j, Title: "Task", Body: "Original task instructions", Outcome: "Deliver behavior", Scope: "Bounded task", Acceptance: []string{"Observed result"}, Validation: gate, Priority: "normal", Severity: "normal", Type: "task"})
		}

		plan.Slices = append(plan.Slices, slice)
	}

	return plan
}

// TestWorkflowPlanHierarchy rejects ambiguity before persistence and permits
// label changes without replacing set, slice or task request identities.
func TestWorkflowPlanHierarchy(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*GeneratedWorkflowPlan)
		invalid bool
	}{
		{name: "six tasks"},
		{name: "renamed labels", mutate: func(p *GeneratedWorkflowPlan) {
			p.Title = "Renamed"
			p.Slices[0].Label = "First"
			p.Slices[0].Tasks[0].Label = "First task"
		}},
		{name: "duplicate slice", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[1].ID = p.Slices[0].ID }, invalid: true},
		{name: "duplicate slice order", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[1].Order = 1 }, invalid: true},
		{name: "duplicate task", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[1].Tasks[0].ID = p.Slices[0].Tasks[0].ID }, invalid: true},
		{name: "duplicate task order", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[0].Tasks[1].Order = 1 }, invalid: true},
		{name: "label identity", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[0].Tasks[0].ID = p.Slices[0].Tasks[0].Label }, invalid: true},
		{name: "empty slice", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[0].Tasks = nil }, invalid: true},
		{name: "missing acceptance", mutate: func(p *GeneratedWorkflowPlan) { p.Slices[0].Tasks[0].Acceptance = nil }, invalid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := workflowPlanFixture()
			if tc.mutate != nil {
				tc.mutate(&plan)
			}

			if err := ValidateGeneratedWorkflowPlan(plan); (err != nil) != tc.invalid {
				t.Fatalf("validation = %v, invalid = %v", err, tc.invalid)
			}

			checkWorkflowPlanRoundTrip(t, plan)
		})
	}
}

func checkWorkflowPlanRoundTrip(t *testing.T, plan GeneratedWorkflowPlan) {
	t.Helper()

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}

	var decoded GeneratedWorkflowPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.MaterializationID != plan.MaterializationID || decoded.Slices[0].ID != plan.Slices[0].ID {
		t.Fatal("round trip replaced identity")
	}
}
