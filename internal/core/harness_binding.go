package core

import (
	"errors"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

func cloneHarnessBinding(binding *harnesscore.Binding) *harnesscore.Binding {
	if binding == nil {
		return nil
	}

	copy := *binding

	return &copy
}

func validateWorkflowHarnessPair(binding *harnesscore.Binding, command []string) error {
	if binding == nil {
		if len(command) != 0 {
			return errors.New("unbound workflow cannot retain a harness command")
		}

		return nil
	}

	if err := harnesscore.ValidateHarnessBinding(*binding); err != nil {
		return err
	}

	return checkHarnessCommandImage(*binding, command)
}

func validateWorkflowRequestedHarness(binding *harnesscore.Binding, command []string) error {
	if binding != nil && len(command) == 0 {
		return harnesscore.ValidateHarnessBinding(*binding)
	}

	return validateWorkflowHarnessPair(binding, command)
}
