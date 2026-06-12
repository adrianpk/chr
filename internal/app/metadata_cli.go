package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"
)

func newMetadataCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "metadata",
		Usage: "Manage project and use-case metadata",
		Commands: []*cli.Command{
			newMetadataProjectCommand(stdout),
			newMetadataCaseCommand(stdout),
		},
	}
}

func newMetadataProjectCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage project metadata",
		Commands: []*cli.Command{
			newMetadataProjectSetCommand(stdout),
		},
	}
}

func newMetadataCaseCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "case",
		Usage: "Manage use-case metadata",
		Commands: []*cli.Command{
			newMetadataCaseSetCommand(stdout),
		},
	}
}

func newMetadataProjectSetCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "set",
		Usage:  "Save project metadata",
		Flags:  metadataSetFlags(),
		Action: metadataProjectSetAction(stdout),
	}
}

func newMetadataCaseSetCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Save use-case metadata",
		ArgsUsage: "<slug>",
		Flags:     metadataSetFlags(),
		Action:    metadataCaseSetAction(stdout),
	}
}

func metadataSetFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "name",
			Usage: "metadata name",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "metadata description",
		},
		&cli.StringFlag{
			Name:  "context",
			Usage: "long-form markdown context",
		},
	)
}

func metadataProjectSetAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMetadataProjectSet(ctx, cmd, stdout)
	}
}

func metadataCaseSetAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMetadataCaseSet(ctx, cmd, stdout)
	}
}

func runMetadataProjectSet(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("metadata project set does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.PatchProjectMetadata(projectMetadataPatchFromCommand(cmd))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "project_metadata: %s\n", result.Metadata.SourcePath)
	if err != nil {
		return fmt.Errorf("write metadata project result: %w", err)
	}

	return nil
}

func runMetadataCaseSet(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("metadata case set requires exactly one slug argument")
	}

	slug := strings.TrimSpace(cmd.Args().First())
	if err := validateCaseSlug(slug); err != nil {
		return err
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.PatchCaseMetadata(caseMetadataPatchFromCommand(cmd, slug))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "case_metadata: %s\n", result.Case.SourcePath)
	if err != nil {
		return fmt.Errorf("write metadata case result: %w", err)
	}

	return nil
}

func projectMetadataPatchFromCommand(cmd *cli.Command) ProjectMetadataPatchRequest {
	return ProjectMetadataPatchRequest{
		Name:        optionalStringFlag(cmd, "name"),
		Description: optionalStringFlag(cmd, "description"),
		Context:     optionalStringFlag(cmd, "context"),
	}
}

func caseMetadataPatchFromCommand(cmd *cli.Command, slug string) CaseMetadataPatchRequest {
	return CaseMetadataPatchRequest{
		Slug:        slug,
		Name:        optionalStringFlag(cmd, "name"),
		Description: optionalStringFlag(cmd, "description"),
		Context:     optionalStringFlag(cmd, "context"),
	}
}

func optionalStringFlag(cmd *cli.Command, name string) *string {
	if !cmd.IsSet(name) {
		return nil
	}

	value := cmd.String(name)

	return &value
}

func serviceFromCommand(ctx context.Context, cmd *cli.Command) (service, commandProjectContext, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return service{}, commandProjectContext{}, err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return service{}, commandProjectContext{}, err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return service{}, commandProjectContext{}, err
	}

	return newService(ctx, env, projectContext, deps.assetsFS), projectContext, nil
}
