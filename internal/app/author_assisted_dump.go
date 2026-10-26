package app

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

// Reuse a prepared curation or its confirmed dump receipt before attempting a
// new curation. A changed input conflicts instead of replacing the source.
func (factory commandFactory) tryAuthorAssistedDump(ctx context.Context, cmd *cli.Command, state authorAssistedState, modelSummary string, now time.Time) (bool, error) {
	if core.AuthoringArtifactKind(cmd.String("target-kind")) != core.AuthoringArtifactKindDump {
		return false, nil
	}

	summary := core.BuildAssistedCuratedAnswerSummary(state.Session.ID, state.Events, now)
	summary.ModelSummary = modelSummary
	sourcePath := filepath.Join(state.Paths.SessionDir, "curated-summary.md")
	source := core.AuthoringArtifactRef{ID: "curated-summary-" + state.Session.ID, Kind: core.AuthoringArtifactKindDraft, Role: core.AuthoringArtifactRoleOutput, Path: filepath.ToSlash(sourcePath), Description: "Curated assisted authoring answer summary."}

	synthesis, err := core.SynthesizeAssistedArtifact(ctx, core.AssistedArtifactSynthesisOptions{
		ProjectDir: state.ProjectDir, OperationID: cmd.String("operation-id"), Session: state.Session,
		CuratedSummary: summary, SourceRef: source, TargetKind: core.AuthoringArtifactKindDump,
		TargetName: strings.TrimSpace(cmd.String("target-name")), AllowReplace: cmd.Bool("allow-replace"),
		Repository: state.Repository, Now: func() time.Time { return now },
	})
	if errors.Is(err, core.ErrAssistedCurationMissing) {
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return true, printAuthorAssistedFinalized(factory.io.Stdout, core.AssistedAnswerCurationResult{ArtifactPath: sourcePath}, synthesis)
}
