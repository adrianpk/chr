package workspace

import (
	"fmt"
	"path/filepath"
)

// UserEnvironment defines a package boundary interface.
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

// UserEnvironmentOptions describes a package boundary value.
type UserEnvironmentOptions struct {
	ConfigDir string
	DataDir   string
	CacheDir  string
	StateDir  string
	HomeDir   string
	XDGConfig string
	XDGData   string
	XDGCache  string
	XDGState  string
	DevMode   bool
	DevRoot   string
}

// NewUserEnvironment constructs the corresponding package service or value.
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

	if !opts.DevMode {
		env.applyDefaults(opts)
	}

	err := env.validate()
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// ConfigDir executes the corresponding package operation.
func (e *userEnvironment) ConfigDir() string {
	return e.configDir
}

// DataDir executes the corresponding package operation.
func (e *userEnvironment) DataDir() string {
	return e.dataDir
}

// CacheDir executes the corresponding package operation.
func (e *userEnvironment) CacheDir() string {
	return e.cacheDir
}

// StateDir executes the corresponding package operation.
func (e *userEnvironment) StateDir() string {
	return e.stateDir
}

// DevMode executes the corresponding package operation.
func (e *userEnvironment) DevMode() bool {
	return e.devMode
}

func (e *userEnvironment) applyDefaults(opts UserEnvironmentOptions) {
	if e.configDir == "" {
		e.configDir = userConfigDir(opts)
	}

	if e.dataDir == "" {
		e.dataDir = userDataDir(opts)
	}

	if e.cacheDir == "" {
		e.cacheDir = userCacheDir(opts)
	}

	if e.stateDir == "" {
		e.stateDir = userStateDir(opts)
	}
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

func userConfigDir(opts UserEnvironmentOptions) string {
	if opts.XDGConfig != "" {
		return filepath.Join(opts.XDGConfig, "chc")
	}

	if opts.HomeDir == "" {
		return ""
	}

	return filepath.Join(opts.HomeDir, ".config", "chc")
}

func userDataDir(opts UserEnvironmentOptions) string {
	if opts.XDGData != "" {
		return filepath.Join(opts.XDGData, "chc")
	}

	if opts.HomeDir == "" {
		return ""
	}

	return filepath.Join(opts.HomeDir, ".local", "share", "chc")
}

func userCacheDir(opts UserEnvironmentOptions) string {
	if opts.XDGCache != "" {
		return filepath.Join(opts.XDGCache, "chc")
	}

	if opts.HomeDir == "" {
		return ""
	}

	return filepath.Join(opts.HomeDir, ".cache", "chc")
}

func userStateDir(opts UserEnvironmentOptions) string {
	if opts.XDGState != "" {
		return filepath.Join(opts.XDGState, "chc")
	}

	if opts.HomeDir == "" {
		return ""
	}

	return filepath.Join(opts.HomeDir, ".local", "state", "chc")
}
