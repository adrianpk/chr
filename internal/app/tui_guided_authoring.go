package app

import (
	"context"
	"fmt"
	"slices"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiGuidedExecutor) projectPreparedOriginals(ctx context.Context, observed GuidedUseCaseInspection, view *tui.GuidedView) error {
	if observed.Session == nil {
		return nil
	}

	if !slices.ContainsFunc(observed.Session.Guided.Checkpoints, func(checkpoint authoringcore.GuidedStageCheckpoint) bool { return checkpoint.Stage == "preparation" }) {
		return nil
	}

	for _, kind := range []string{"plan", "tracker"} {
		original, err := e.operations.ReadGuidedUseCaseOriginal(ctx, observed.Session.ID, observed.Namespace, kind)
		if err != nil {
			return err
		}

		view.Originals = append(view.Originals, tui.GuidedDocument{Kind: kind, Reference: original.Reference, Commit: original.Commit, Content: original.Content})
	}

	return nil
}

func tuiGuidedAuthoringLines(observed GuidedUseCaseInspection) []string {
	state := observed.Session.Guided

	lines := []string{}
	if state.SpecApproval != nil {
		lines = append(lines, "Specification approved from committed Dump and Quiz inputs.")
	} else if state.SpecDraft != nil {
		lines = append(lines, "Specification draft saved; automatic approval is pending.")
	}

	work := state.Work
	if work == nil {
		return lines
	}

	if work.Publication != nil && work.Materialization == nil {
		lines = append(lines, "Plan saved; ticket preparation is pending. The saved plan will be reused.")
	}

	if work.Materialization != nil {
		lines = append(lines, fmt.Sprintf("Plan and %d tickets prepared.", len(work.Materialization.Members)))
	}

	return lines
}
