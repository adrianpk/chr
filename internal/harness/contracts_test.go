package harness

import "testing"

// Moved request validation must reject absent identity before adapter execution.
func TestAgentRunRequest(t *testing.T) {
	if err := ValidateAgentRunRequest(AgentRunRequest{}); err == nil {
		t.Fatal("empty request accepted")
	}

	request := AgentRunRequest{ID: "call", Purpose: "generation", ProjectDir: ".", Prompt: "Return JSON.", OutputName: "../output.json"}
	if err := ValidateAgentRunRequest(request); err == nil {
		t.Fatal("output outside the owned run directory accepted")
	}
}
