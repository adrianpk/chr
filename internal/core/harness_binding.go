package core

import (
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
)

func validateWorkflowRequestedHarness(binding *harnesscore.Binding, command []string) error {
	if binding != nil && len(command) == 0 {
		return harnesscore.ValidateBinding(*binding)
	}

	return harnesshost.ValidateBindingCommand(binding, command)
}
