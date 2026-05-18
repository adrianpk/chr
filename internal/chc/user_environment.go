package chc

import (
	"fmt"
	"path/filepath"
)

type UserEnvironment interface {
	ConfigDir() string
	DataDir() string
	CacheDir() string
	StateDir() string
	DevMode() bool
}

type userEnvironment struct {
	configDir string
	dataDir   string
	cacheDir  string
	stateDir  string
	devMode   bool
}

type UserEnvironmentOptions struct {
	ConfigDir string
	DataDir   string
	CacheDir  string
	StateDir  string
	DevMode   bool
	DevRoot   string
}

func NewUserEnvironment(opts UserEnvironmentOptions) (UserEnvironment, error) {
	env := userEnvironment{
		configDir: opts.ConfigDir,
		dataDir:   opts.DataDir,
		cacheDir:  opts.CacheDir,
		stateDir:  opts.StateDir,
		devMode:   opts.DevMode,
	}

	if opts.DevMode {
		if opts.DevRoot == "" {
			return nil, fmt.Errorf("development root is required in dev mode")
		}

		env.configDir = filepath.Join(opts.DevRoot, "config", "chc")
		env.dataDir = filepath.Join(opts.DevRoot, "data", "chc")
		env.cacheDir = filepath.Join(opts.DevRoot, "cache", "chc")
		env.stateDir = filepath.Join(opts.DevRoot, "state", "chc")
	}

	if err := env.validate(); err != nil {
		return nil, err
	}

	return &env, nil
}

func (e *userEnvironment) ConfigDir() string {
	return e.configDir
}

func (e *userEnvironment) DataDir() string {
	return e.dataDir
}

func (e *userEnvironment) CacheDir() string {
	return e.cacheDir
}

func (e *userEnvironment) StateDir() string {
	return e.stateDir
}

func (e *userEnvironment) DevMode() bool {
	return e.devMode
}

func (e *userEnvironment) validate() error {
	if e.configDir == "" {
		return fmt.Errorf("config directory is required")
	}

	if e.dataDir == "" {
		return fmt.Errorf("data directory is required")
	}

	if e.cacheDir == "" {
		return fmt.Errorf("cache directory is required")
	}

	if e.stateDir == "" {
		return fmt.Errorf("state directory is required")
	}

	return nil
}
