package app

import (
	"github.com/urfave/cli/v3"
)

func (factory commandFactory) maintenanceCommand() *cli.Command {
	return &cli.Command{
		Name:  "maintenance",
		Usage: "Inspect and recover local chr state",
		Commands: []*cli.Command{
			factory.maintenanceBundleCommand(),
			factory.maintenanceDBCommand(),
			factory.maintenanceExecutionCommand(),
			factory.maintenanceRegistryCommand(),
		},
	}
}

func maintenanceProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}
