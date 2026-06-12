package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

var notifyTelegramSender = core.NewTelegramSender(http.DefaultClient)

func newNotifyCommand(stdout io.Writer, deps appDependencies) *cli.Command {
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
			newNotifyTelegramPollCommand(stdout),
			newNotifyTelegramTestCommand(stdout),
		},
	}
}

func newNotifyTelegramPollCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "poll",
		Usage: "Parse Telegram updates into remote command handoff records",
		Flags: append(projectRegistryFlags(),
			&cli.StringFlag{
				Name:  "fixture",
				Usage: "Telegram getUpdates JSON fixture path",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print parsed command records without persisting them",
			},
		),
		Action: notifyTelegramPollAction(stdout),
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

func notifyTelegramPollAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runNotifyTelegramPoll(ctx, cmd, stdout)
	}
}

func runNotifyTelegramPoll(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify telegram poll does not accept positional arguments")
	}

	if cmd.String("fixture") == "" {
		return fmt.Errorf("notify telegram poll requires --fixture")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	records, err := loadNotifyTelegramPollFixture(cmd.String("fixture"))
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		deps, err := appDependenciesFromCommand(cmd)
		if err != nil {
			return err
		}

		err = persistRemoteCommandRecords(ctx, env, records, deps.assetsFS)
		if err != nil {
			return err
		}
	}

	return printNotifyTelegramPollResult(stdout, records, cmd.Bool("dry-run"))
}

func runNotifyStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify status does not accept positional arguments")
	}

	_, projectContext, config, err := loadNotifyProjectConfig(ctx, cmd)
	if err != nil {
		return err
	}

	return printNotifyStatus(stdout, projectContext, config)
}

func loadNotifyTelegramPollFixture(path string) ([]core.RemoteCommandRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open telegram update fixture: %w", err)
	}

	defer func() {
		//nolint:errcheck // Decode errors are returned by DecodeTelegramUpdates.
		_ = file.Close()
	}()

	updates, err := core.DecodeTelegramUpdates(file)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	records := make([]core.RemoteCommandRecord, 0, len(updates))
	for _, update := range updates {
		record, err := core.RemoteCommandRecordFromTelegramUpdate(update, now)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil
}

func persistRemoteCommandRecords(
	ctx context.Context,
	env core.UserEnvironment,
	records []core.RemoteCommandRecord,
	assetsFS fs.FS,
) error {
	repository, err := core.NewRemoteCommandRecordRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = repository.Close()
	}()

	for _, record := range records {
		err = repository.CreateRemoteCommandRecord(ctx, record)
		if err != nil {
			return err
		}
	}

	return nil
}

func printNotifyTelegramPollResult(
	stdout io.Writer,
	records []core.RemoteCommandRecord,
	dryRun bool,
) error {
	lines := []string{
		"telegram_poll: parsed",
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("records: %d", len(records)),
	}

	for _, record := range records {
		lines = append(lines, fmt.Sprintf(
			"remote_command: %s\t%s\t%s\t%s",
			record.Request.ID,
			record.Status,
			record.Request.Intent,
			record.HandoffCommand,
		))
		if record.RejectionReason != "" {
			lines = append(lines, "rejection: "+record.RejectionReason)
		}
	}

	return printNotifyLines(stdout, lines)
}

func runNotifyTelegramTest(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify telegram test does not accept positional arguments")
	}

	env, projectContext, config, err := loadNotifyProjectConfig(ctx, cmd)
	if err != nil {
		return err
	}

	err = core.ValidateNotificationConfig(config.Notifications)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		deps, err := appDependenciesFromCommand(cmd)
		if err != nil {
			return err
		}

		return runNotifyTelegramTestDispatch(ctx, stdout, env, projectContext, config.Notifications.Telegram, deps.assetsFS)
	}

	return printNotifyTelegramTestPlan(stdout, projectContext, config.Notifications.Telegram)
}

func loadNotifyProjectConfig(
	ctx context.Context,
	cmd *cli.Command,
) (core.UserEnvironment, commandProjectContext, core.WorkspaceConfig, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return nil, commandProjectContext{}, core.WorkspaceConfig{}, err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return nil, commandProjectContext{}, core.WorkspaceConfig{}, err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return nil, commandProjectContext{}, core.WorkspaceConfig{}, err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return nil, commandProjectContext{}, core.WorkspaceConfig{}, err
	}

	return env, projectContext, config, nil
}

func printNotifyStatus(stdout io.Writer, projectContext commandProjectContext, config core.WorkspaceConfig) error {
	telegram := config.Notifications.Telegram
	telegramStatus := "disabled"

	err := core.ValidateNotificationConfig(config.Notifications)
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
	telegram core.TelegramNotificationConfig,
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

func runNotifyTelegramTestDispatch(
	ctx context.Context,
	stdout io.Writer,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	telegram core.TelegramNotificationConfig,
	assetsFS fs.FS,
) error {
	now := time.Now().UTC()
	event := notifyTelegramTestEvent(projectContext, now)
	record := core.NotificationAuditRecord{
		Event:   event,
		Channel: "telegram",
	}

	repository, err := core.NewNotificationEventRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.CreateNotificationEvent(ctx, record)
	if err != nil {
		return err
	}

	message := core.RenderTelegramNotificationMessage(event)
	message.ChatID = telegram.ChatID

	token := os.Getenv(telegram.BotTokenEnv)
	result, sendErr := notifyTelegramSender.SendMessage(ctx, token, message)
	finishedAt := time.Now().UTC()

	if sendErr != nil {
		record.Event.Status = core.NotificationStatusFailed
		record.Event.FailureReason = sendErr.Error()
		record.Event.UpdatedAt = finishedAt

		updateErr := repository.UpdateNotificationEvent(ctx, record)
		if updateErr != nil {
			return updateErr
		}

		return sendErr
	}

	record.Event.Status = core.NotificationStatusDelivered
	record.Event.DeliveredAt = &finishedAt
	record.Event.UpdatedAt = finishedAt

	err = repository.UpdateNotificationEvent(ctx, record)
	if err != nil {
		return err
	}

	return printNotifyTelegramTestSent(stdout, projectContext, record, result)
}

func notifyTelegramTestEvent(projectContext commandProjectContext, now time.Time) core.NotificationEvent {
	return core.NotificationEvent{
		ID:        fmt.Sprintf("telegram-test-%d", now.UnixNano()),
		ProjectID: projectContext.ProjectID,
		Type:      core.NotificationEventManualTest,
		Severity:  core.NotificationSeverityInfo,
		Status:    core.NotificationStatusPending,
		Source:    core.NotificationSourceManual,
		Payload: core.NotificationPayload{
			Summary: "ChrysoCoord Telegram notification test for " + projectContext.ProjectID,
			Detail:  "Project directory: " + projectContext.ProjectDir,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func printNotifyTelegramTestSent(
	stdout io.Writer,
	projectContext commandProjectContext,
	record core.NotificationAuditRecord,
	result core.TelegramSendResult,
) error {
	return printNotifyLines(stdout, []string{
		"telegram_test: sent",
		"dry_run: false",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"notification_id: " + record.Event.ID,
		fmt.Sprintf("status_code: %d", result.StatusCode),
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
