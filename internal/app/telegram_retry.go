package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) telegramRetryCommand() *cli.Command {
	return &cli.Command{
		Name: "retry", Usage: "Open one exhausted transport cycle while the receiver is stopped; preserve model and work budgets",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{Name: "poll", Usage: "Retry exhausted or blocked polling"}, &cli.StringFlag{Name: "delivery", Usage: "Retry one exact blocked or exhausted delivery ID from status"}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() != 0 || cmd.Bool("poll") == (cmd.String("delivery") != "") {
				return errors.New("telegram retry requires exactly --poll or --delivery <id>")
			}

			env, err := newUserEnvironmentFromCommand(cmd)
			if err != nil {
				return errors.New("telegram environment is unavailable")
			}

			botID, err := core.TelegramBoundBotID(env)
			if err != nil {
				return errors.New("telegram bot is unbound or invalid; run telegram doctor")
			}

			actor := "local-uid-" + strconv.Itoa(os.Getuid())

			request := core.TelegramTransportRetry{BotID: botID, Poll: cmd.Bool("poll"), DeliveryID: cmd.String("delivery"), Actor: actor}
			if err := core.RetryLocalTelegramTransport(ctx, env, factory.deps.assetsFS, request, time.Now()); err != nil {
				var failure *core.TelegramAPIError
				if errors.As(err, &failure) && failure.Code == "local_receiver_conflict" {
					return errors.New("telegram retry: local_receiver_conflict; stop the receiver before retrying")
				}

				return errors.New("telegram retry refused; select an exhausted cycle from status and run doctor before retrying")
			}

			_, err = fmt.Fprintf(factory.io.Stdout,
				"telegram_retry: cycle_opened\nactor: %s\nreceiver_activation: unchanged\n", actor)

			return err
		},
	}
}
