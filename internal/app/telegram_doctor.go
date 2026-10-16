package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func newTelegramDoctorCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name: "doctor", Usage: "Check receiver readiness without starting agents or sending messages",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{Name: "check-connection"}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runTelegramDoctor(ctx, cmd, stdout, core.NewTelegramReceiverAPI())
		},
	}
}

func runTelegramDoctor(ctx context.Context, cmd *cli.Command, stdout io.Writer, api core.TelegramReceiverAPI) (err error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	config, err := core.LoadTelegramReceiverConfig(env)
	if err != nil {
		return err
	}

	token, err := core.TelegramReceiverToken(env, config)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return err
	}

	operations := newService(ctx, env, commandProjectContext{}, deps.assetsFS)
	if err := printTelegramProjectReadiness(stdout, operations, config); err != nil {
		return err
	}

	if !cmd.Bool("check-connection") {
		_, err := fmt.Fprintln(stdout, "telegram_doctor: local_configuration_ready\nconnection: unchecked")

		return err
	}

	return checkTelegramConnection(ctx, env, token, api, stdout)
}

func checkTelegramConnection(ctx context.Context, env core.UserEnvironment, token string, api core.TelegramReceiverAPI, stdout io.Writer) (err error) {
	botID, err := api.CheckConnection(ctx, token)
	if err != nil {
		return err
	}

	bound, bindingErr := core.TelegramBoundBotID(env)
	if bindingErr != nil && !errors.Is(bindingErr, os.ErrNotExist) {
		return bindingErr
	}

	if bound != 0 && bound != botID {
		return errors.New("telegram bot identity changed; explicit setup --replace is required")
	}

	lock, err := core.AcquireTelegramReceiverLock(env, botID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	_, err = fmt.Fprintf(stdout, "telegram_doctor: connection_ready\nbot_id: %d\n", botID)

	return err
}

func printTelegramProjectReadiness(stdout io.Writer, operations service, config core.TelegramReceiverConfig) error {
	var ids []string

	for _, user := range config.AllowedUsers {
		for _, id := range user.ProjectIDs {
			if len(ids) < 20 && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
	}

	projects, err := operations.ListRegisteredProjects(ids)
	if err != nil {
		_, writeErr := fmt.Fprintln(stdout, "project_registry: unavailable\nmodel_connection: unchecked")

		return writeErr
	}

	for _, project := range projects {
		readiness := operations.InspectProjectReadiness(project)
		if _, err := fmt.Fprintf(stdout, "project: %s enabled=%t runtime_ready=%t model_credential_present=%t model_executable_present=%t\n",
			project.ID, project.Enabled, readiness.RuntimeReady, readiness.ModelCredentialPresent, readiness.ModelExecutablePresent); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(stdout, "registered_projects: %d/%d (at most 20 inspected)\nmodel_connection: unchecked\n", len(projects), len(ids))

	return err
}
