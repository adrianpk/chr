package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/x/term"
	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func newTelegramCommand(stdin io.Reader, stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name: "telegram", Usage: "Configure and run the dedicated text receiver",
		Commands: []*cli.Command{
			newTelegramSetupCommand(stdin, stdout),
			newTelegramDoctorCommand(stdout),
			{Name: "status", Usage: "Inspect local receiver configuration", Flags: projectRegistryFlags(), Action: telegramStatusAction(stdout)},
		},
	}
}

func newTelegramSetupCommand(stdin io.Reader, stdout io.Writer) *cli.Command {
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
			&cli.BoolFlag{Name: "replace", Usage: "Explicitly replace differing receiver configuration or credentials"},
		),
		Action: func(_ context.Context, cmd *cli.Command) error {
			return runTelegramSetup(cmd, stdin, stdout)
		},
	}
}

func runTelegramSetup(cmd *cli.Command, stdin io.Reader, stdout io.Writer) (err error) {
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

	lock, err := core.AcquireTelegramConfigurationLock(env)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	// Check replacement authority before prompting or storing any credential.
	prior, loadErr := core.LoadTelegramReceiverConfig(env)
	if !cmd.Bool("replace") && !errors.Is(loadErr, os.ErrNotExist) && (loadErr != nil || prior.Digest() != config.Digest()) {
		return errors.New("telegram configuration exists; use --replace to change it")
	}

	if err := storeTelegramSetupToken(cmd, env, config, stdin, stdout); err != nil {
		return err
	}

	if err := core.SaveTelegramReceiverConfig(env, config, cmd.Bool("replace")); err != nil {
		return err
	}

	if cmd.Bool("replace") {
		if err := core.ResetTelegramReceiverIdentity(env); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(stdout, "telegram_setup: configured")

	return err
}

func telegramSetupConfig(cmd *cli.Command) core.TelegramReceiverConfig {
	name := cmd.String("bot-token-env")
	if name == "" {
		name = "CHR_TELEGRAM_BOT_TOKEN"
	}

	return core.TelegramReceiverConfig{
		Enabled: true, BotTokenEnv: name, PollTimeoutSeconds: cmd.Int("poll-timeout-seconds"),
		AllowedChatIDs: cmd.Int64Slice("chat-id"), AllowedTopicIDs: cmd.Int64Slice("topic-id"),
		AllowedUsers: []core.TelegramAllowedUser{{TelegramUserID: cmd.Int64("user-id"), OperatorID: cmd.String("operator-id"), ProjectIDs: cmd.StringSlice("project-id")}},
	}
}

func storeTelegramSetupToken(cmd *cli.Command, env core.UserEnvironment, config core.TelegramReceiverConfig, stdin io.Reader, stdout io.Writer) error {
	if cmd.IsSet("bot-token-env") {
		_, err := core.TelegramReceiverToken(env, config)

		return err
	}

	file, ok := stdin.(*os.File)
	if !ok || !term.IsTerminal(file.Fd()) {
		return errors.New("masked token entry requires a terminal; use --bot-token-env for noninteractive setup")
	}

	if _, err := fmt.Fprint(stdout, "Bot token: "); err != nil {
		return err
	}

	token, err := term.ReadPassword(file.Fd())
	if err != nil {
		return errors.New("read masked telegram token failed")
	}

	defer clear(token)
	if _, err := fmt.Fprintln(stdout); err != nil {
		return err
	}

	return core.SaveTelegramReceiverToken(env, config, string(token), cmd.Bool("replace"))
}

func telegramStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		env, err := newUserEnvironmentFromCommand(cmd)
		if err != nil {
			return err
		}

		config, err := core.LoadTelegramReceiverConfig(env)
		if err != nil {
			return err
		}

		_, tokenErr := core.TelegramReceiverToken(env, config)
		_, err = fmt.Fprintf(stdout, "telegram: configured\ntoken_available: %t\npoll_timeout_seconds: %d\nallowed_users: %d\nconfig: %s\n", tokenErr == nil, config.PollTimeoutSeconds, len(config.AllowedUsers), filepath.Join(env.ConfigDir(), "telegram.yml"))

		return err
	}
}
