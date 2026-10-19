package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/x/term"
	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func newWebLoginCommand(stdin io.Reader, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name: "login", Usage: "Configure the private single-user Web UI account",
		Commands: []*cli.Command{{
			Name: "setup", Usage: "Save a password verifier using masked terminal entry",
			Flags: append(projectRegistryFlags(),
				&cli.StringFlag{Name: "username", Usage: "single Web UI account name", Required: true},
				&cli.BoolFlag{Name: "replace", Usage: "replace the existing account; restart Web servers afterward"},
			),
			Action: func(_ context.Context, cmd *cli.Command) error {
				return setupWebLogin(cmd, stdin, stdout)
			},
		}},
	}
}

func loadWebLogin(env core.UserEnvironment, required bool) (*webui.LoginCredentials, error) {
	credentials, err := webui.LoadLoginCredentials(filepath.Join(env.ConfigDir(), "web-login.json"))
	if errors.Is(err, os.ErrNotExist) {
		if required {
			return nil, errors.New("web login is required; run chr web login setup --username <name>")
		}

		return nil, nil
	}

	return credentials, err
}

func setupWebLogin(cmd *cli.Command, stdin io.Reader, stdout io.Writer) error {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	path := filepath.Join(env.ConfigDir(), "web-login.json")
	if _, err := os.Lstat(path); !cmd.Bool("replace") && !errors.Is(err, os.ErrNotExist) {
		return errors.New("web login account already exists or cannot be inspected; use --replace to change it")
	}

	password, err := readWebLoginPassword(stdin, stdout)
	if err != nil {
		return err
	}
	defer clear(password)

	credentials, err := webui.CreateLoginCredentials(cmd.String("username"), string(password))
	if err != nil {
		return err
	}

	if err := webui.SaveLoginCredentials(path, credentials, cmd.Bool("replace")); err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, "Web login configured. Restart Web servers to load the account and invalidate existing sessions.")

	return err
}

func readWebLoginPassword(stdin io.Reader, stdout io.Writer) ([]byte, error) {
	file, ok := stdin.(*os.File)
	if !ok || !term.IsTerminal(file.Fd()) {
		return nil, errors.New("web password entry requires an interactive terminal")
	}

	password, err := promptWebPassword(file, stdout, "Password: ")
	if err != nil {
		return nil, err
	}

	confirmation, err := promptWebPassword(file, stdout, "Confirm password: ")
	defer clear(confirmation)

	if err != nil || !bytes.Equal(password, confirmation) {
		clear(password)

		return nil, errors.New("password confirmation failed")
	}

	return password, nil
}

func promptWebPassword(file *os.File, stdout io.Writer, prompt string) ([]byte, error) {
	if _, err := fmt.Fprint(stdout, prompt); err != nil {
		return nil, err
	}

	password, err := term.ReadPassword(file.Fd())
	if err != nil {
		return nil, errors.New("read masked web password failed")
	}

	if _, err := fmt.Fprintln(stdout); err != nil {
		clear(password)

		return nil, err
	}

	return password, nil
}
