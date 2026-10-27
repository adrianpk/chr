package generation

// PlannedWorkflowOrder projects stable slice and task identities without allocation.
func PlannedWorkflowOrder(plan GeneratedWorkflowPlan) []PlanOrderSlice {
	result := make([]PlanOrderSlice, 0, len(plan.Slices))
	for _, slice := range plan.Slices {
		item := PlanOrderSlice{ID: slice.ID, Order: slice.Order, Title: slice.Title}
		for _, task := range slice.Tasks {
			item.Tasks = append(item.Tasks, PlanOrderTask{ID: task.ID, Order: task.Order, Title: task.Title})
		}

		result = append(result, item)
	}

	return result
}
