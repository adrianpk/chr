package harness

import (
	"slices"

	"forge.adrianpk.com/chr/internal/harness"
)

// Readiness describes local preparation separately from remote authentication.
type Readiness struct {
	Executable           string `json:"executable"`
	Credentials          string `json:"credentials"`
	Runtime              string `json:"runtime"`
	Supported            bool   `json:"supported"`
	RemoteAuthentication string `json:"remote_authentication"`
	Problem              string `json:"problem,omitempty"`
}

// Ready reports whether local dispatch is supported; it makes no remote call.
func (r Readiness) Ready() bool {
	return r.Supported && slices.Contains([]string{"available", "not_required"}, r.Executable) && slices.Contains([]string{"native", "not_required", "available"}, r.Runtime) &&
		(r.Credentials == "present" || r.Credentials == "cli_managed" || r.Credentials == "api_managed")
}

// Selection is a readable projection without executable or credential paths.
type Selection struct {
	Role              Role                 `json:"role"`
	Action            string               `json:"action,omitempty"`
	Profile           string               `json:"profile,omitempty"`
	Harness           string               `json:"harness,omitempty"`
	Backend           string               `json:"backend"`
	ModelProvider     string               `json:"model_provider,omitempty"`
	Model             string               `json:"model"`
	Thinking          string               `json:"thinking,omitempty"`
	Source            string               `json:"source"`
	SelectionRevision string               `json:"selection_revision"`
	ProfileRevision   string               `json:"profile_revision,omitempty"`
	Capabilities      harness.Capabilities `json:"capabilities"`
	Readiness         Readiness            `json:"readiness"`
}
