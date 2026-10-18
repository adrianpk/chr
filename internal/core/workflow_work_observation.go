package core

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"
)

// WorkflowWorkObservation is a coherent, non-recovering read of native work.
// Queue members retain shared readiness and native slice/task order.
type WorkflowWorkObservation struct {
	Tickets     []IndexedTicket
	Records     []WorkflowRecord
	Invocations []WorkflowInvocation
	Queue       []WorkflowPreviewMember
	Excluded    map[string]int
}

// InspectWorkflowWork observes all bounded native records without recovering,
// claiming, admitting work, rebuilding SQL or maintaining tracker projections.
func (s *WorkflowStore) InspectWorkflowWork(ctx context.Context, target WorkflowPreviewTarget) (observation WorkflowWorkObservation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withWorkflowInspection(ctx, func() error {
		selection, readErr := s.readWorkflowSelectionSources(ctx, false)
		if readErr != nil {
			return readErr
		}

		observation.Records = selection.snapshot.Records
		observation.Tickets = selection.tickets

		observation.Excluded = map[string]int{}
		if readErr = s.observeWorkflowQueues(ctx, selection, target, &observation); readErr != nil {
			return readErr
		}

		runs, readErr := s.repositoryWorkflowInvocationsPrepared(ctx, func(owner *WorkflowStore) error { return owner.requireNoWorkflowTransition() })
		if readErr != nil {
			return readErr
		}

		for _, run := range runs {
			if run.Namespace == s.layout.Namespace() {
				observation.Invocations = append(observation.Invocations, run.WorkflowInvocation)
			}
		}

		return nil
	})

	return observation, err
}

func (s *WorkflowStore) observeWorkflowQueues(ctx context.Context, selection workflowSelection, target WorkflowPreviewTarget, observation *WorkflowWorkObservation) error {
	sets := slices.Collect(maps.Values(selection.sets))
	slices.SortFunc(sets, func(a, b WorkflowRecord) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}

		if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}

		return strings.Compare(a.ID, b.ID)
	})

	selectors := []WorkflowSelector{}
	for _, set := range sets {
		selectors = append(selectors, WorkflowSelector{Extent: "slice_set", Identity: set.ID})
	}

	selectors = append(selectors, WorkflowSelector{Extent: "loose"})
	for _, selector := range selectors {
		request, err := normalizeWorkflowPreviewRequest(WorkflowPreviewRequest{Selector: selector, Limit: WorkflowPageLimit, Target: target})
		if err != nil {
			return err
		}

		if err = s.observeWorkflowQueue(ctx, selection, request, observation); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) observeWorkflowQueue(ctx context.Context, selection workflowSelection, request WorkflowPreviewRequest, observation *WorkflowWorkObservation) error {
	for {
		preview, err := selection.preview(request)
		if err != nil {
			return err
		}

		if err = s.observePrerequisites(ctx, selection, request, &preview); err != nil {
			return err
		}

		appendWorkflowObservedQueue(observation, preview, request.Offset == 0)

		if preview.NextOffset < 0 {
			return nil
		}

		request.Offset = preview.NextOffset
	}
}

func appendWorkflowObservedQueue(observation *WorkflowWorkObservation, preview WorkflowPreview, first bool) {
	if first {
		for reason, count := range preview.Excluded {
			if reason != "slice_owned" {
				observation.Excluded[reason] += count
			}
		}
	}

	for _, member := range preview.Members {
		if member.Phase != "pending" {
			observation.Excluded["phase_"+member.Phase]++

			continue
		}

		observation.Queue = append(observation.Queue, member)
	}
}
