package state

// NativeUnsupported identifies a condition and the required decision.
type NativeUnsupported struct {
	Condition string `json:"condition"`
	Decision  string `json:"decision"`
}

// NativeCapability binds current policy and transport to the original PR.
type NativeCapability struct {
	PublicationCapability
	SourceRef                     string              `json:"source_ref"`
	TargetRef                     string              `json:"target_ref"`
	OwnershipRevision             string              `json:"ownership_revision"`
	PolicyFingerprint             string              `json:"policy_fingerprint"`
	ServerCapabilityFingerprint   string              `json:"server_capability_fingerprint"`
	CredentialIdentityFingerprint string              `json:"credential_identity_fingerprint"`
	ObservationContract           string              `json:"observation_contract"`
	AdapterVersion                string              `json:"adapter_version"`
	OriginalSourceCommit          string              `json:"original_source_commit"`
	ExpectedBaseCommit            string              `json:"expected_base_commit"`
	PreparedResultCommit          string              `json:"prepared_result_commit"`
	Unsupported                   []NativeUnsupported `json:"unsupported,omitempty"`
}
