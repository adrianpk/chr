package app

import (
	"context"
	"fmt"
	"sort"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiGuidedExecutor) openCase(ctx context.Context, request tui.GuidedRequest) (tui.GuidedView, error) {
	observed, err := e.operations.InspectGuidedUseCase(ctx, "", request.Command.Namespace)
	if err != nil {
		return tui.GuidedView{}, err
	}

	matches := []authoringcore.Session{}

	for _, run := range observed.Runs {
		if request.CaseSlug == "" || run.Guided.CaseSlug == request.CaseSlug {
			matches = append(matches, run)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].UpdatedAt.Equal(matches[j].UpdatedAt) {
			return matches[i].ID < matches[j].ID
		}

		return matches[i].UpdatedAt.After(matches[j].UpdatedAt)
	})

	if len(matches) == 0 {
		return tui.GuidedView{}, fmt.Errorf("no retained workflow for this case; choose Wizard to start one")
	}

	selected := matches[0]
	for _, run := range matches {
		if run.ID == request.Command.RunID {
			selected = run

			break
		}
	}

	request.Action, request.Command.RunID = "inspect", selected.ID

	return e.read(ctx, request)
}
