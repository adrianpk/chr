package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiGuidedExecutor) openCase(ctx context.Context, request tui.GuidedRequest) (tui.GuidedView, error) {
	observed, err := e.operations.InspectGuidedUseCase(ctx, "", request.Command.Namespace)
	if err != nil {
		return tui.GuidedView{}, err
	}

	matches := tuiWizardCaseRuns(request, observed.Runs)

	if len(matches) == 0 {
		return tui.GuidedView{}, fmt.Errorf("no retained workflow for this case; choose Wizard to start one")
	}

	if request.CaseSlug == "" && len(matches) > 1 {
		return tui.GuidedView{}, fmt.Errorf("select a named case before opening its workflow")
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

func tuiWizardOriginals(observed GuidedUseCaseInspection) []tui.GuidedDocument {
	originals := []tui.GuidedDocument{}

	for _, checkpoint := range observed.Session.Guided.Checkpoints {
		if checkpoint.Stage != "dump" {
			continue
		}

		for _, artifact := range checkpoint.Artifacts {
			if strings.HasSuffix(artifact.Path, "/artifact.md") {
				originals = append(originals, tui.GuidedDocument{Kind: "dump", Reference: artifact.Path, Commit: checkpoint.Commit, Content: observed.CommittedDump})

				break
			}
		}
	}

	for _, artifact := range observed.Artifacts {
		if artifact.Stage == "dump" {
			continue
		}

		originals = append(originals, tui.GuidedDocument{Kind: artifact.Stage, Reference: artifact.Reference, Commit: artifact.Commit, Content: artifact.Content})
	}

	return originals
}

func tuiWizardCaseRuns(request tui.GuidedRequest, runs []authoringcore.Session) []authoringcore.Session {
	matches := []authoringcore.Session{}

	for _, run := range runs {
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

	return matches
}
