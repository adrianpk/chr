package core

import "forge.adrianpk.com/chr/internal/harness"

// HarnessReadiness describes local preparation separately from remote authentication.
type HarnessReadiness struct {
	Executable           string `json:"executable"`
	Credentials          string `json:"credentials"`
	Runtime              string `json:"runtime"`
	Supported            bool   `json:"supported"`
	RemoteAuthentication string `json:"remote_authentication"`
}

// Ready reports whether local dispatch is supported; it makes no remote call.
func (r HarnessReadiness) Ready() bool {
	return r.Supported && workflowConfigChoice(r.Executable, []string{"available", "not_required"}) && workflowConfigChoice(r.Runtime, []string{"native", "not_required", "available"}) &&
		(r.Credentials == "present" || r.Credentials == "cli_managed" || r.Credentials == "api_managed")
}

// HarnessSelection is a readable projection without executable or credential paths.
type HarnessSelection struct {
	Role              HarnessRole          `json:"role"`
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
	Readiness         HarnessReadiness     `json:"readiness"`
}

// Revision identifies the complete persisted configuration image.
func (c WorkspaceConfig) Revision() string {
	if c.workflowRevision == "" {
		return workflowHash(nil)
	}

	return c.workflowRevision
}
