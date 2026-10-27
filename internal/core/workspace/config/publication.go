package config

const (
	// WorkflowPublicationDisabled preserves the existing publication policy.
	WorkflowPublicationDisabled = "disabled"
	// WorkflowPublicationGitAtomicDualRef explicitly permits the owned source rewrite.
	WorkflowPublicationGitAtomicDualRef = "git_atomic_dual_ref"
)

// WorkIntegrationConfig selects a publication mechanism independently of review.
type WorkIntegrationConfig struct {
	PublicationProfile   string
	MaxPreparedPerTarget int
}
