package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) gateCommand() *cli.Command {
	return &cli.Command{
		Name:  "gate",
		Usage: "Run project promotion gates",
		Commands: []*cli.Command{
			factory.gateDevMainCommand(),
			factory.gateWoodpeckerCommand(),
		},
	}
}

func (factory commandFactory) gateDevMainCommand() *cli.Command {
	return &cli.Command{
		Name:  "dev-main",
		Usage: "Collect dev to main promotion evidence",
		Flags: gateDevMainFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return factory.runGateDevMain(ctx, cmd)
		},
	}
}

func (factory commandFactory) gateWoodpeckerCommand() *cli.Command {
	return &cli.Command{
		Name:  "woodpecker",
		Usage: "Configure Woodpecker gate integration",
		Commands: []*cli.Command{
			factory.gateWoodpeckerInitCommand(),
		},
	}
}

func (factory commandFactory) gateWoodpeckerInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create or print a Woodpecker dev-main gate config",
		Flags: gateWoodpeckerInitFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return factory.runGateWoodpeckerInit(ctx, cmd)
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

func gateWoodpeckerInitFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the Woodpecker gate config without writing it",
		},
		&cli.BoolFlag{
			Name:  "write",
			Usage: "update an existing generated Woodpecker gate config",
		},
	)
}

func (factory commandFactory) runGateDevMain(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("gate dev-main does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	return runGateDevMainForProject(ctx, factory.io.Stdout, projectContext.ProjectDir, cmd.String("base"), cmd.String("head"))
}

func (factory commandFactory) runGateWoodpeckerInit(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("gate woodpecker init does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	result, err := core.InitGateWoodpecker(core.GateWoodpeckerInitRequest{
		ProjectDir: projectContext.ProjectDir,
		DryRun:     cmd.Bool("dry-run"),
		Write:      cmd.Bool("write"),
	})
	if err != nil {
		return err
	}

	return printGateWoodpeckerInitResult(factory.io.Stdout, result)
}

func printGateWoodpeckerInitResult(stdout io.Writer, result core.GateWoodpeckerInitResult) error {
	_, err := fmt.Fprintf(stdout, "woodpecker: init\nconfig: %s\naction: %s\nfeedback: %s\nwrote: %t\n\n%s",
		result.ConfigPath,
		result.Action,
		result.FeedbackMechanism,
		result.Wrote,
		result.Snippet,
	)

	return err
}

func runGateDevMainForProject(ctx context.Context, stdout io.Writer, projectDir, baseRef, headRef string) error {
	createdAt := time.Now().UTC()

	artifacts, err := core.NewGateReportArtifacts(projectDir, createdAt)
	if err != nil {
		return err
	}

	config, err := core.LoadGateConfig(projectDir)
	if err != nil {
		return err
	}

	evidence, err := core.NewGateGitEvidenceService().Collect(ctx, core.GateGitEvidenceRequest{
		ProjectDir: projectDir,
		BaseRef:    baseRef,
		HeadRef:    headRef,
	})
	if err != nil {
		return writeGateFailedEvidenceReport(ctx, createdAt, baseRef, headRef, artifacts, err)
	}

	validation, err := core.NewGateValidationRunner().Run(ctx, core.GateValidationRequest{
		ProjectDir: projectDir,
		Command:    config.DevMain.ValidationCommand,
	})
	if err != nil {
		return writeGateFailedValidationRunReport(ctx, createdAt, evidence, validation, artifacts, err)
	}

	classification, err := core.NewGateEvidenceClassifier().Classify(ctx, core.GateEvidenceClassificationRequest{
		ProjectDir:    projectDir,
		Files:         evidence.Files,
		HighRiskPaths: config.DevMain.HighRiskPaths,
	})
	if err != nil {
		return err
	}

	report := buildGateDevMainReport(createdAt, evidence, validation, classification, config.DevMain, artifacts)

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

	if shouldFailGateMissingAgentEvidence(config.DevMain, classification) {
		return fmt.Errorf("missing agent evidence")
	}

	return nil
}

func buildGateDevMainReport(
	createdAt time.Time,
	evidence core.GateGitEvidence,
	validation core.GateValidationResult,
	classification core.GateEvidenceClassification,
	config core.GateDevMainConfig,
	artifacts core.GateReportArtifacts,
) core.GateReport {
	reportClassification := classification
	if !config.ManualChanges.List {
		reportClassification.ManualOrExternalChanges = nil
	}

	report := core.NewGateReport(createdAt, evidence, validation, reportClassification, artifacts)
	report.Failures = gateDevMainFailures(validation, classification, config)

	report.Warnings = gateDevMainWarnings(classification, config)
	if len(report.Failures) > 0 {
		report.Status = core.GateReportStatusFailed
	}

	return report
}

func gateDevMainFailures(
	validation core.GateValidationResult,
	classification core.GateEvidenceClassification,
	config core.GateDevMainConfig,
) []string {
	failures := []string{}
	if !validation.Passed {
		failures = append(failures, "validation failed")
	}

	if shouldFailGateMissingAgentEvidence(config, classification) {
		failures = append(failures, "missing agent evidence")
	}

	return failures
}

func gateDevMainWarnings(
	classification core.GateEvidenceClassification,
	config core.GateDevMainConfig,
) []string {
	warnings := []string{}
	if config.Traceability.WarnOnManualChanges && len(classification.ManualOrExternalChanges) > 0 {
		warnings = append(warnings, "manual or external changes present")
	}

	if config.ManualChanges.WarnOnHighRiskPaths && len(classification.HighRiskManualChanges) > 0 {
		warnings = append(warnings, "high-risk manual changes present")
	}

	return warnings
}

func shouldFailGateMissingAgentEvidence(
	config core.GateDevMainConfig,
	classification core.GateEvidenceClassification,
) bool {
	return config.Traceability.FailOnMissingAgentEvidence && len(classification.MissingAgentEvidence) > 0
}

func writeGateFailedEvidenceReport(
	ctx context.Context,
	createdAt time.Time,
	baseRef string,
	headRef string,
	artifacts core.GateReportArtifacts,
	cause error,
) error {
	report := core.NewFailedGateReport(createdAt, baseRef, headRef, artifacts, []string{cause.Error()})

	err := writeGateReports(ctx, report)
	if err != nil {
		return err
	}

	return cause
}

func writeGateFailedValidationRunReport(
	ctx context.Context,
	createdAt time.Time,
	evidence core.GateGitEvidence,
	validation core.GateValidationResult,
	artifacts core.GateReportArtifacts,
	cause error,
) error {
	report := core.NewGateReport(createdAt, evidence, validation, core.GateEvidenceClassification{}, artifacts)
	report.Failures = []string{cause.Error()}

	err := writeGateReports(ctx, report)
	if err != nil {
		return err
	}

	return cause
}

func writeGateReports(ctx context.Context, report core.GateReport) error {
	writer := core.NewGateReportWriter()

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

func printGateDevMainReport(stdout io.Writer, report core.GateReport) error {
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
