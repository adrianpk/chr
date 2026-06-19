package app

import "forge.adrianpk.com/chc/internal/core"

func workSelectionBranchPolicy(projectDir string) (core.WorkBranchPolicy, error) {
	config, err := workSelectionWorkConfig(projectDir)
	if err != nil {
		return core.WorkBranchPolicy{}, err
	}

	return config.BranchPolicy, nil
}

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
