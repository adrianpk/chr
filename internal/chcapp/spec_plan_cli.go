package chcapp

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

func newSpecCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "spec",
		Usage: "Manage generated specs",
		Commands: []*cli.Command{
			newSpecGenerateCommand(stdout),
			newSpecListCommand(stdout),
		},
	}
}

func newPlanCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "plan",
		Usage: "Manage generated plans",
		Commands: []*cli.Command{
			newPlanGenerateCommand(stdout),
			newPlanListCommand(stdout),
			newPlanRegenerateCommand(stdout),
			newPlanTrackingCommand(stdout),
		},
	}
}

func newSpecListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List generated specs",
		Flags:  ticketProjectFlags(),
		Action: specListAction(stdout),
	}
}

func newSpecGenerateCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "generate",
		Usage:  "Generate a spec",
		Flags:  quizScopeFlags(),
		Action: specGenerateAction(stdout),
	}
}

func newPlanListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List generated plans",
		Flags:  ticketProjectFlags(),
		Action: planListAction(stdout),
	}
}

func newPlanGenerateCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "generate",
		Usage:  "Generate a plan",
		Flags:  quizScopeFlags(),
		Action: planGenerateAction(stdout),
	}
}

func newPlanRegenerateCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "regenerate",
		Usage:  "Regenerate a plan",
		Flags:  quizScopeFlags(),
		Action: planRegenerateAction(stdout),
	}
}

func newPlanTrackingCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "tracking",
		Usage: "Manage generated plan tracking",
		Commands: []*cli.Command{
			newPlanTrackingCreateCommand(stdout),
		},
	}
}

func newPlanTrackingCreateCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "create",
		Usage:  "Create tracking and tickets for an existing generated plan",
		Flags:  quizScopeFlags(),
		Action: planTrackingCreateAction(stdout),
	}
}

func specListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runSpecList(ctx, cmd, stdout)
	}
}

func specGenerateAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runSpecGenerate(ctx, cmd, stdout)
	}
}

func planListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runPlanList(ctx, cmd, stdout)
	}
}

func planGenerateAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runPlanGenerate(ctx, cmd, stdout)
	}
}

func planRegenerateAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runPlanRegenerate(ctx, cmd, stdout)
	}
}

func planTrackingCreateAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runPlanTrackingCreate(ctx, cmd, stdout)
	}
}

func runSpecList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("spec list does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListSpecs()
	if err != nil {
		return err
	}

	for _, spec := range result.Specs {
		_, err = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", spec.RelativePath, spec.Scope.CaseSlug, spec.Status, spec.Title)
		if err != nil {
			return fmt.Errorf("write spec list result: %w", err)
		}
	}

	return nil
}

func runSpecGenerate(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("spec generate does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.GenerateSpec(ctx, SpecGenerateRequest{Scope: quizScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "%s\nspec: %s\n", result.Message, result.Spec.RelativePath)
	if err != nil {
		return fmt.Errorf("write spec generate result: %w", err)
	}

	return nil
}

func runPlanList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan list does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListPlans()
	if err != nil {
		return err
	}

	for _, plan := range result.Plans {
		_, err = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", plan.RelativePath, plan.Scope.CaseSlug, plan.Status, plan.Title)
		if err != nil {
			return fmt.Errorf("write plan list result: %w", err)
		}
	}

	return nil
}

func runPlanGenerate(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan generate does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.GeneratePlan(ctx, PlanGenerateRequest{Scope: quizScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "%s\nplan: %s\n", result.Message, result.Plan.RelativePath)
	if err != nil {
		return fmt.Errorf("write plan generate result: %w", err)
	}

	return nil
}

func runPlanRegenerate(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan regenerate does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.RegeneratePlan(ctx, PlanRegenerateRequest{Scope: quizScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "%s\nplan: %s\n", result.Message, result.Plan.RelativePath)
	if err != nil {
		return fmt.Errorf("write plan regenerate result: %w", err)
	}

	return nil
}

func runPlanTrackingCreate(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("plan tracking create does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.MaterializePlanTracking(ctx, PlanTrackingRequest{Scope: quizScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "%s\ntracking: %s\ntickets: %d\n", result.Message, result.TrackingPath, len(result.Tickets))
	if err != nil {
		return fmt.Errorf("write plan tracking result: %w", err)
	}

	return nil
}
