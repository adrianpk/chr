package app

import "forge.adrianpk.com/chr/internal/core"

func workSelectionWorkConfig(projectDir string) (core.WorkConfig, error) {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return core.WorkConfig{}, err
	}

	err = core.ValidateWorkConfig(config.Work)
	if err != nil {
		return core.WorkConfig{}, err
	}

	return config.Work, nil
}
