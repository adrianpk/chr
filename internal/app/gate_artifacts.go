package app

import (
	"os"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/core/gate"
)

// newGateReportArtifacts constructs the corresponding package service or value.
func newGateReportArtifacts(projectDir string, createdAt time.Time) (gate.ReportArtifacts, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return gate.ReportArtifacts{}, err
	}

	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        "report",
		Title:       "Dev main gate " + createdAt.UTC().Format("2006-01-02 15:04:05Z"),
		Producer:    "chr",
		ExternalIDs: []string{"dev-main-gate"},
		Now:         func() time.Time { return createdAt },
	})
	if err != nil {
		return gate.ReportArtifacts{}, err
	}

	return gate.ReportArtifacts{
		JSONPath:     filepath.Join(artifact.Dir, "evidence", "gate-report.json"),
		MarkdownPath: filepath.Join(artifact.Dir, core.OpsArtifactBodyFile()),
	}, nil
}

// RenderGateReportMarkdown renders the corresponding package data.
func refreshGateReportArtifactMetadata(path string) error {
	artifactDir := filepath.Dir(path)
	if filepath.Base(artifactDir) == "evidence" {
		artifactDir = filepath.Dir(artifactDir)
	}

	if _, err := os.Stat(filepath.Join(artifactDir, "metadata.json")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return core.RefreshOpsArtifactMetadata(artifactDir, "created", time.Now().UTC())
}
