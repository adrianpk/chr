package harness

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// GenerationBinding retains a non-secret selection for one operation.
// Readiness records admission-time local evidence, never remote authentication.
type GenerationBinding struct {
	Selection        Selection     `json:"selection"`
	ExecutableRef    ExecutableRef `json:"executable_ref"`
	ProtocolRevision string        `json:"protocol_revision"`
	BindingHash      string        `json:"binding_hash"`
}

// SealGenerationBinding seals the complete admitted selection.
func SealGenerationBinding(binding GenerationBinding) (GenerationBinding, error) {
	binding.BindingHash = ""

	data, err := json.Marshal(binding)
	if err != nil {
		return binding, err
	}

	binding.BindingHash = contenthash.Sum(data)

	return binding, ValidateGenerationBinding(binding)
}

// ValidateGenerationBinding checks retained identity without reading live settings.
func ValidateGenerationBinding(binding GenerationBinding) error {
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

func validateGenerationBindingSelection(s Selection) error {
	if s.Role != Generation || !s.Readiness.Ready() || !SelectionDigest(s.SelectionRevision) || !SelectionDigest(s.ProfileRevision) {
		return errors.New("generation binding requires a ready versioned selection")
	}

	return validateGenerationBindingSource(s)
}

func validateGenerationBindingText(binding GenerationBinding) error {
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

func validateGenerationBindingSource(s Selection) error {
	if s.Source == "project" && ValidProfileName(s.Profile) {
		return nil
	}

	if s.Source == "legacy" && s.Profile == "" {
		return nil
	}

	return errors.New("generation binding requires an explicit profile or legacy provenance")
}

func validateGenerationBindingTransport(binding GenerationBinding) error {
	s := binding.Selection
	if s.Backend == "openai" {
		return validateGenerationBindingAPI(binding)
	}

	if !slices.Contains([]string{"codex", "pi"}, s.Backend) || s.Harness != s.Backend || !s.Capabilities.Generation {
		return errors.New("generation binding requires executable and capability evidence")
	}

	if !SelectionDigest(binding.ExecutableRef.Hash) || !SelectionDigest(binding.ExecutableRef.CommandHash) {
		return errors.New("generation binding requires executable fingerprints")
	}

	return nil
}

func validateGenerationBindingAPI(binding GenerationBinding) error {
	s := binding.Selection
	if s.Harness != "" || s.Source != "legacy" || s.ModelProvider != "openai" || binding.ExecutableRef != (ExecutableRef{}) {
		return errors.New("direct API generation cannot claim a harness executable")
	}

	return nil
}
