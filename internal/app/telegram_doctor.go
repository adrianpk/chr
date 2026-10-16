package app

import (
	"context"
	"errors"
	"fmt"
	"io"

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

	if !cmd.Bool("check-connection") {
		_, err := fmt.Fprintln(stdout, "telegram_doctor: local_configuration_ready\nconnection: unchecked")

		return err
	}

	botID, err := api.CheckConnection(ctx, token)
	if err != nil {
		return err
	}

	lock, err := core.AcquireTelegramReceiverLock(env, botID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	_, err = fmt.Fprintf(stdout, "telegram_doctor: connection_ready\nbot_id: %d\n", botID)

	return err
}
