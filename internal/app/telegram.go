package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (factory commandFactory) telegramCommand() *cli.Command {
	return &cli.Command{
		Name: "telegram", Usage: "Configure and run the dedicated text receiver",
		Description: "Receive authorized text and buttons for retained project conversations and native work. Poll/model/delivery cycles retain five attempts across restart; pending input and delivery each stop at 1,024. No webhook takeover, arbitrary shell commands or implicit service activation.",
		Commands: []*cli.Command{
			factory.telegramSetupCommand(),
			factory.telegramDoctorCommand(),
			factory.telegramServeCommand(),
			factory.telegramRetryCommand(),
			{Name: "status", Usage: "Inspect retained health, retry cycles and queues without migration or network access", Flags: projectRegistryFlags(), Action: factory.telegramStatusAction},
		},
	}
}

func (factory commandFactory) telegramSetupCommand() *cli.Command {
	return &cli.Command{
		Name: "setup", Usage: "Configure private receiver bindings without starting the receiver",
		Flags: append(projectRegistryFlags(),
			&cli.Int64Flag{Name: "user-id", Required: true},
			&cli.StringFlag{Name: "operator-id", Required: true},
			&cli.StringSliceFlag{Name: "project-id", Required: true},
			&cli.Int64SliceFlag{Name: "chat-id"},
			&cli.Int64SliceFlag{Name: "topic-id"},
			&cli.StringFlag{Name: "bot-token-env", Usage: "Read token from this environment reference instead of masked terminal input"},
			&cli.IntFlag{Name: "poll-timeout-seconds", Value: 30},
			&cli.BoolFlag{Name: "install-service", Usage: "Install a user service for this executable without enabling or starting it"},
			&cli.BoolFlag{Name: "replace", Usage: "Explicitly replace differing receiver configuration or credentials"},
		),
		Action: func(_ context.Context, cmd *cli.Command) error {
			return factory.runTelegramSetup(cmd)
		},
	}
}

func (factory commandFactory) runTelegramSetup(cmd *cli.Command) (err error) {
	if cmd.NArg() != 0 {
		return errors.New("telegram setup does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	config := telegramSetupConfig(cmd)
	if err := config.Validate(); err != nil {
		return err
	}

	lock, err := telegramcore.AcquireTelegramConfigurationLock(env)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	if err := guardTelegramSetupReplacement(env, config, cmd.Bool("replace")); err != nil {
		return err
	}

	if err := factory.storeTelegramSetupToken(cmd, env, config); err != nil {
		return err
	}

	if err := telegramcore.SaveTelegramReceiverConfig(env, config, cmd.Bool("replace")); err != nil {
		return err
	}

	return factory.finishTelegramSetup(cmd, env)
}

func (factory commandFactory) finishTelegramSetup(cmd *cli.Command, env core.UserEnvironment) (err error) {
	if cmd.Bool("replace") {
		if err := telegramcore.ResetTelegramReceiverIdentity(env); err != nil {
			return err
		}
	}

	if cmd.Bool("install-service") {
		if err := factory.installTelegramUserService(cmd, env); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(factory.io.Stdout,
		"telegram_setup: configured")

	return err
}

func guardTelegramSetupReplacement(env core.UserEnvironment, config telegramcore.TelegramReceiverConfig, replace bool) error {
	// Check replacement authority before prompting or storing any credential.
	prior, loadErr := telegramcore.LoadTelegramReceiverConfig(env)
	if !replace && !errors.Is(loadErr, os.ErrNotExist) && (loadErr != nil || prior.Digest() != config.Digest()) {
		return errors.New("telegram configuration exists; use --replace to change it")
	}

	return nil
}

func telegramSetupConfig(cmd *cli.Command) telegramcore.TelegramReceiverConfig {
	name := cmd.String("bot-token-env")
	if name == "" {
		name = "CHR_TELEGRAM_BOT_TOKEN"
	}

	return telegramcore.TelegramReceiverConfig{
		Enabled: true, BotTokenEnv: name, PollTimeoutSeconds: cmd.Int("poll-timeout-seconds"),
		AllowedChatIDs: cmd.Int64Slice("chat-id"), AllowedTopicIDs: cmd.Int64Slice("topic-id"),
		AllowedUsers: []telegramcore.TelegramAllowedUser{{TelegramUserID: cmd.Int64("user-id"), OperatorID: cmd.String("operator-id"), ProjectIDs: cmd.StringSlice("project-id")}},
	}
}

func (factory commandFactory) storeTelegramSetupToken(cmd *cli.Command, env core.UserEnvironment, config telegramcore.TelegramReceiverConfig) error {
	if cmd.IsSet("bot-token-env") {
		_, err := telegramcore.TelegramReceiverToken(env, config)

		return err
	}

	file, ok := factory.io.Stdin.(*os.File)
	if !ok || !term.IsTerminal(file.Fd()) {
		return errors.New("masked token entry requires a terminal; use --bot-token-env for noninteractive setup")
	}

	if _, err := fmt.Fprint(factory.io.Stdout,
		"Bot token: "); err != nil {
		return err
	}

	token, err := term.ReadPassword(file.Fd())
	if err != nil {
		return errors.New("read masked telegram token failed")
	}

	defer clear(token)

	if _, err := fmt.Fprintln(factory.io.Stdout); err != nil {
		return err
	}

	return telegramcore.SaveTelegramReceiverToken(env, config, string(token), cmd.Bool("replace"))
}
