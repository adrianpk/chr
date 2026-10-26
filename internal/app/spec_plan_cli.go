package app

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (factory commandFactory) specCommand() *cli.Command {
	return &cli.Command{
		Name:  "spec",
		Usage: "Manage generated specs",
		Commands: []*cli.Command{
			factory.specGenerateCommand(),
			factory.specListCommand(),
		},
	}
}

func (factory commandFactory) planCommand() *cli.Command {
	return &cli.Command{
		Name:  "plan",
		Usage: "Manage generated plans",
		Commands: []*cli.Command{
			factory.planGenerateCommand(),
			factory.planListCommand(),
			factory.planRegenerateCommand(),
			factory.planTrackingCommand(),
		},
	}
}

func (factory commandFactory) specListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List generated specs",
		Flags:  ticketProjectFlags(),
		Action: factory.runSpecList,
	}
}

func (factory commandFactory) specGenerateCommand() *cli.Command {
	return &cli.Command{
		Name:   "generate",
		Usage:  "Generate a spec",
		Flags:  workspaceScopeFlags(),
		Action: factory.runSpecGenerate,
	}
}

func (factory commandFactory) planListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List generated plans",
		Flags:  ticketProjectFlags(),
		Action: factory.runPlanList,
	}
}

func (factory commandFactory) planGenerateCommand() *cli.Command {
	return &cli.Command{
		Name:   "generate",
		Usage:  "Generate a plan",
		Flags:  workspaceScopeFlags(),
		Action: factory.runPlanGenerate,
	}
}

func (factory commandFactory) planRegenerateCommand() *cli.Command {
	return &cli.Command{
		Name:   "regenerate",
		Usage:  "Regenerate a plan",
		Flags:  workspaceScopeFlags(),
		Action: factory.runPlanRegenerate,
	}
}

func (factory commandFactory) planTrackingCommand() *cli.Command {
	return &cli.Command{
		Name:  "tracking",
		Usage: "Manage generated plan tracking",
		Commands: []*cli.Command{
			factory.planTrackingCreateCommand(),
		},
	}
}

func (factory commandFactory) planTrackingCreateCommand() *cli.Command {
	return &cli.Command{
		Name:   "create",
		Usage:  "Create tracking and tickets for an existing generated plan",
		Flags:  workspaceScopeFlags(),
		Action: factory.runPlanTrackingCreate,
	}
}

func (factory commandFactory) runSpecList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("spec list does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListSpecs()
	if err != nil {
		return err
	}

	for _, spec := range result.Specs {
		_, err = fmt.Fprintf(factory.io.Stdout,
			"%s\t%s\t%s\t%s\n", spec.RelativePath, spec.Scope.CaseSlug, spec.Status, spec.Title)
		if err != nil {
			return fmt.Errorf("write spec list result: %w", err)
		}
	}

	return nil
}

func (factory commandFactory) runSpecGenerate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("spec generate does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.GenerateSpec(ctx, SpecGenerateRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"%s\nspec: %s\n", result.Message, result.Spec.RelativePath)
	if err != nil {
		return fmt.Errorf("write spec generate result: %w", err)
	}

	return nil
}

func (factory commandFactory) runPlanList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan list does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListPlans()
	if err != nil {
		return err
	}

	for _, plan := range result.Plans {
		_, err = fmt.Fprintf(factory.io.Stdout,
			"%s\t%s\t%s\t%s\n", plan.RelativePath, plan.Scope.CaseSlug, plan.Status, plan.Title)
		if err != nil {
			return fmt.Errorf("write plan list result: %w", err)
		}
	}

	return nil
}

func (factory commandFactory) runPlanGenerate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan generate does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.GeneratePlan(ctx, PlanGenerateRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"%s\nplan: %s\n", result.Message, result.Plan.RelativePath)
	if err != nil {
		return fmt.Errorf("write plan generate result: %w", err)
	}

	return nil
}

func (factory commandFactory) runPlanRegenerate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan regenerate does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.RegeneratePlan(ctx, PlanRegenerateRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"%s\nplan: %s\n", result.Message, result.Plan.RelativePath)
	if err != nil {
		return fmt.Errorf("write plan regenerate result: %w", err)
	}

	return nil
}

func (factory commandFactory) runPlanTrackingCreate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan tracking create does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.MaterializePlanTracking(ctx, PlanTrackingRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"%s\ntracking: %s\ntickets: %d\n", result.Message, result.TrackingPath, len(result.Tickets))
	if err != nil {
		return fmt.Errorf("write plan tracking result: %w", err)
	}

	return nil
}
