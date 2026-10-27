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

	"forge.adrianpk.com/chr/internal/core"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

var notifyTelegramSender = telegramcore.NewTelegramSender(http.DefaultClient)

func (factory commandFactory) notifyCommand() *cli.Command {
	return &cli.Command{
		Name:  "notify",
		Usage: "Inspect and test notification setup",
		Commands: []*cli.Command{
			factory.notifyStatusCommand(),
			factory.notifyTelegramCommand(),
		},
	}
}

func (factory commandFactory) notifyStatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Report notification configuration status",
		Flags: append(projectRegistryFlags(), &cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
		}),
		Action: factory.runNotifyStatus,
	}
}

func (factory commandFactory) notifyTelegramCommand() *cli.Command {
	return &cli.Command{
		Name:  "telegram",
		Usage: "Inspect and test Telegram notification setup",
		Commands: []*cli.Command{
			factory.notifyTelegramPollCommand(),
			factory.notifyTelegramTestCommand(),
		},
	}
}

func (factory commandFactory) notifyTelegramPollCommand() *cli.Command {
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
		Action: factory.runNotifyTelegramPoll,
	}
}

func (factory commandFactory) notifyTelegramTestCommand() *cli.Command {
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
		Action: factory.runNotifyTelegramTest,
	}
}

func (factory commandFactory) runNotifyTelegramPoll(ctx context.Context, cmd *cli.Command) error {
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
		err = persistRemoteCommandRecords(ctx, env, records, factory.deps.assetsFS)
		if err != nil {
			return err
		}
	}

	return printNotifyTelegramPollResult(factory.io.Stdout, records, cmd.Bool("dry-run"))
}

func (factory commandFactory) runNotifyStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify status does not accept positional arguments")
	}

	notifyProject, err := factory.loadNotifyProjectConfig(ctx, cmd)
	projectContext, config := notifyProject.Project, notifyProject.Config

	if err != nil {
		return err
	}

	return printNotifyStatus(factory.io.Stdout, projectContext, config)
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

	updates, err := telegramcore.DecodeTelegramUpdates(file)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	records := make([]core.RemoteCommandRecord, 0, len(updates))
	for _, update := range updates {
		record, err := remoteCommandRecordFromTelegramUpdate(update, now)
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

func (factory commandFactory) runNotifyTelegramTest(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("notify telegram test does not accept positional arguments")
	}

	notifyProject, err := factory.loadNotifyProjectConfig(ctx, cmd)
	env, projectContext, config := notifyProject.Environment, notifyProject.Project, notifyProject.Config

	if err != nil {
		return err
	}

	err = notifycore.ValidateConfig(config.Notifications)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		return factory.runNotifyTelegramTestDispatch(ctx, env, projectContext, config.Notifications.Telegram)
	}

	return printNotifyTelegramTestPlan(factory.io.Stdout, projectContext, config.Notifications.Telegram)
}

func (factory commandFactory) loadNotifyProjectConfig(ctx context.Context, cmd *cli.Command) (notifyProjectConfig, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return notifyProjectConfig{Environment: nil, Project: commandProjectContext{}, Config: core.WorkspaceConfig{}}, err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return notifyProjectConfig{Environment: nil, Project: commandProjectContext{}, Config: core.WorkspaceConfig{}}, err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return notifyProjectConfig{Environment: nil, Project: commandProjectContext{}, Config: core.WorkspaceConfig{}}, err
	}

	return notifyProjectConfig{Environment: env, Project: projectContext, Config: config}, nil
}

func printNotifyStatus(stdout io.Writer, projectContext commandProjectContext, config core.WorkspaceConfig) error {
	telegram := config.Notifications.Telegram
	telegramStatus := "disabled"

	err := notifycore.ValidateConfig(config.Notifications)
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
	telegram notifycore.TelegramConfig,
) error {
	return printNotifyLines(stdout, []string{
		"telegram_test: planned",
		"dry_run: true",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"bot_token_env: " + telegram.BotTokenEnv,
		"chat_id: " + telegram.ChatID,
		"message: Telegram notification test for " + projectContext.ProjectName,
	})
}

func (factory commandFactory) runNotifyTelegramTestDispatch(ctx context.Context, env core.UserEnvironment, projectContext commandProjectContext, telegram notifycore.TelegramConfig) error {
	now := time.Now().UTC()
	event := notifyTelegramTestEvent(projectContext, now)
	record := notifycore.AuditRecord{
		Event:   event,
		Channel: "telegram",
	}

	repository, err := notifycore.NewRepository(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	target, err := createTelegramNotification(ctx, repository, record, projectContext, telegram)
	if err != nil {
		return err
	}

	result, err := dispatchTelegramNotification(ctx, repository, record, projectContext, target)
	if err != nil {
		return err
	}

	return printNotifyTelegramTestSent(factory.io.Stdout, projectContext, record, result)
}

func notifyTelegramTestEvent(projectContext commandProjectContext, now time.Time) notifycore.Event {
	return notifycore.Event{
		ID:        fmt.Sprintf("telegram-test-%d", now.UnixNano()),
		ProjectID: projectContext.ProjectID,
		Type:      notifycore.EventManualTest,
		Severity:  notifycore.SeverityInfo,
		Status:    notifycore.StatusPending,
		Source:    notifycore.SourceManual,
		Payload: notifycore.Payload{
			Summary: "chr Telegram notification test for " + projectContext.ProjectID,
			Detail:  "Project directory: " + projectContext.ProjectDir,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func printNotifyTelegramTestSent(
	stdout io.Writer,
	projectContext commandProjectContext,
	record notifycore.AuditRecord,
	result telegramcore.TelegramSendResult,
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

type notifyProjectConfig struct {
	Environment core.UserEnvironment
	Project     commandProjectContext
	Config      core.WorkspaceConfig
}
