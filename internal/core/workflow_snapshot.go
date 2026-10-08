package core

import (
	"fmt"
	"sort"
)

// Validate checks shared hierarchy and uniqueness independently of SQLite.
func (s WorkflowSnapshot) Validate() error {
	if len(s.Records) > WorkflowRecordLimit {
		return fmt.Errorf("workflow namespace record limit exceeded")
	}

	ids := map[string]bool{}
	sets := map[string]WorkflowRecord{}
	sliceOwners := map[string]string{}
	graph := map[WorkflowRef][]WorkflowRef{}

	for _, r := range s.Records {
		if err := r.Validate(); err != nil {
			return err
		}

		if ids[r.ID] {
			return fmt.Errorf("duplicate workflow artifact identity: %s", r.ID)
		}

		ids[r.ID] = true

		if r.Set != nil {
			sets[r.ID] = r

			if err := addWorkflowSlices(r, sliceOwners, graph); err != nil {
				return err
			}
		}
	}

	if err := s.validateTasks(sets, sliceOwners, graph); err != nil {
		return err
	}

	if err := validateWorkflowSetOrders(sets); err != nil {
		return err
	}

	return validateWorkflowGraph(graph)
}

func addWorkflowSlices(r WorkflowRecord, owners map[string]string, graph map[WorkflowRef][]WorkflowRef) error {
	ordered := append([]WorkflowSlice(nil), r.Set.Slices...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Order < ordered[j].Order })

	for i, slice := range ordered {
		if owners[slice.ID] != "" {
			return fmt.Errorf("duplicate scoped workflow slice: %s", slice.ID)
		}

		owners[slice.ID] = r.ID
		ref := WorkflowRef{Kind: "slice", ID: slice.ID}
		graph[ref] = append([]WorkflowRef(nil), slice.DependsOn...)

		if i > 0 {
			graph[ref] = append(graph[ref], WorkflowRef{Kind: "slice", ID: ordered[i-1].ID})
		}
	}

	return nil
}

func (s WorkflowSnapshot) validateTasks(sets map[string]WorkflowRecord, owners map[string]string, graph map[WorkflowRef][]WorkflowRef) error {
	orders := map[string]map[int]bool{}

	for _, r := range s.Records {
		if r.ProjectID != s.Records[0].ProjectID || r.Namespace != s.Records[0].Namespace {
			return fmt.Errorf("foreign workflow record scope")
		}

		if r.Task == nil {
			continue
		}

		task := r.Task

		ref := WorkflowRef{Kind: "ticket", ID: task.Key}
		if _, exists := graph[ref]; exists {
			return fmt.Errorf("duplicate scoped workflow ticket key: %s", task.Key)
		}

		graph[ref] = append([]WorkflowRef(nil), task.DependsOn...)
		if task.SetID == "" {
			continue
		}

		if err := validateWorkflowTaskOwner(*task, sets, owners, orders); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowTaskOwner(task WorkflowTask, sets map[string]WorkflowRecord, owners map[string]string, orders map[string]map[int]bool) error {
	set, exists := sets[task.SetID]
	if !exists || owners[task.SliceID] != task.SetID || (task.UseCase != "" && task.UseCase != set.Set.UseCase) {
		return fmt.Errorf("foreign or unresolved workflow task association")
	}

	if orders[task.SliceID] == nil {
		orders[task.SliceID] = map[int]bool{}
	}

	if orders[task.SliceID][task.Order] {
		return fmt.Errorf("duplicate workflow task order")
	}

	orders[task.SliceID][task.Order] = true

	return nil
}

func validateWorkflowSetOrders(sets map[string]WorkflowRecord) error {
	orders := map[string]map[int]bool{}

	for _, r := range sets {
		if r.Set.UseCase == "" {
			continue
		}

		seen := orders[r.Set.UseCase]
		if seen == nil {
			seen = map[int]bool{}
			orders[r.Set.UseCase] = seen
		}

		order := r.Set.UseCaseOrder
		if invalidWorkflowSetOrder(order, seen) {
			return fmt.Errorf("mixed or duplicate workflow use-case set order")
		}

		seen[order] = true
	}

	return nil
}

func invalidWorkflowSetOrder(order int, seen map[int]bool) bool {
	if order == 0 {
		return len(seen) > 0 && !seen[0]
	}

	return seen[0] || seen[order]
}

func validateWorkflowGraph(graph map[WorkflowRef][]WorkflowRef) error {
	state := map[WorkflowRef]int{}

	var visit func(WorkflowRef) error

	visit = func(ref WorkflowRef) error {
		if state[ref] == 1 {
			return fmt.Errorf("cyclic workflow dependency: %s", ref.ID)
		}

		if state[ref] == 2 {
			return nil
		}

		deps, exists := graph[ref]
		if !exists {
			return fmt.Errorf("unresolved workflow dependency: %s", ref.ID)
		}

		state[ref] = 1

		for _, dep := range deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		state[ref] = 2

		return nil
	}
	for ref := range graph {
		if err := visit(ref); err != nil {
			return err
		}
	}

	return nil
}
