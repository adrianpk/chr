package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func (factory commandFactory) telegramDoctorCommand() *cli.Command {
	return &cli.Command{
		Name: "doctor", Usage: "Check local project/provider readiness; optional getMe/webhook checks send no message",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{Name: "check-connection"}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return factory.runTelegramDoctor(ctx, cmd, core.NewTelegramReceiverAPI())
		},
	}
}

func (factory commandFactory) runTelegramDoctor(ctx context.Context, cmd *cli.Command, api core.TelegramReceiverAPI) (err error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	config, err := core.LoadTelegramReceiverConfig(env)
	if err != nil {
		return errors.New("telegram configuration is unavailable or invalid; run telegram setup")
	}

	token, err := core.TelegramReceiverToken(env, config)
	if err != nil {
		return errors.New("telegram token reference is unavailable or invalid; correct private configuration")
	}

	if err := printTelegramProviderReadiness(factory.io.Stdout); err != nil {
		return err
	}

	operations := newService(ctx, env, commandProjectContext{}, factory.deps.assetsFS)
	if err := printTelegramProjectReadiness(factory.io.Stdout, operations, config); err != nil {
		return err
	}

	if !cmd.Bool("check-connection") {
		_, err := fmt.Fprintln(factory.io.Stdout,
			"telegram_doctor: local_configuration_ready\nconnection: unchecked")

		return err
	}

	return telegramDoctorFailure(checkTelegramConnection(ctx, env, token, api, factory.io.Stdout))
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
		return &core.TelegramAPIError{Code: "bot_identity_changed", Permanent: true}
	}

	configurationLock, err := core.AcquireTelegramConfigurationLock(env)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, configurationLock.Close()) }()

	lock, err := core.AcquireTelegramReceiverLock(env, botID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	_, err = fmt.Fprintf(stdout, "telegram_doctor: connection_ready\nbot_id: %d\nremote_polling: unchecked; conflicts are detected by serve without takeover\n", botID)

	return err
}

func printTelegramProjectReadiness(stdout io.Writer, operations service, config core.TelegramReceiverConfig) error {
	ids := telegramDoctorProjectIDs(config)

	projects, err := operations.ListRegisteredProjects(ids)
	if err != nil {
		_, writeErr := fmt.Fprintln(stdout, "project_registry: unavailable\nmodel_connection: unchecked")

		return writeErr
	}

	if err := printTelegramMissingProjects(stdout, ids, projects); err != nil {
		return err
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

func telegramDoctorFailure(err error) error {
	if err == nil {
		return nil
	}

	var apiFailure *core.TelegramAPIError
	if errors.As(err, &apiFailure) {
		remedies := map[string]string{
			"bot_identity_changed":    "run setup --replace only after confirming the intended bot",
			"webhook_conflict":        "remove the webhook explicitly before starting this receiver",
			"local_receiver_conflict": "another local receiver owns this configuration or bot; stop it explicitly",
			"receiver_conflict":       "another poller or webhook owns remote delivery; stop the conflicting receiver explicitly",
			"credentials_or_access":   "correct the token and bot access",
		}

		remedy := remedies[apiFailure.Code]
		if remedy == "" {
			remedy = "check connectivity and retained retry state"
		}

		return fmt.Errorf("telegram doctor: %s; %s", telegramAPIFailureCode(apiFailure.Code), remedy)
	}

	return errors.New("telegram doctor: local_state_unavailable; inspect private configuration and retained bot binding")
}

func printTelegramProviderReadiness(stdout io.Writer) error {
	for _, boundary := range []struct{ name, variable string }{
		{"conversation", "CHR_CONVERSATION_GENERATION_BACKEND"}, {"dump", "CHR_DUMP_GENERATION_BACKEND"}, {"spec", "CHR_SPEC_GENERATION_BACKEND"}, {"plan", "CHR_PLAN_GENERATION_BACKEND"},
	} {
		provider := "codex"
		ready := false

		if openAIGenerationBackend(boundary.variable) {
			provider = "openai"
			ready = strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""
		} else {
			command := os.Getenv("CHR_CODEX_COMMAND")
			if command == "" {
				command = "codex"
			}

			_, err := exec.LookPath(command)
			ready = err == nil
		}

		if _, err := fmt.Fprintf(stdout, "model_boundary: %s provider=%s local_ready=%t connection=unchecked\n", boundary.name, provider, ready); err != nil {
			return err
		}
	}

	return nil
}

func telegramDoctorProjectIDs(config core.TelegramReceiverConfig) []string {
	var ids []string

	for _, user := range config.AllowedUsers {
		for _, id := range user.ProjectIDs {
			if len(ids) < 20 && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
	}

	return ids
}

func printTelegramMissingProjects(stdout io.Writer, ids []string, projects []projectcore.Project) error {
	for _, id := range ids {
		if !slices.ContainsFunc(projects, func(project projectcore.Project) bool { return project.ID == id }) {
			if _, err := fmt.Fprintf(stdout, "project: %s registered=false\n", id); err != nil {
				return err
			}
		}
	}

	return nil
}
