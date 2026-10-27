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
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramReceiver struct {
	env            core.UserEnvironment
	api            telegramcore.ReceiverAPI
	store          *telegramcore.ReceiverStore
	operations     telegramReadService
	botID          int64
	token          string
	slots          *telegramActivationSlots
	processingLane *telegramcore.ProcessingLane
}

func (factory commandFactory) telegramServeCommand() *cli.Command {
	return &cli.Command{
		Name: "serve", Usage: "Run the foreground receiver with retained conversations and guarded operations",
		Flags: projectRegistryFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			env, err := newUserEnvironmentFromCommand(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			return telegramCommandFailure(serveTelegramReceiver(ctx, env, factory.deps.assetsFS, telegramcore.NewReceiverAPI()))
		},
	}
}

func serveTelegramReceiver(ctx context.Context, env core.UserEnvironment, assets fs.FS, api telegramcore.ReceiverAPI) (err error) {
	configurationLock, err := telegramcore.AcquireConfigurationLock(env)
	if err != nil {
		return err
	}

	resources := telegramReceiverResources{closers: []io.Closer{configurationLock}}
	defer func() { err = resources.finish(err) }()

	config, err := telegramcore.LoadReceiverConfig(env)
	if err != nil {
		return &telegramServiceError{code: "configuration_invalid", permanent: true}
	}

	token, err := telegramcore.ReceiverToken(env, config)
	if err != nil {
		return &telegramServiceError{code: "token_unavailable_or_invalid", permanent: true}
	}

	botID, err := api.CheckConnection(ctx, token)
	if err != nil {
		return err
	}

	botLock, err := telegramcore.AcquireReceiverLock(env, botID)
	if err != nil {
		return err
	}

	resources.closers = append(resources.closers, botLock)

	if err := telegramcore.BindReceiverIdentity(env, botID); err != nil {
		return err
	}

	store, err := telegramcore.NewReceiverStore(ctx, env, assets)
	if err != nil {
		return &telegramServiceError{code: "retained_schema_or_store_invalid", permanent: true}
	}

	resources.closers = append(resources.closers, store)

	if err := store.BindReceiver(ctx, botID, config, time.Now()); err != nil {
		return err
	}

	receiver := telegramReceiver{env: env, api: api, store: store, botID: botID, token: token, operations: newService(ctx, env, commandProjectContext{}, assets)}

	return receiver.run(ctx)
}

func (r telegramReceiver) run(ctx context.Context) error {
	r.slots = newTelegramActivationSlots()

	shared := newGuidedWorkHost(ctx)
	if operations, ok := r.operations.(service); ok {
		operations.guidedWorkHost = shared
		r.operations = operations
	}

	guided := &telegramGuidedHost{receiver: r, shared: shared}
	hosts := []func(context.Context) error{r.runIntake, newTelegramWorkHost(r).run, newTelegramApplicationHost(r).run, r.runDeliveries, guided.run}

	return runTelegramReceiverHosts(ctx, hosts, 30*time.Second)
}

func (r telegramReceiver) runIntake(ctx context.Context) error {
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

	if err := r.drainSynchronous(ctx); err != nil {
		return 0, err
	}

	state, err := r.store.PreparePoll(ctx, r.botID, time.Now())
	if err != nil {
		return 0, err
	}

	if state.Health == telegramcore.HealthBlocked || state.RetryAttempt >= 5 {
		return 0, &telegramServiceError{code: "poll_retry_blocked_or_exhausted", permanent: true}
	}

	if delay := time.Until(state.RetryAt); delay > 0 {
		return min(delay, time.Second), nil
	}

	if err := r.store.ReservePoll(ctx, r.botID, time.Now()); err != nil {
		return 0, err
	}

	return r.receiveBatch(ctx, config, state)
}

func (r telegramReceiver) receiveBatch(ctx context.Context, config telegramcore.ReceiverConfig, state telegramcore.ReceiverState) (time.Duration, error) {
	updates, err := r.api.Poll(ctx, r.token, state.NextUpdateID, config.PollTimeoutSeconds)
	if err != nil {
		return r.pollFailed(ctx, state, err)
	}

	config, err = r.currentConfig()
	if err != nil {
		return 0, err
	}

	for index, update := range updates {
		updates[index] = telegramcore.AuthorizeReceiverUpdate(config, r.botID, update)
	}

	if err := r.store.RetainBatch(ctx, r.botID, updates, time.Now()); err != nil {
		return 0, errors.Join(err, r.store.PollFailed(ctx, r.botID, err, time.Now()))
	}

	return 0, r.drainSynchronous(ctx)
}

func (r telegramReceiver) drainSynchronous(ctx context.Context) error {
	if r.slots != nil {
		return nil
	}

	return r.drain(ctx)
}

func (r telegramReceiver) pollFailed(ctx context.Context, prior telegramcore.ReceiverState, failure error) (time.Duration, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if err := r.store.PollFailed(ctx, r.botID, failure, time.Now()); err != nil {
		return 0, errors.Join(failure, err)
	}

	var apiErr *telegramcore.APIError

	if prior.RetryAttempt+1 >= 5 {
		return 0, &telegramServiceError{code: "poll_retry_exhausted", permanent: true}
	}

	if !errors.As(failure, &apiErr) || apiErr.Permanent {
		return 0, failure
	}

	return time.Second, nil
}

func (r telegramReceiver) currentConfig() (telegramcore.ReceiverConfig, error) {
	config, err := telegramcore.LoadReceiverConfig(r.env)
	if err != nil {
		return config, &telegramServiceError{code: "configuration_or_token_invalid", permanent: true}
	}

	token, err := telegramcore.ReceiverToken(r.env, config)
	if err != nil {
		return config, &telegramServiceError{code: "configuration_or_token_invalid", permanent: true}
	}

	if token != r.token {
		return config, &telegramServiceError{code: "credentials_changed_restart_required", permanent: true}
	}

	return config, nil
}
