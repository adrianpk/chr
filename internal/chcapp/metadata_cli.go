package chcapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
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

	operations, projectContext, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	metadata, err := chc.NewProjectMetadataService(projectContext.ProjectDir).Load()
	if err != nil {
		return err
	}

	metadata = metadataFromCommand(cmd, metadata)

	result, err := operations.SaveProjectMetadata(metadata)
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
	if err := validateCaseSlugForCLI(slug); err != nil {
		return err
	}

	operations, projectContext, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	metadata, err := loadCaseMetadataForCLI(projectContext.ProjectDir, slug)
	if err != nil {
		return err
	}

	metadata = metadataFromCommand(cmd, metadata)

	if strings.TrimSpace(metadata.Name) == "" {
		metadata.Name = caseNameFromSlug(slug)
	}

	result, err := operations.SaveCaseMetadata(metadata)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "case_metadata: %s\n", result.Case.SourcePath)
	if err != nil {
		return fmt.Errorf("write metadata case result: %w", err)
	}

	return nil
}

type cliMetadata interface {
	chc.ProjectMetadata | chc.CaseMetadata
}

func metadataFromCommand[T cliMetadata](cmd *cli.Command, metadata T) T {
	switch value := any(&metadata).(type) {
	case *chc.ProjectMetadata:
		if cmd.IsSet("name") {
			value.Name = cmd.String("name")
		}

		if cmd.IsSet("description") {
			value.Description = cmd.String("description")
		}

		if cmd.IsSet("context") {
			value.Context = cmd.String("context")
		}
	case *chc.CaseMetadata:
		if cmd.IsSet("name") {
			value.Name = cmd.String("name")
		}

		if cmd.IsSet("description") {
			value.Description = cmd.String("description")
		}

		if cmd.IsSet("context") {
			value.Context = cmd.String("context")
		}
	}

	return metadata
}

func loadCaseMetadataForCLI(projectDir, slug string) (chc.CaseMetadata, error) {
	path := chc.CaseMetadataPath(projectDir, slug)

	metadata, err := chc.NewCaseMetadataService(projectDir).LoadPath(path)
	if errors.Is(err, os.ErrNotExist) {
		return chc.CaseMetadata{
			SchemaVersion: chc.CaseMetadataSchemaVersion,
			Slug:          slug,
			SourcePath:    path,
		}, nil
	}

	return metadata, err
}

func validateCaseSlugForCLI(slug string) error {
	if slug == "" {
		return fmt.Errorf("metadata case slug is required")
	}

	for _, current := range slug {
		if current >= 'a' && current <= 'z' {
			continue
		}

		if current >= '0' && current <= '9' {
			continue
		}

		if current == '-' {
			continue
		}

		return fmt.Errorf("metadata case slug must contain only lowercase letters, numbers, and dashes")
	}

	return nil
}

func caseNameFromSlug(slug string) string {
	parts := strings.Split(filepath.Base(slug), "-")
	for index, part := range parts {
		if part == "" {
			continue
		}

		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
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
