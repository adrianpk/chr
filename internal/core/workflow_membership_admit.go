package core

import (
	"context"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) admitWorkflowLaterMembers(ctx context.Context, run *workflowInvocationFile, selection workflowSelection, later []WorkflowPreviewMember) error {
	if len(run.Members)+len(later) > WorkflowRecordLimit {
		return fmt.Errorf("workflow later admission member bound exceeded")
	}

	ordered, err := mergeWorkflowLaterMembers(*run, selection, later)
	if err != nil {
		return err
	}

	preview := WorkflowPreview{Members: later, admitted: workflowAdmittedRefs(ordered, run.Selection.Selector)}
	if err := s.observePrerequisites(ctx, selection, run.ExecutionSelection(), &preview); err != nil {
		return err
	}

	if err := s.checkWorkflowLaterOwnership(ctx, *run, preview.Members); err != nil {
		return err
	}

	known := map[string]WorkflowMember{}
	for _, member := range append(workflowNewMembers(preview.Members), run.Members...) {
		known[member.Admission.Key] = member
	}

	run.Members = nil
	for _, member := range ordered {
		run.Members = append(run.Members, known[member.Key])
	}

	appendWorkflowLaterSetImages(run, selection)
	sortWorkflowLaterSetImages(run, selection)

	delete(run.Excluded, "later")
	workflowContinueAmendedSet(run)

	return nil
}

func sortWorkflowLaterSetImages(run *workflowInvocationFile, selection workflowSelection) {
	slices.SortFunc(run.SetImages, func(a, b WorkflowInputImage) int {
		return selection.compareSets(workflowInvocationSetID(a), workflowInvocationSetID(b), run.SetOrderRule)
	})
}

func workflowContinueAmendedSet(run *workflowInvocationFile) {
	if run.Outcome == "set_closure_required" && run.Cursor < len(run.Members) && run.Members[run.Cursor].Admission.SetID == workflowNextClosingSet(run.WorkflowInvocation) {
		run.Status, run.Outcome = "pending", "later_work_admitted"
	}
}

func workflowNewMembers(members []WorkflowPreviewMember) []WorkflowMember {
	result := make([]WorkflowMember, 0, len(members))
	for _, member := range members {
		result = append(result, WorkflowMember{Admission: member, ExpectedRevision: member.Revision, ExpectedSourceHash: member.SourceHash})
	}

	return result
}

func mergeWorkflowLaterMembers(run workflowInvocationFile, selection workflowSelection, later []WorkflowPreviewMember) ([]WorkflowPreviewMember, error) {
	ordered := append([]WorkflowPreviewMember(nil), later...)

	known := map[string]int{}
	for i, member := range run.Members {
		known[member.Admission.Key] = i
		ordered = append(ordered, member.Admission)
	}

	if _, err := selection.orderMembers(ordered, run.Selection.Selector); err != nil {
		return nil, err
	}

	nextOld := 0

	for i, member := range ordered {
		index, exists := known[member.Key]
		if exists {
			if index != nextOld || i < run.Cursor && index != i {
				return nil, fmt.Errorf("%w: frozen work reordered", workflowstate.ErrConflict)
			}

			nextOld++
		} else if i < run.Cursor {
			return nil, fmt.Errorf("%w: later work precedes confirmed or validated progress", workflowstate.ErrConflict)
		}
	}

	return ordered, nil
}

func appendWorkflowLaterSetImages(run *workflowInvocationFile, selection workflowSelection) {
	known := map[string]bool{}
	for _, image := range run.SetImages {
		known[image.Path] = true
	}

	for _, member := range run.Members {
		set := selection.sets[member.Admission.SetID]
		if known[set.Path] {
			continue
		}

		run.SetImages = append(run.SetImages, WorkflowInputImage{Path: set.Path, Revision: set.Revision, SourceHash: set.SourceHash, UseCase: set.Set.UseCase})
		known[set.Path] = true
	}
}

func (s *WorkflowStore) checkWorkflowLaterOwnership(ctx context.Context, run workflowInvocationFile, later []WorkflowPreviewMember) error {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return err
	}

	others := []workflowInvocationFile{}

	for _, other := range runs {
		if other.ID != run.ID || other.Namespace != run.Namespace || other.OwnerRoot != run.OwnerRoot {
			others = append(others, other)
		}
	}

	run.Members = append(run.Members, workflowNewMembers(later)...)

	if err := s.guardWorkflowAdmission(others, run); err != nil {
		return err
	}

	limits, err := s.workflowRepositoryCapacity(ctx, runs)
	if err != nil {
		return err
	}

	return guardWorkflowReadyQueue(others, run, limits)
}
