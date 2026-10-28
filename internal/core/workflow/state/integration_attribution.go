package state

// IntegrationIdentity preserves an object's identity and Git date.
type IntegrationIdentity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// IntegrationAttribution binds metadata to the complete commit object.
// MessageHash covers the original bytes, including trailing newlines.
type IntegrationAttribution struct {
	Commit      string              `json:"commit"`
	Tree        string              `json:"tree"`
	Parents     []string            `json:"parents"`
	Author      IntegrationIdentity `json:"author"`
	Committer   IntegrationIdentity `json:"committer"`
	MessageHash string              `json:"message_hash"`
	Encoding    string              `json:"encoding,omitempty"`
}
