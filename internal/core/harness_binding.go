package core

import (
	"encoding/json"
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

// SealHarnessBinding hashes a prepared identity without incorporating credentials.
func SealHarnessBinding(binding harnesscore.Binding) (harnesscore.Binding, error) {
	binding.BindingHash = ""

	capabilities, err := json.Marshal(binding.Capabilities)
	if err != nil {
		return binding, err
	}

	binding.CapabilitiesRef = contenthash.Sum(capabilities)

	data, err := json.Marshal(binding)
	if err != nil {
		return binding, err
	}

	binding.BindingHash = contenthash.Sum(data)

	return binding, ValidateHarnessBinding(binding)
}

// ValidateHarnessBinding validates retained semantics without installing or running IO.
func ValidateHarnessBinding(binding harnesscore.Binding) error {
	if err := validateHarnessBindingProtocol(binding); err != nil {
		return err
	}

	if err := validateHarnessBindingSelection(binding); err != nil {
		return err
	}

	sealed := binding
	sealed.BindingHash = ""

	capabilities, err := json.Marshal(binding.Capabilities)
	if err != nil {
		return err
	}

	data, err := json.Marshal(sealed)
	if err != nil {
		return err
	}

	if contenthash.Sum(capabilities) != binding.CapabilitiesRef || contenthash.Sum(data) != binding.BindingHash {
		return errors.New("harness binding evidence changed")
	}

	return nil
}

func validateHarnessBindingProtocol(binding harnesscore.Binding) error {
	if binding.SchemaVersion != 1 || binding.Role != harnesscore.Worker || !workflowConfigChoice(binding.Harness, []string{"codex", "pi"}) {
		return errors.New("unsupported worker harness binding")
	}

	if !harnessSelectionDigest(binding.ExecutableRef.Hash) || !harnessSelectionDigest(binding.ExecutableRef.CommandHash) || !binding.Capabilities.Worker {
		return errors.New("worker binding requires executable identity and capability evidence")
	}

	return nil
}

func validateHarnessBindingSelection(b harnesscore.Binding) error {
	if err := validateHarnessBindingSource(b); err != nil {
		return err
	}

	for _, value := range []string{b.ModelProvider, b.Model, b.RuntimeRef, b.ProtocolRevision} {
		if value == "" || len(value) > 4096 || strings.ContainsAny(value, "\x00\r\n") {
			return errors.New("harness binding requires bounded provider, model, runtime and protocol identities")
		}
	}

	if len(b.Thinking) > 128 {
		return errors.New("harness thinking selection exceeds bounds")
	}

	return nil
}

func validateHarnessBindingSource(b harnesscore.Binding) error {
	if !workflowConfigChoice(b.Source, []string{"project", "legacy"}) || !harnessSelectionDigest(b.SelectionRevision) || !harnessSelectionDigest(b.ProfileRevision) {
		return errors.New("harness binding requires selection and profile revisions")
	}

	if b.Source == "project" && !validHarnessProfileName(b.Profile) {
		return errors.New("explicit harness binding requires a named profile")
	}

	if b.Source == "legacy" && b.Profile != "" {
		return errors.New("legacy harness binding cannot claim a selected profile")
	}

	return nil
}

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

	if err := ValidateHarnessBinding(*binding); err != nil {
		return err
	}

	return checkHarnessCommandImage(*binding, command)
}

func validateWorkflowRequestedHarness(binding *harnesscore.Binding, command []string) error {
	if binding != nil && len(command) == 0 {
		return ValidateHarnessBinding(*binding)
	}

	return validateWorkflowHarnessPair(binding, command)
}
