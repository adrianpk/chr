package core

import (
	"encoding/json"
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

// HarnessGenerationBinding retains a non-secret selection for one operation.
// Readiness records admission-time local evidence, never remote authentication.
type HarnessGenerationBinding struct {
	Selection        HarnessSelection          `json:"selection"`
	ExecutableRef    harnesscore.ExecutableRef `json:"executable_ref"`
	ProtocolRevision string                    `json:"protocol_revision"`
	BindingHash      string                    `json:"binding_hash"`
}

// SealHarnessGenerationBinding seals the complete admitted selection.
func SealHarnessGenerationBinding(binding HarnessGenerationBinding) (HarnessGenerationBinding, error) {
	binding.BindingHash = ""

	data, err := json.Marshal(binding)
	if err != nil {
		return binding, err
	}

	binding.BindingHash = contenthash.Sum(data)

	return binding, ValidateHarnessGenerationBinding(binding)
}

// ValidateHarnessGenerationBinding checks retained identity without reading live settings.
func ValidateHarnessGenerationBinding(binding HarnessGenerationBinding) error {
	if err := validateGenerationBindingSelection(binding.Selection); err != nil {
		return err
	}

	if err := validateGenerationBindingText(binding); err != nil {
		return err
	}

	if err := validateGenerationBindingTransport(binding); err != nil {
		return err
	}

	sealed := binding
	sealed.BindingHash = ""

	data, err := json.Marshal(sealed)
	if err != nil || contenthash.Sum(data) != binding.BindingHash {
		return errors.New("generation binding evidence changed")
	}

	return nil
}

func validateGenerationBindingSelection(s HarnessSelection) error {
	if s.Role != harnesscore.Generation || !s.Readiness.Ready() || !harnesscore.SelectionDigest(s.SelectionRevision) || !harnesscore.SelectionDigest(s.ProfileRevision) {
		return errors.New("generation binding requires a ready versioned selection")
	}

	return validateGenerationBindingSource(s)
}

func validateGenerationBindingText(binding HarnessGenerationBinding) error {
	s := binding.Selection
	for _, value := range []string{s.Action, s.ModelProvider, s.Model, binding.ProtocolRevision} {
		if value == "" || len(value) > 4096 || strings.ContainsAny(value, "\x00\r\n") {
			return errors.New("generation binding requires bounded action, provider, model and protocol identities")
		}
	}

	if len(s.Thinking) > 128 {
		return errors.New("generation thinking selection exceeds bounds")
	}

	return nil
}

func validateGenerationBindingSource(s HarnessSelection) error {
	if s.Source == "project" && harnesscore.ValidHarnessProfileName(s.Profile) {
		return nil
	}

	if s.Source == "legacy" && s.Profile == "" {
		return nil
	}

	return errors.New("generation binding requires an explicit profile or legacy provenance")
}

func validateGenerationBindingTransport(binding HarnessGenerationBinding) error {
	s := binding.Selection
	if s.Backend == "openai" {
		return validateGenerationBindingAPI(binding)
	}

	if !workflowConfigChoice(s.Backend, []string{"codex", "pi"}) || s.Harness != s.Backend || !s.Capabilities.Generation {
		return errors.New("generation binding requires executable and capability evidence")
	}

	if !harnesscore.SelectionDigest(binding.ExecutableRef.Hash) || !harnesscore.SelectionDigest(binding.ExecutableRef.CommandHash) {
		return errors.New("generation binding requires executable fingerprints")
	}

	return nil
}

func validateGenerationBindingAPI(binding HarnessGenerationBinding) error {
	s := binding.Selection
	if s.Harness != "" || s.Source != "legacy" || s.ModelProvider != "openai" || binding.ExecutableRef != (harnesscore.ExecutableRef{}) {
		return errors.New("direct API generation cannot claim a harness executable")
	}

	return nil
}
