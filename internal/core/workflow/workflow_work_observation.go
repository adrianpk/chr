package workflow

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkObservation is a coherent, non-recovering read of native work.
// Queue members retain shared readiness and native slice/task order.
type WorkObservation struct {
	Tickets     []ticketcore.Indexed
	Records     []Record
	Invocations []Invocation
	Queue       []PreviewMember
	Excluded    map[string]int
}

// InspectWorkflowWork observes all bounded native records without recovering,
// claiming, admitting work, rebuilding SQL or maintaining tracker projections.
func (s *Store) InspectWorkflowWork(ctx context.Context, target workflowstate.PreviewTarget) (observation WorkObservation, err error) {
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

		runs, readErr := s.repositoryWorkflowInvocationsPrepared(ctx, func(owner *Store) error { return owner.requireNoWorkflowTransition() })
		if readErr != nil {
			return readErr
		}

		for _, run := range runs {
			if run.Namespace == s.layout.Namespace() {
				observation.Invocations = append(observation.Invocations, run.Invocation)
			}
		}

		return nil
	})

	return observation, err
}

func (s *Store) observeWorkflowQueues(ctx context.Context, selection workflowSelection, target workflowstate.PreviewTarget, observation *WorkObservation) error {
	sets := slices.Collect(maps.Values(selection.sets))
	slices.SortFunc(sets, func(a, b Record) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}

		if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}

		return strings.Compare(a.ID, b.ID)
	})

	selectors := []workflowstate.Selector{}
	for _, set := range sets {
		selectors = append(selectors, workflowstate.Selector{Extent: "slice_set", Identity: set.ID})
	}

	selectors = append(selectors, workflowstate.Selector{Extent: "loose"})
	for _, selector := range selectors {
		request, err := normalizeWorkflowPreviewRequest(workflowstate.PreviewRequest{Selector: selector, Limit: workflowstate.PageLimit, Target: target})
		if err != nil {
			return err
		}

		if err = s.observeWorkflowQueue(ctx, selection, request, observation); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) observeWorkflowQueue(ctx context.Context, selection workflowSelection, request workflowstate.PreviewRequest, observation *WorkObservation) error {
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

func appendWorkflowObservedQueue(observation *WorkObservation, preview Preview, first bool) {
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
