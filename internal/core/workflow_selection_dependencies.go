package core

import (
	"cmp"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowBlocker identifies one prerequisite without adding it to the extent.
type WorkflowBlocker struct {
	Dependency      workflowstate.Ref
	Reason          string
	NextActionOwner string
}

type workflowDependencies struct {
	selection     workflowSelection
	tasks         map[string][]WorkflowRecord
	orderedSlices map[string][]workflowstate.WorkflowSlice
	refs          map[workflowstate.Ref][]workflowstate.Ref
}

func newWorkflowDependencies(selection workflowSelection) (workflowDependencies, error) {
	dependencies := workflowDependencies{selection: selection, tasks: map[string][]WorkflowRecord{}, orderedSlices: map[string][]workflowstate.WorkflowSlice{}, refs: map[workflowstate.Ref][]workflowstate.Ref{}}
	for _, record := range selection.records {
		if record.Task.SliceID != "" {
			dependencies.tasks[record.Task.SliceID] = append(dependencies.tasks[record.Task.SliceID], record)
		}
	}

	for id, tasks := range dependencies.tasks {
		slices.SortFunc(tasks, func(left, right WorkflowRecord) int { return cmp.Compare(left.Task.Order, right.Task.Order) })

		dependencies.tasks[id] = tasks
	}

	for id, set := range selection.sets {
		ordered := slices.Clone(set.Set.Slices)
		slices.SortFunc(ordered, func(left, right workflowstate.WorkflowSlice) int { return cmp.Compare(left.Order, right.Order) })

		dependencies.orderedSlices[id] = ordered
	}

	dependencies.buildGraph()

	return dependencies, validateWorkflowGraph(dependencies.refs)
}

func (dependencies *workflowDependencies) buildGraph() {
	for _, ordered := range dependencies.orderedSlices {
		for i, descriptor := range ordered {
			ref := workflowstate.Ref{Kind: "slice", ID: descriptor.ID}

			deps := slices.Clone(descriptor.DependsOn)
			if i > 0 {
				deps = append(deps, workflowstate.Ref{Kind: "slice", ID: ordered[i-1].ID})
			}

			dependencies.refs[ref] = deps
			for _, task := range dependencies.tasks[descriptor.ID] {
				dependencies.refs[ref] = append(dependencies.refs[ref], workflowstate.Ref{Kind: "ticket", ID: task.Task.Key})
			}
		}
	}

	for _, record := range dependencies.selection.records {
		ref := workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}
		dependencies.refs[ref] = dependencies.taskPrerequisites(record)
	}
}

func (dependencies workflowDependencies) taskPrerequisites(record WorkflowRecord) []workflowstate.Ref {
	task := record.Task
	refs := slices.Clone(task.DependsOn)

	tasks := dependencies.tasks[task.SliceID]
	for _, sibling := range tasks {
		if sibling.Task.Order < task.Order {
			refs = append(refs, workflowstate.Ref{Kind: "ticket", ID: sibling.Task.Key})
		}
	}

	for i, descriptor := range dependencies.orderedSlices[task.SetID] {
		if descriptor.ID != task.SliceID {
			continue
		}

		refs = append(refs, descriptor.DependsOn...)
		for _, earlier := range dependencies.orderedSlices[task.SetID][:i] {
			refs = append(refs, workflowstate.Ref{Kind: "slice", ID: earlier.ID})
		}
	}

	unique := []workflowstate.Ref{}
	for _, ref := range refs {
		if !slices.Contains(unique, ref) {
			unique = append(unique, ref)
		}
	}

	return unique
}

func (dependencies workflowDependencies) validateMembers(members []WorkflowPreviewMember) error {
	for _, member := range members {
		task := dependencies.selection.records[member.Key].Task

		refs := dependencies.refs[workflowstate.Ref{Kind: "ticket", ID: member.Key}]
		if len(refs) > workflowstate.PageLimit {
			return fmt.Errorf("task prerequisite observation bound exceeded: %s", member.Key)
		}

		for _, ref := range refs {
			other := dependencies.selection.records[ref.ID].Task
			if ref.Kind == "ticket" && other != nil && task.SliceID != "" && other.SliceID == task.SliceID && other.Order >= task.Order {
				return fmt.Errorf("forward task prerequisite: %s requires %s", task.Key, other.Key)
			}
		}
	}

	return nil
}
