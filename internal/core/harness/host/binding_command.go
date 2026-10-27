package host

import (
	"errors"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

// ValidateBindingCommand preserves the admitted binding value and command contract.
func ValidateBindingCommand(binding *harnesscore.Binding, command []string) error {
	if binding == nil {
		if len(command) != 0 {
			return errors.New("unbound workflow cannot retain a harness command")
		}

		return nil
	}

	if err := harnesscore.ValidateBinding(*binding); err != nil {
		return err
	}

	return CheckCommandImage(*binding, command)
}
