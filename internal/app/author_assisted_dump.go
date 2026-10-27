package app

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// Reuse a prepared curation or its confirmed dump receipt before attempting a
// new curation. A changed input conflicts instead of replacing the source.
func (factory commandFactory) tryAuthorAssistedDump(ctx context.Context, cmd *cli.Command, state authorAssistedState, modelSummary string, now time.Time) (bool, error) {
	if authoringcore.ArtifactKind(cmd.String("target-kind")) != authoringcore.ArtifactKindDump {
		return false, nil
	}

	summary := authoringcore.BuildAssistedCuratedAnswerSummary(state.Session.ID, state.Events, now)
	summary.ModelSummary = modelSummary
	sourcePath := filepath.Join(state.Paths.SessionDir, "curated-summary.md")
	source := authoringcore.ArtifactRef{ID: "curated-summary-" + state.Session.ID, Kind: authoringcore.ArtifactKindDraft, Role: authoringcore.ArtifactRoleOutput, Path: filepath.ToSlash(sourcePath), Description: "Curated assisted authoring answer summary."}

	synthesis, err := authoringcore.SynthesizeAssistedArtifact(ctx, authoringcore.AssistedArtifactSynthesisOptions{
		ProjectDir: state.ProjectDir, OperationID: cmd.String("operation-id"), Session: state.Session,
		CuratedSummary: summary, SourceRef: source, TargetKind: authoringcore.ArtifactKindDump,
		TargetName: strings.TrimSpace(cmd.String("target-name")), AllowReplace: cmd.Bool("allow-replace"),
		Repository: state.Repository, Now: func() time.Time { return now },
	})
	if errors.Is(err, authoringcore.ErrAssistedCurationMissing) {
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return true, printAuthorAssistedFinalized(factory.io.Stdout, authoringcore.AssistedAnswerCurationResult{ArtifactPath: sourcePath}, synthesis)
}
