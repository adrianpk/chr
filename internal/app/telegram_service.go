package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) installTelegramUserService(cmd *cli.Command, env core.UserEnvironment) error {
	executable, err := os.Executable()
	if err != nil {
		return errors.New("telegram service executable is unavailable")
	}

	data, err := renderTelegramUserService(env, executable, cmd.String("dev-root"))
	if err != nil {
		return err
	}

	path := filepath.Join(filepath.Dir(env.ConfigDir()), "systemd", "user", "chr-telegram.service")
	if err := saveTelegramUserService(path, data, cmd.Bool("replace")); err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"service: installed\nservice_path: %s\nservice_activation: unchanged\n", path)

	return err
}

func renderTelegramUserService(env core.UserEnvironment, executable, devRoot string) ([]byte, error) {
	values := []string{executable, devRoot, env.ConfigDir(), env.DataDir(), env.CacheDir(), env.StateDir()}
	for _, value := range values {
		if strings.ContainsFunc(value, unicode.IsControl) {
			return nil, errors.New("telegram service paths contain unsupported characters")
		}
	}

	if !filepath.IsAbs(executable) {
		return nil, errors.New("telegram service requires an absolute executable")
	}

	command := telegramUnitQuote(executable, true) + " telegram serve"

	if devRoot != "" {
		absolute, err := filepath.Abs(devRoot)
		if err != nil {
			return nil, errors.New("telegram service development root is unavailable")
		}

		command += " --dev-root " + telegramUnitQuote(absolute, true)
	}

	var unit strings.Builder
	unit.WriteString("[Unit]\nDescription=chr Telegram receiver\nStartLimitIntervalSec=60\nStartLimitBurst=5\n\n[Service]\nType=simple\n")
	unit.WriteString("ExecStart=" + command + "\n")

	for _, variable := range []struct{ name, directory string }{
		{"XDG_CONFIG_HOME", env.ConfigDir()}, {"XDG_DATA_HOME", env.DataDir()}, {"XDG_CACHE_HOME", env.CacheDir()}, {"XDG_STATE_HOME", env.StateDir()},
	} {
		unit.WriteString("Environment=" + telegramUnitQuote(variable.name+"="+filepath.Dir(variable.directory), false) + "\n")
	}

	unit.WriteString("EnvironmentFile=" + telegramUnitQuote("-"+filepath.Join(env.ConfigDir(), "telegram.env"), false) + "\n")
	unit.WriteString("Restart=on-failure\nRestartSec=5s\nRestartPreventExitStatus=78\nTimeoutStopSec=35s\nKillMode=process\nUMask=0077\n\n[Install]\nWantedBy=default.target\n")

	return []byte(unit.String()), nil
}

func telegramUnitQuote(value string, command bool) string {
	value = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "%", "%%").Replace(value)
	if command {
		value = strings.ReplaceAll(value, "$", "$$")
	}

	return `"` + value + `"`
}

func saveTelegramUserService(path string, data []byte, replace bool) (err error) {
	unchanged, err := guardTelegramUserService(path, data, replace)
	if err != nil || unchanged {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".chr-telegram-*")
	if err != nil {
		return err
	}

	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, removeErr)
		}
	}()

	if _, err := file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), path)
}

func guardTelegramUserService(path string, data []byte, replace bool) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, errors.New("telegram service is unavailable")
	}

	if !info.Mode().IsRegular() || info.Size() > 64*1024 {
		return false, errors.New("telegram service must be a bounded regular file")
	}

	prior, err := os.ReadFile(path)
	if err != nil {
		return false, errors.New("telegram service is unreadable")
	}

	if string(prior) == string(data) {
		return true, nil
	}

	if !replace {
		return false, errors.New("telegram service exists; use --replace to change it")
	}

	return false, nil
}
