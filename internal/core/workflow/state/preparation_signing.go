package state

// CommitSigningProfile identifies the certified SSH signer policy.
type CommitSigningProfile struct {
	Hash      string `json:"hash"`
	PublicKey string `json:"public_key"`
	Principal string `json:"principal"`
}

// IntegrationSignature binds cryptographic verification to one object.
type IntegrationSignature struct {
	Commit        string `json:"commit"`
	SignatureHash string `json:"signature_hash"`
}
