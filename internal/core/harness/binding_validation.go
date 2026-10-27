package harness

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// SealHarnessBinding hashes a prepared identity without incorporating credentials.
func SealHarnessBinding(binding Binding) (Binding, error) {
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
func ValidateHarnessBinding(binding Binding) error {
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

func validateHarnessBindingProtocol(binding Binding) error {
	if binding.SchemaVersion != 1 || binding.Role != Worker || !slices.Contains([]string{"codex", "pi"}, binding.Harness) {
		return errors.New("unsupported worker harness binding")
	}

	if !SelectionDigest(binding.ExecutableRef.Hash) || !SelectionDigest(binding.ExecutableRef.CommandHash) || !binding.Capabilities.Worker {
		return errors.New("worker binding requires executable identity and capability evidence")
	}

	return nil
}

func validateHarnessBindingSelection(b Binding) error {
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

func validateHarnessBindingSource(b Binding) error {
	if !slices.Contains([]string{"project", "legacy"}, b.Source) || !SelectionDigest(b.SelectionRevision) || !SelectionDigest(b.ProfileRevision) {
		return errors.New("harness binding requires selection and profile revisions")
	}

	if b.Source == "project" && !ValidHarnessProfileName(b.Profile) {
		return errors.New("explicit harness binding requires a named profile")
	}

	if b.Source == "legacy" && b.Profile != "" {
		return errors.New("legacy harness binding cannot claim a selected profile")
	}

	return nil
}

// ValidHarnessProfileName checks the existing bounded profile identifier.
func ValidHarnessProfileName(value string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`).MatchString(value)
}

// SelectionDigest recognizes an exact SHA-256 selection digest.
func SelectionDigest(value string) bool {
	decoded, err := hex.DecodeString(value)

	return err == nil && len(decoded) == 32
}
