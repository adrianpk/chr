package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func (factory commandFactory) metadataCommand() *cli.Command {
	return &cli.Command{
		Name:  "metadata",
		Usage: "Manage project and use-case metadata",
		Commands: []*cli.Command{
			factory.metadataProjectCommand(),
			factory.metadataCaseCommand(),
		},
	}
}

func (factory commandFactory) metadataProjectCommand() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage project metadata",
		Commands: []*cli.Command{
			factory.metadataProjectSetCommand(),
		},
	}
}

func (factory commandFactory) metadataCaseCommand() *cli.Command {
	return &cli.Command{
		Name:  "case",
		Usage: "Manage use-case metadata",
		Commands: []*cli.Command{
			factory.metadataCaseSetCommand(),
		},
	}
}

func (factory commandFactory) metadataProjectSetCommand() *cli.Command {
	return &cli.Command{
		Name:   "set",
		Usage:  "Save project metadata",
		Flags:  metadataSetFlags(),
		Action: factory.runMetadataProjectSet,
	}
}

func (factory commandFactory) metadataCaseSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Save use-case metadata",
		ArgsUsage: "<slug>",
		Flags:     metadataSetFlags(),
		Action:    factory.runMetadataCaseSet,
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

func (factory commandFactory) runMetadataProjectSet(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("metadata project set does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.PatchProjectMetadata(projectMetadataPatchFromCommand(cmd))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"project_metadata: %s\n", result.Metadata.SourcePath)
	if err != nil {
		return fmt.Errorf("write metadata project result: %w", err)
	}

	return nil
}

func (factory commandFactory) runMetadataCaseSet(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("metadata case set requires exactly one slug argument")
	}

	slug := strings.TrimSpace(cmd.Args().First())
	if err := validateCaseSlug(slug); err != nil {
		return err
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.PatchCaseMetadata(caseMetadataPatchFromCommand(cmd, slug))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"case_metadata: %s\n", result.Case.SourcePath)
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

func (factory commandFactory) projectService(ctx context.Context, cmd *cli.Command) (service, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return service{}, err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return service{}, err
	}

	return newService(ctx, env, projectContext, factory.deps.assetsFS), nil
}
