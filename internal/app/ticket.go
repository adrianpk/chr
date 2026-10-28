package app

import (
	"github.com/urfave/cli/v3"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

var (
	reviewProviderDiagnosticRunner workflowcore.ReviewProviderDiagnosticRunner = workflowcore.OSReviewProviderDiagnosticRunner{}
	reviewRemoteCommandRunner      workflowcore.ReviewRemoteCommandRunner      = workflowcore.OSReviewRemoteCommandRunner{}
)

func (factory commandFactory) ticketCommand() *cli.Command {
	return &cli.Command{
		Name:  "ticket",
		Usage: "Manage workspace tickets",
		Commands: []*cli.Command{
			factory.ticketApplyResultCommand(),
			factory.ticketBoardCommand(),
			factory.ticketCommentCommand(),
			factory.ticketListCommand(),
			factory.ticketLoopCommand(),
			factory.ticketMergeCommand(),
			factory.ticketNewCommand(),
			factory.ticketReviewCommand(),
			factory.ticketReviewModeCommand(),
			factory.ticketReviewRefCommand(),
			factory.ticketShowCommand(),
			factory.ticketStartCommand(),
			factory.ticketStatusCommand(),
		},
	}
}

func ticketProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}
