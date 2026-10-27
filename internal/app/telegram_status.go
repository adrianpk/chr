package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (factory commandFactory) telegramStatusAction(ctx context.Context, cmd *cli.Command) error {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	config, err := telegramcore.LoadReceiverConfig(env)
	if err != nil {
		return errors.New("telegram configuration is unavailable or invalid; run telegram setup")
	}

	_, tokenErr := telegramcore.ReceiverToken(env, config)

	_, err = fmt.Fprintf(factory.io.Stdout,
		"telegram: configured\ntoken_available: %t\npoll_timeout_seconds: %d\nallowed_users: %d\nconfig: %s\n", tokenErr == nil, config.PollTimeoutSeconds, len(config.AllowedUsers), filepath.Join(env.ConfigDir(), "telegram.yml"))
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(factory.io.Stdout,
		"installed_build: %s\n", telegramInstalledBuild()); err != nil {
		return err
	}

	return printTelegramLocalState(ctx, env, factory.io.Stdout)
}

func printTelegramLocalState(ctx context.Context, env core.UserEnvironment, stdout io.Writer) error {
	botID, err := telegramcore.BoundBotID(env)
	if errors.Is(err, os.ErrNotExist) {
		_, err := fmt.Fprintln(stdout, "receiver: unbound")

		return err
	}

	if err != nil {
		return errors.New("telegram retained bot identity is invalid; run telegram doctor")
	}

	diagnostics, err := telegramcore.InspectReceiverDiagnostics(ctx, env, botID)
	if err != nil {
		return errors.New("telegram retained state is unavailable or requires migration; run telegram doctor")
	}

	state := diagnostics.Receiver

	cursor := "unset"
	if state.NextUpdateID != nil {
		cursor = fmt.Sprint(*state.NextUpdateID)
	}

	_, err = fmt.Fprintf(stdout, "receiver: %s\nbot_id: %d\ncursor_epoch: %d\nnext_update_id: %s\nlast_poll_at: %s\nretry_attempt: %d\nretry_at: %s\nlast_error_code: %s\n",
		telegramDiagnosticAtom(string(state.Health)), botID, state.Epoch, cursor, telegramDiagnosticTime(state.LastPollAt), state.RetryAttempt, telegramDiagnosticTime(state.RetryAt), telegramDiagnosticAtom(state.LastErrorCode))
	if err != nil {
		return err
	}

	return printTelegramDiagnostics(stdout, diagnostics)
}

func telegramInstalledBuild() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	revision := "development"

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision = telegramDiagnosticAtom(setting.Value)
		}

		if setting.Key == "vcs.modified" && setting.Value == "true" {
			revision += "+modified"
		}
	}

	return revision + " (" + telegramDiagnosticAtom(info.GoVersion) + ")"
}

func telegramDiagnosticAtom(value string) string {
	if value == "" {
		return "none"
	}

	if !regexp.MustCompile(`^[A-Za-z0-9_.+-]{1,128}$`).MatchString(value) {
		return "invalid"
	}

	return value
}

func telegramDiagnosticTime(value time.Time) string {
	if value.IsZero() {
		return "unset"
	}

	return value.UTC().Format(time.RFC3339)
}

func printTelegramDiagnostics(stdout io.Writer, diagnostics telegramcore.ReceiverDiagnostics) error {
	state := diagnostics.Receiver
	if _, err := fmt.Fprintf(stdout, "retry_cycle: %d\nlifetime_poll_attempts: %d\nretry_actor: %s\nretry_started_at: %s\nretention_gap_at: %s\nremote_gap_possible: %t\nactive_conversations: %d\npending_updates: %d\npending_deliveries: %d\nexhausted_models: %d\ndelivery_detail_limit: 20\n", state.RetryCycle, state.TotalRetryAttempts, telegramDiagnosticAtom(state.RetryActor), telegramDiagnosticTime(state.RetryStartedAt), telegramDiagnosticTime(state.RetentionGapAt), !state.RetentionGapAt.IsZero(), diagnostics.ActiveConversations, diagnostics.PendingUpdates, diagnostics.PendingDeliveries, diagnostics.ExhaustedModels); err != nil {
		return err
	}

	for _, item := range diagnostics.Deliveries {
		if _, err := fmt.Fprintf(stdout, "delivery: %s state=%s attempt=%d cycle=%d lifetime_attempts=%d retry_at=%s error=%s retry_actor=%s retry_started_at=%s\n", telegramDiagnosticAtom(item.ID), telegramDiagnosticAtom(string(item.State)), item.Attempts, item.Cycle, item.TotalAttempts, telegramDiagnosticTime(item.RetryAt), telegramDiagnosticAtom(item.ErrorCode), telegramDiagnosticAtom(item.RetryActor), telegramDiagnosticTime(item.RetryStartedAt)); err != nil {
			return err
		}
	}

	return nil
}
