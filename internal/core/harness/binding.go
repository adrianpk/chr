package harness

import adapterharness "forge.adrianpk.com/chr/internal/harness"

// ExecutableRef binds host-owned executable bytes and the complete argv.
// The resolved path stays in the private invocation command, not this projection.
type ExecutableRef struct {
	Hash        string `json:"hash"`
	CommandHash string `json:"command_hash"`
}

// Binding is the non-secret semantic identity frozen before admission.
// Capabilities describe adapter support, not a whole-call accounting certificate.
type Binding struct {
	SchemaVersion     int                         `json:"schema_version"`
	Role              Role                        `json:"role"`
	Harness           string                      `json:"harness"`
	Source            string                      `json:"source"`
	SelectionRevision string                      `json:"selection_revision"`
	Profile           string                      `json:"profile,omitempty"`
	ProfileRevision   string                      `json:"profile_revision"`
	ModelProvider     string                      `json:"model_provider"`
	Model             string                      `json:"model"`
	Thinking          string                      `json:"thinking,omitempty"`
	ExecutableRef     ExecutableRef               `json:"executable_ref"`
	RuntimeRef        string                      `json:"runtime_ref"`
	ProtocolRevision  string                      `json:"protocol_revision"`
	Capabilities      adapterharness.Capabilities `json:"capabilities"`
	CapabilitiesRef   string                      `json:"capabilities_ref"`
	BindingHash       string                      `json:"binding_hash"`
}

// Role separates worker authority from structured generation.
type Role string

// Supported role identities are independent workspace selections.
const (
	Worker     Role = "worker"
	Generation Role = "generation"
)
