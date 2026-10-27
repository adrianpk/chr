package app

import workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"

func workSelectionWorkConfig(projectDir string) (workspaceconfig.WorkConfig, error) {
	config, err := workspaceconfig.Load(projectDir)
	if err != nil {
		return workspaceconfig.WorkConfig{}, err
	}

	err = workspaceconfig.ValidateWorkConfig(config.Work)
	if err != nil {
		return workspaceconfig.WorkConfig{}, err
	}

	return config.Work, nil
}
