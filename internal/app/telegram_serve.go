package app

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramReceiver struct {
	env        core.UserEnvironment
	api        core.TelegramReceiverAPI
	store      *core.TelegramReceiverStore
	operations telegramReadService
	botID      int64
	token      string
}

func newTelegramServeCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name: "serve", Usage: "Run the foreground read-only receiver with durable intake and delivery",
		Flags: projectRegistryFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			env, err := newUserEnvironmentFromCommand(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			return serveTelegramReceiver(ctx, env, deps.assetsFS, core.NewTelegramReceiverAPI())
		},
	}
}

func serveTelegramReceiver(ctx context.Context, env core.UserEnvironment, assets fs.FS, api core.TelegramReceiverAPI) (err error) {
	configurationLock, err := core.AcquireTelegramConfigurationLock(env)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, configurationLock.Close()) }()

	config, err := core.LoadTelegramReceiverConfig(env)
	if err != nil {
		return err
	}

	token, err := core.TelegramReceiverToken(env, config)
	if err != nil {
		return err
	}

	botID, err := api.CheckConnection(ctx, token)
	if err != nil {
		return err
	}

	botLock, err := core.AcquireTelegramReceiverLock(env, botID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, botLock.Close()) }()

	if err := core.BindTelegramReceiverIdentity(env, botID); err != nil {
		return err
	}

	store, err := core.NewTelegramReceiverStore(ctx, env, assets)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()

	if err := store.BindReceiver(ctx, botID, config, time.Now()); err != nil {
		return err
	}

	receiver := telegramReceiver{env: env, api: api, store: store, botID: botID, token: token, operations: newService(ctx, env, commandProjectContext{}, assets)}

	return receiver.run(ctx)
}

func (r telegramReceiver) run(ctx context.Context) error {
	for ctx.Err() == nil {
		delay, err := r.step(ctx)
		if err != nil && ctx.Err() == nil {
			return err
		}

		timer := time.NewTimer(max(250*time.Millisecond, delay))
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
	}

	return nil
}

func (r telegramReceiver) step(ctx context.Context) (time.Duration, error) {
	config, err := r.currentConfig()
	if err != nil {
		return 0, err
	}

	if err := r.drain(ctx); err != nil {
		return 0, err
	}

	state, err := r.store.Receiver(ctx, r.botID)
	if err != nil {
		return 0, err
	}

	if state.Health == "blocked" || state.RetryAttempt >= 5 {
		return 0, errors.New("telegram poll retry is blocked or exhausted; retained state requires local recovery")
	}

	if delay := time.Until(state.RetryAt); delay > 0 {
		return min(delay, time.Second), nil
	}

	if err := r.store.ReservePoll(ctx, r.botID, time.Now()); err != nil {
		return 0, err
	}

	return r.receiveBatch(ctx, config, state)
}

func (r telegramReceiver) receiveBatch(ctx context.Context, config core.TelegramReceiverConfig, state core.TelegramReceiverState) (time.Duration, error) {
	updates, err := r.api.Poll(ctx, r.token, state.NextUpdateID, config.PollTimeoutSeconds)
	if err != nil {
		return r.pollFailed(ctx, state, err)
	}

	config, err = r.currentConfig()
	if err != nil {
		return 0, err
	}

	for index, update := range updates {
		updates[index] = core.AuthorizeTelegramReceiverUpdate(config, r.botID, update)
	}

	if err := r.store.RetainBatch(ctx, r.botID, updates, time.Now()); err != nil {
		return 0, errors.Join(err, r.store.PollFailed(ctx, r.botID, err, time.Now()))
	}

	return 0, r.drain(ctx)
}

func (r telegramReceiver) pollFailed(ctx context.Context, prior core.TelegramReceiverState, failure error) (time.Duration, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if err := r.store.PollFailed(ctx, r.botID, failure, time.Now()); err != nil {
		return 0, errors.Join(failure, err)
	}

	var apiErr *core.TelegramAPIError
	if !errors.As(failure, &apiErr) || apiErr.Permanent || prior.RetryAttempt+1 >= 5 {
		return 0, failure
	}

	return time.Second, nil
}

func (r telegramReceiver) currentConfig() (core.TelegramReceiverConfig, error) {
	config, err := core.LoadTelegramReceiverConfig(r.env)
	if err != nil {
		return config, err
	}

	token, err := core.TelegramReceiverToken(r.env, config)
	if err != nil {
		return config, err
	}

	if token != r.token {
		return config, errors.New("telegram credentials changed; restart is required to verify the bot identity")
	}

	return config, nil
}
