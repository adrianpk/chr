package app

import "forge.adrianpk.com/chc/internal/core"

func workSelectionBranchPolicy(projectDir string) (core.WorkBranchPolicy, error) {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return core.WorkBranchPolicy{}, err
	}

	err = core.ValidateWorkConfig(config.Work)
	if err != nil {
		return core.WorkBranchPolicy{}, err
	}

	return config.Work.BranchPolicy, nil
}
