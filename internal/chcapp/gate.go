package chcapp

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

func newGateCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "gate",
		Usage: "Run project promotion gates",
		Commands: []*cli.Command{
			newGateDevMainCommand(stdout, deps),
		},
	}
}

func newGateDevMainCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "dev-main",
		Usage: "Collect dev to main promotion evidence",
		Flags: gateDevMainFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runGateDevMain(ctx, cmd, stdout)
		},
	}
}

func gateDevMainFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
		&cli.StringFlag{
			Name:     "base",
			Usage:    "base ref for promotion evidence",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "head",
			Usage:    "head ref for promotion evidence",
			Required: true,
		},
	)
}

func runGateDevMain(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("gate dev-main does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	return runGateDevMainForProject(ctx, stdout, projectContext.ProjectDir, cmd.String("base"), cmd.String("head"))
}

func runGateDevMainForProject(ctx context.Context, stdout io.Writer, projectDir, baseRef, headRef string) error {
	createdAt := time.Now().UTC()
	artifacts := chc.NewGateReportArtifacts(projectDir, createdAt)
	config, err := chc.LoadGateConfig(projectDir)
	if err != nil {
		return err
	}

	evidence, err := chc.NewGateGitEvidenceService().Collect(ctx, chc.GateGitEvidenceRequest{
		ProjectDir: projectDir,
		BaseRef:    baseRef,
		HeadRef:    headRef,
	})
	if err != nil {
		return writeGateFailedEvidenceReport(ctx, createdAt, baseRef, headRef, artifacts, err)
	}

	validation, err := chc.NewGateValidationRunner().Run(ctx, chc.GateValidationRequest{
		ProjectDir: projectDir,
		Command:    config.DevMain.ValidationCommand,
	})
	if err != nil {
		return writeGateFailedValidationRunReport(ctx, createdAt, evidence, validation, artifacts, err)
	}

	classification, err := chc.NewGateEvidenceClassifier().Classify(ctx, chc.GateEvidenceClassificationRequest{
		ProjectDir:    projectDir,
		Files:         evidence.Files,
		HighRiskPaths: config.DevMain.HighRiskPaths,
	})
	if err != nil {
		return err
	}

	reportClassification := classification
	if !config.DevMain.ManualChanges.List {
		reportClassification.ManualOrExternalChanges = nil
	}

	report := chc.NewGateReport(createdAt, evidence, validation, reportClassification, artifacts)
	if !validation.Passed {
		report.Failures = []string{"validation failed"}
	}

	if config.DevMain.Traceability.WarnOnManualChanges && len(classification.ManualOrExternalChanges) > 0 {
		report.Warnings = append(report.Warnings, "manual or external changes present")
	}

	if config.DevMain.ManualChanges.WarnOnHighRiskPaths && len(classification.HighRiskManualChanges) > 0 {
		report.Warnings = append(report.Warnings, "high-risk manual changes present")
	}

	if config.DevMain.Traceability.FailOnMissingAgentEvidence && len(classification.MissingAgentEvidence) > 0 {
		report.Status = chc.GateReportStatusFailed
		report.Failures = append(report.Failures, "missing agent evidence")
	}

	err = writeGateReports(ctx, report)
	if err != nil {
		return err
	}

	err = printGateDevMainReport(stdout, report)
	if err != nil {
		return err
	}

	if !validation.Passed {
		return fmt.Errorf("gate validation failed")
	}

	if config.DevMain.Traceability.FailOnMissingAgentEvidence && len(classification.MissingAgentEvidence) > 0 {
		return fmt.Errorf("missing agent evidence")
	}

	return nil
}

func writeGateFailedEvidenceReport(
	ctx context.Context,
	createdAt time.Time,
	baseRef string,
	headRef string,
	artifacts chc.GateReportArtifacts,
	cause error,
) error {
	report := chc.NewFailedGateReport(createdAt, baseRef, headRef, artifacts, []string{cause.Error()})

	err := writeGateReports(ctx, report)
	if err != nil {
		return err
	}

	return cause
}

func writeGateFailedValidationRunReport(
	ctx context.Context,
	createdAt time.Time,
	evidence chc.GateGitEvidence,
	validation chc.GateValidationResult,
	artifacts chc.GateReportArtifacts,
	cause error,
) error {
	report := chc.NewGateReport(createdAt, evidence, validation, chc.GateEvidenceClassification{}, artifacts)
	report.Failures = []string{cause.Error()}

	err := writeGateReports(ctx, report)
	if err != nil {
		return err
	}

	return cause
}

func writeGateReports(ctx context.Context, report chc.GateReport) error {
	writer := chc.NewGateReportWriter()

	_, err := writer.WriteJSON(ctx, report)
	if err != nil {
		return err
	}

	_, err = writer.WriteMarkdown(ctx, report)
	if err != nil {
		return err
	}

	return nil
}

func printGateDevMainReport(stdout io.Writer, report chc.GateReport) error {
	lines := []string{
		"gate: dev-main",
		fmt.Sprintf("status: %s", report.Status),
		fmt.Sprintf("base: %s %s", report.Base.Ref, report.Base.SHA),
		fmt.Sprintf("head: %s %s", report.Head.Ref, report.Head.SHA),
		fmt.Sprintf("merge_base: %s", report.MergeBase),
		fmt.Sprintf("commits: %d", len(report.Commits)),
		fmt.Sprintf("changed_files: %d", len(report.ChangedFiles)),
		fmt.Sprintf("validation: %t", report.Validation.Passed),
		fmt.Sprintf("json_report: %s", report.Artifacts.JSONPath),
		fmt.Sprintf("markdown_report: %s", report.Artifacts.MarkdownPath),
	}

	if report.DiffStat != "" {
		lines = append(lines, fmt.Sprintf("diff_stat: %s", report.DiffStat))
	}

	for _, commit := range report.Commits {
		value := commit.SHA
		if commit.Subject != "" {
			value += " " + commit.Subject
		}

		lines = append(lines, fmt.Sprintf("commit: %s", value))
	}

	for _, file := range report.ChangedFiles {
		if file.Old != "" {
			lines = append(lines, fmt.Sprintf("file: %s %s -> %s", file.Status, file.Old, file.Path))

			continue
		}

		lines = append(lines, fmt.Sprintf("file: %s %s", file.Status, file.Path))
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write gate evidence: %w", err)
		}
	}

	return nil
}
