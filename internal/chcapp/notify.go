package chcapp

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

func newNotifyCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "notify",
		Usage: "Inspect and test notification setup",
		Commands: []*cli.Command{
			newNotifyStatusCommand(stdout),
			newNotifyTelegramCommand(stdout),
		},
	}
}

func newNotifyStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Report notification configuration status",
		Flags: append(projectRegistryFlags(), &cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
		}),
		Action: notifyStatusAction(stdout),
	}
}

func newNotifyTelegramCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "telegram",
		Usage: "Inspect and test Telegram notification setup",
		Commands: []*cli.Command{
			newNotifyTelegramTestCommand(stdout),
		},
	}
}

func newNotifyTelegramTestCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "Render a Telegram notification setup test",
		Flags: append(projectRegistryFlags(),
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print the Telegram test payload without sending it",
			},
		),
		Action: notifyTelegramTestAction(stdout),
	}
}

func notifyStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runNotifyStatus(ctx, cmd, stdout)
	}
}

func notifyTelegramTestAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runNotifyTelegramTest(ctx, cmd, stdout)
	}
}

func runNotifyStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify status does not accept positional arguments")
	}

	projectContext, config, err := loadNotifyProjectConfig(ctx, cmd)
	if err != nil {
		return err
	}

	return printNotifyStatus(stdout, projectContext, config)
}

func runNotifyTelegramTest(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify telegram test does not accept positional arguments")
	}

	if !cmd.Bool("dry-run") {
		return fmt.Errorf("notify telegram test requires --dry-run until Telegram dispatch is implemented")
	}

	projectContext, config, err := loadNotifyProjectConfig(ctx, cmd)
	if err != nil {
		return err
	}

	err = chc.ValidateNotificationConfig(config.Notifications)
	if err != nil {
		return err
	}

	return printNotifyTelegramTestPlan(stdout, projectContext, config.Notifications.Telegram)
}

func loadNotifyProjectConfig(
	ctx context.Context,
	cmd *cli.Command,
) (commandProjectContext, chc.WorkspaceConfig, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return commandProjectContext{}, chc.WorkspaceConfig{}, err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return commandProjectContext{}, chc.WorkspaceConfig{}, err
	}

	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return commandProjectContext{}, chc.WorkspaceConfig{}, err
	}

	return projectContext, config, nil
}

func printNotifyStatus(stdout io.Writer, projectContext commandProjectContext, config chc.WorkspaceConfig) error {
	telegram := config.Notifications.Telegram
	telegramStatus := "disabled"

	err := chc.ValidateNotificationConfig(config.Notifications)
	if err != nil {
		telegramStatus = "invalid"
	} else if telegram.Enabled {
		telegramStatus = "ready"
	}

	lines := []string{
		"notify: status",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		fmt.Sprintf("telegram_enabled: %t", telegram.Enabled),
		"telegram_status: " + telegramStatus,
		"telegram_bot_token_env: " + telegram.BotTokenEnv,
		"telegram_chat_id: " + telegram.ChatID,
	}

	if err != nil {
		lines = append(lines, "telegram_error: "+err.Error())
	}

	return printNotifyLines(stdout, lines)
}

func printNotifyTelegramTestPlan(
	stdout io.Writer,
	projectContext commandProjectContext,
	telegram chc.TelegramNotificationConfig,
) error {
	return printNotifyLines(stdout, []string{
		"telegram_test: planned",
		"dry_run: true",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"bot_token_env: " + telegram.BotTokenEnv,
		"chat_id: " + telegram.ChatID,
		"message: ChrysoCoord Telegram notification test for " + projectContext.ProjectID,
	})
}

func printNotifyLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write notify result: %w", err)
		}
	}

	return nil
}
