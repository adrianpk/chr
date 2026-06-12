package app

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/workspace"
)

func projectRegistryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "dev-root",
			Usage: "development user environment root",
		},
	}
}

func newUserEnvironmentFromCommand(cmd *cli.Command) (workspace.UserEnvironment, error) {
	return newUserEnvironment(cmd.String("dev-root"))
}

func newUserEnvironment(devRoot string) (workspace.UserEnvironment, error) {
	if devRoot != "" {
		return workspace.NewUserEnvironment(workspace.UserEnvironmentOptions{
			DevMode: true,
			DevRoot: devRoot,
		})
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	return workspace.NewUserEnvironment(workspace.UserEnvironmentOptions{
		HomeDir:   homeDir,
		XDGConfig: os.Getenv("XDG_CONFIG_HOME"),
		XDGData:   os.Getenv("XDG_DATA_HOME"),
		XDGCache:  os.Getenv("XDG_CACHE_HOME"),
		XDGState:  os.Getenv("XDG_STATE_HOME"),
	})
}

func requireWorkspace(projectDir string) error {
	info, err := os.Stat(filepath.Join(projectDir, workspace.DirectoryName))
	if err != nil {
		return fmt.Errorf("inspect project workspace: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("project workspace is not a directory: %s", filepath.Join(projectDir, workspace.DirectoryName))
	}

	return nil
}

func projectID(projectDir string) string {
	sum := sha256.Sum256([]byte(projectDir))

	return "project-" + hex.EncodeToString(sum[:8])
}
