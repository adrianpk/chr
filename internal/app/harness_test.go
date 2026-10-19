package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Exercise default service assembly through a real owned Codex subprocess.
func TestHarnessCodexConversation(t *testing.T) {
	dir := t.TempDir()
	program := filepath.Join(dir, "codex-fixture")

	script := "#!/bin/sh\ncat >/dev/null\nprintf '%s' '{\"answer\":\"ready\"}'\n"
	if err := os.WriteFile(program, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CHR_CODEX_COMMAND", program)
	t.Setenv("CHR_CONVERSATION_GENERATION_BACKEND", "codex")

	operations := newService(t.Context(), nil, commandProjectContext{ProjectDir: dir}, nil)

	output, err := operations.conversationModel.InvokeStructuredModel(t.Context(), core.StructuredModelRequest{
		ID: "conversation-fixture", Purpose: "conversation", ProjectDir: dir,
		Instructions: "Return an answer.", Input: json.RawMessage(`{"question":"ready"}`),
		Schema: map[string]any{"type": "object"},
	})
	if err != nil || string(output) != `{"answer":"ready"}` {
		t.Fatalf("conversation = %q, %v", output, err)
	}
}
