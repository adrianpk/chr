package generation

import (
	"sort"
)

// OrderedWorkflowPlan orders a copy without changing the caller-owned plan.
func OrderedWorkflowPlan(plan GeneratedWorkflowPlan) GeneratedWorkflowPlan {
	plan.Slices = append([]GeneratedWorkflowSlice(nil), plan.Slices...)
	sort.Slice(plan.Slices, func(i, j int) bool { return plan.Slices[i].Order < plan.Slices[j].Order })

	for i := range plan.Slices {
		plan.Slices[i].Tasks = append([]GeneratedWorkflowTask(nil), plan.Slices[i].Tasks...)
		sort.Slice(plan.Slices[i].Tasks, func(a, b int) bool { return plan.Slices[i].Tasks[a].Order < plan.Slices[i].Tasks[b].Order })
	}

	return plan
}
