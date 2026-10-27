package core

import (
	"encoding/json"
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/harness"
)

// HarnessExecutableRef binds host-owned executable bytes and the complete argv.
// The resolved path stays in the private invocation command, not this projection.
type HarnessExecutableRef struct {
	Hash        string `json:"hash"`
	CommandHash string `json:"command_hash"`
}

// HarnessBinding is the non-secret semantic identity frozen before admission.
// Capabilities describe adapter support, not a whole-call accounting certificate.
type HarnessBinding struct {
	SchemaVersion     int                  `json:"schema_version"`
	Role              HarnessRole          `json:"role"`
	Harness           string               `json:"harness"`
	Source            string               `json:"source"`
	SelectionRevision string               `json:"selection_revision"`
	Profile           string               `json:"profile,omitempty"`
	ProfileRevision   string               `json:"profile_revision"`
	ModelProvider     string               `json:"model_provider"`
	Model             string               `json:"model"`
	Thinking          string               `json:"thinking,omitempty"`
	ExecutableRef     HarnessExecutableRef `json:"executable_ref"`
	RuntimeRef        string               `json:"runtime_ref"`
	ProtocolRevision  string               `json:"protocol_revision"`
	Capabilities      harness.Capabilities `json:"capabilities"`
	CapabilitiesRef   string               `json:"capabilities_ref"`
	BindingHash       string               `json:"binding_hash"`
}

// SealHarnessBinding hashes a prepared identity without incorporating credentials.
func SealHarnessBinding(binding HarnessBinding) (HarnessBinding, error) {
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
func ValidateHarnessBinding(binding HarnessBinding) error {
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

func validateHarnessBindingProtocol(binding HarnessBinding) error {
	if binding.SchemaVersion != 1 || binding.Role != HarnessWorker || !workflowConfigChoice(binding.Harness, []string{"codex", "pi"}) {
		return errors.New("unsupported worker harness binding")
	}

	if !harnessSelectionDigest(binding.ExecutableRef.Hash) || !harnessSelectionDigest(binding.ExecutableRef.CommandHash) || !binding.Capabilities.Worker {
		return errors.New("worker binding requires executable identity and capability evidence")
	}

	return nil
}

func validateHarnessBindingSelection(b HarnessBinding) error {
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

func validateHarnessBindingSource(b HarnessBinding) error {
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

func cloneHarnessBinding(binding *HarnessBinding) *HarnessBinding {
	if binding == nil {
		return nil
	}

	copy := *binding

	return &copy
}

func validateWorkflowHarnessPair(binding *HarnessBinding, command []string) error {
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

func validateWorkflowRequestedHarness(binding *HarnessBinding, command []string) error {
	if binding != nil && len(command) == 0 {
		return ValidateHarnessBinding(*binding)
	}

	return validateWorkflowHarnessPair(binding, command)
}
