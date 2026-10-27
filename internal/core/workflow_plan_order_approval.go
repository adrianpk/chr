package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Plan/map publication and approved binding share authoring exclusion before
// taking workflow exclusion. Progress writes continue to use the workflow owner.
func (s *WorkflowStore) withWorkflowPlanMutation(ctx context.Context, action func(context.Context) error) error {
	return WithAuthoringMutation(ctx, s.layout.ProjectDir(), func(ctx context.Context) error {
		return s.withLock(ctx, func() error { return action(ctx) })
	})
}

func (r *fileAuthoringSessionRepository) checkPlanWorkflowPending(reference string) error {
	root := filepath.Dir(filepath.Dir(filepath.Dir(reference)))

	path, err := r.authoringPath(filepath.ToSlash(filepath.Join(root, workflowJournalPath)))
	if err != nil {
		return err
	}

	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	return nil
}

func (s *WorkflowStore) validateApprovedOrder(plan GeneratedWorkflowPlan, setID string) error {
	if plan.PlanRevision == nil {
		return s.requirePlanRevisionBinding(plan.PlanRef)
	}

	binding := *plan.PlanRevision
	if err := s.validatePlannedApproval(plan, binding); err != nil {
		return err
	}

	current, err := s.readPlanOrderBinding(binding)
	if err != nil {
		return err
	}

	if current != binding {
		return workflowstate.ErrConflict
	}

	return s.validateApprovalOrderImages(binding, setID)
}

func (s *WorkflowStore) validatePlannedApproval(plan GeneratedWorkflowPlan, binding PlanRevisionBinding) error {
	if !validPlanBindingPaths(s.layout.Namespace(), binding) || plan.PlanRef != binding.BodyRef || plan.PlanHash != binding.BodyHash {
		return workflowstate.ErrConflict
	}

	if err := s.validateOrderScope(binding, plan.UseCase); err != nil {
		return err
	}

	planned, err := s.workflowPlannedOrder(binding)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(orderedPlanSlices(planned.Slices), orderedPlanSlices(plannedWorkflowOrder(plan))) {
		return fmt.Errorf("%w: approved hierarchy differs from retained planned order", workflowstate.ErrConflict)
	}

	return nil
}

func (s *WorkflowStore) validateApprovalOrderImages(binding PlanRevisionBinding, setID string) error {
	order, err := s.readWorkflowTrackerFile(binding.OrderRef)
	if err != nil {
		return err
	}

	tracker, err := s.readWorkflowTrackerFile(binding.TrackerRef)
	if err != nil {
		return err
	}

	if contenthash.Sum(order) != binding.OrderHash || contenthash.Sum(tracker) != binding.TrackerHash {
		return ErrWorkflowProjectionStale
	}

	var document PlanOrderDocument
	if err := decodeAuthoringState(order, &document); err != nil {
		return err
	}

	if document.SetID != "" && document.SetID != setID {
		return fmt.Errorf("%w: revision is already bound to another set", workflowstate.ErrConflict)
	}

	return s.checkCurrentOrderOutput(binding)
}

func (s *WorkflowStore) requirePlanRevisionBinding(reference string) error {
	full := filepath.ToSlash(filepath.Join(s.layout.RelativeRoot(), reference))
	if validPlanRevisionJournalPath(full) {
		return fmt.Errorf("%w: generated plan requires its explicit revision binding", workflowstate.ErrConflict)
	}

	metadata, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join(filepath.Dir(reference), "metadata.json")))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(metadata, &fields); err != nil {
		return err
	}

	if fields["plan_binding"] != nil {
		return fmt.Errorf("%w: generated plan requires its explicit revision binding", workflowstate.ErrConflict)
	}

	return nil
}

func (s *WorkflowStore) readPlanOrderBinding(binding PlanRevisionBinding) (PlanRevisionBinding, error) {
	var current PlanRevisionBinding

	data, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), "binding.json")))
	if err != nil {
		return current, err
	}

	err = decodeAuthoringState(data, &current)

	return current, err
}

func plannedWorkflowOrder(plan GeneratedWorkflowPlan) []PlanOrderSlice {
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

func (s *WorkflowStore) validateOrderScope(binding PlanRevisionBinding, useCase string) error {
	body, err := s.readWorkflowTrackerFile(binding.BodyRef)
	if err != nil {
		return err
	}

	front, _, err := splitWorkflowMarkdown(body)
	if err != nil {
		return err
	}

	fields, err := projectDocumentFields(front)
	if err != nil {
		return err
	}

	if fields["project_id"] != s.projectID || fields["case_slug"] != useCase {
		return workflowstate.ErrConflict
	}

	plannedTracker, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join(filepath.Dir(binding.TrackerRef), "planned-tracker.md")))
	if err != nil {
		return err
	}

	if contenthash.Sum(plannedTracker) != binding.PlannedTrackerHash {
		return ErrWorkflowProjectionStale
	}

	return nil
}
