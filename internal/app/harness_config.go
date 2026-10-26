package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) harnessCommand() *cli.Command {
	return &cli.Command{Name: "harness", Usage: "Inspect or change named harness profiles for new project operations", Commands: []*cli.Command{
		factory.harnessStatusCommand(),
		factory.harnessChangeCommand(false),
		factory.harnessChangeCommand(true),
	}}
}

func (factory commandFactory) harnessStatusCommand() *cli.Command {
	flags := append(workProjectFlags(), &cli.BoolFlag{Name: "json", Usage: "print current selections and prepared choices as JSON"},
		&cli.StringFlag{Name: "action", Value: "spec_generation", Usage: "generation action whose legacy default is displayed"})

	return &cli.Command{Name: "status", Usage: "Inspect selections and readiness without calling a model", Flags: flags, Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() != 0 {
			return fmt.Errorf("harness status does not accept positional arguments")
		}

		operations, _, err := factory.workServiceFromCommand(ctx, cmd)
		if err != nil {
			return err
		}

		result, err := operations.InspectHarnessSelections(ctx, cmd.String("action"))
		if err != nil {
			return err
		}

		if cmd.Bool("json") {
			return json.NewEncoder(factory.io.Stdout).Encode(result)
		}

		return renderHarnessSelectionStatus(factory.io.Stdout, result)
	}}
}

func (factory commandFactory) harnessChangeCommand(clearChoice bool) *cli.Command {
	name, usage := "select", "Select prepared profiles for one or both roles"

	flags := append(workProjectFlags(), &cli.StringFlag{Name: "config-revision", Required: true, Usage: "complete observed configuration revision"},
		&cli.StringFlag{Name: "request-id", Required: true, Usage: "stable identity for this exact change"}, &cli.BoolFlag{Name: "json", Usage: "print the accepted selection snapshot as JSON"})

	if clearChoice {
		name, usage = "clear", "Clear explicit role selections and restore their legacy defaults"

		flags = append(flags, &cli.BoolFlag{Name: "worker"}, &cli.BoolFlag{Name: "generation"})
	} else {
		flags = append(flags, &cli.StringFlag{Name: "worker", Usage: "prepared worker profile name"}, &cli.StringFlag{Name: "generation", Usage: "prepared generation profile name"})
	}

	return &cli.Command{Name: name, Usage: usage, Flags: flags, Action: func(ctx context.Context, cmd *cli.Command) error {
		request, err := harnessChangeFromCommand(cmd, clearChoice)
		if err != nil {
			return err
		}

		operations, project, err := factory.workServiceFromCommand(ctx, cmd)
		if err != nil {
			return err
		}

		userspace, err := core.ResolveProjectUserspace(core.ResolveProjectUserspaceOptions{ProjectDir: project.ProjectDir})
		if err != nil {
			return err
		}

		request.ProjectRef, request.Binding.Operator = project.ProjectID, userspace.OperatorID

		result, err := operations.ChangeHarnessSelections(ctx, request, localHarnessSelectionAuthority(project.ProjectDir, userspace.OperatorID))
		if err != nil {
			return err
		}

		if cmd.Bool("json") {
			return json.NewEncoder(factory.io.Stdout).Encode(result)
		}

		_, err = fmt.Fprintf(factory.io.Stdout,
			"Worker: %s\nGeneration: %s\nConfiguration revision: %s\nApplies to new operations.\n", harnessChoiceLabel(result.Worker), harnessChoiceLabel(result.Generation), result.Revision)

		return err
	}}
}

func harnessChangeFromCommand(cmd *cli.Command, clearChoice bool) (core.HarnessSelectionChangeRequest, error) {
	result := core.HarnessSelectionChangeRequest{ExpectedRevision: cmd.String("config-revision"), Binding: core.WorkflowMutationBinding{RequestID: cmd.String("request-id")}}
	if cmd.NArg() != 0 {
		return result, fmt.Errorf("harness changes do not accept positional arguments")
	}

	for _, field := range []struct {
		name   string
		target **string
	}{{"worker", &result.Worker}, {"generation", &result.Generation}} {
		if !cmd.IsSet(field.name) {
			continue
		}

		value, err := harnessRoleChoiceFromCommand(cmd, field.name, clearChoice)
		if err != nil {
			return result, err
		}

		*field.target = &value
	}

	if result.Worker == nil && result.Generation == nil {
		return result, fmt.Errorf("select at least one role")
	}

	return result, nil
}

func harnessRoleChoiceFromCommand(cmd *cli.Command, name string, clearChoice bool) (string, error) {
	if clearChoice {
		if !cmd.Bool(name) {
			return "", fmt.Errorf("clear requires a true role flag")
		}

		return "", nil
	}

	value := cmd.String(name)
	if value == "" || strings.TrimSpace(value) != value {
		return "", fmt.Errorf("select requires a prepared profile name")
	}

	return value, nil
}

func harnessChoiceLabel(name string) string {
	if name == "" {
		return "legacy"
	}

	return name
}

func renderHarnessSelectionStatus(stdout io.Writer, status HarnessSelectionStatus) error {
	if _, err := fmt.Fprintf(stdout, "Project: %s\nConfiguration revision: %s\n", status.Snapshot.ProjectRef, status.Snapshot.Revision); err != nil {
		return err
	}

	for _, role := range []core.HarnessSelection{status.Worker, status.Generation} {
		if _, err := fmt.Fprintf(stdout, "%s: %s; backend: %s; provider: %s; model: %s; locally ready: %t\n", role.Role, harnessChoiceLabel(role.Profile), role.Backend, role.ModelProvider, role.Model, role.Readiness.Ready()); err != nil {
			return err
		}
	}

	for _, choice := range status.Profiles {
		if _, err := fmt.Fprintf(stdout, "Profile %s: %s; model: %s; worker ready: %t; generation ready: %t\n", choice.Name, choice.Harness, choice.Model, choice.Worker.Ready(), choice.Generation.Ready()); err != nil {
			return err
		}
	}

	for _, problem := range status.Problems {
		if _, err := fmt.Fprintln(stdout, problem); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(stdout, "Selections apply to new operations; legacy generation defaults depend on the displayed action.")

	return err
}
