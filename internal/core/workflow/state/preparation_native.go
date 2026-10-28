package state

// NativePreparationInput freezes the queue owner and full capability
// before the host gate executes. Agent output cannot replace this input.
type NativePreparationInput struct {
	Entry           IntegrationBinding `json:"entry"`
	ClaimGeneration uint64             `json:"claim_generation"`
	Capability      NativeCapability   `json:"capability"`
}
