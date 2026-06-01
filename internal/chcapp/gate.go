package chcapp

import (
	"context"
	"fmt"
	"io"

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

	evidence, err := chc.NewGateGitEvidenceService().Collect(ctx, chc.GateGitEvidenceRequest{
		ProjectDir: projectContext.ProjectDir,
		BaseRef:    cmd.String("base"),
		HeadRef:    cmd.String("head"),
	})
	if err != nil {
		return err
	}

	return printGateGitEvidence(stdout, evidence)
}

func printGateGitEvidence(stdout io.Writer, evidence chc.GateGitEvidence) error {
	lines := []string{
		"gate: dev-main",
		fmt.Sprintf("base: %s %s", evidence.Base.Ref, evidence.Base.SHA),
		fmt.Sprintf("head: %s %s", evidence.Head.Ref, evidence.Head.SHA),
		fmt.Sprintf("merge_base: %s", evidence.MergeBase),
		fmt.Sprintf("commits: %d", len(evidence.Commits)),
		fmt.Sprintf("changed_files: %d", len(evidence.Files)),
	}

	if evidence.DiffStat != "" {
		lines = append(lines, fmt.Sprintf("diff_stat: %s", evidence.DiffStat))
	}

	for _, commit := range evidence.Commits {
		value := commit.SHA
		if commit.Subject != "" {
			value += " " + commit.Subject
		}

		lines = append(lines, fmt.Sprintf("commit: %s", value))
	}

	for _, file := range evidence.Files {
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
