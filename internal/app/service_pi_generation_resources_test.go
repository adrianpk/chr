package app

import (
	"os"
	"path/filepath"
	"testing"
)

// A recovered operation uses its private frozen references after profiles disappear.
func TestPiGenerationResourceRecovery(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"ok":true}`)
	config := operations.env.ConfigDir()
	models := filepath.Join(config, "models.json")
	definition := `{"providers":{"fixture":{"api":"openai-completions","baseUrl":"http://127.0.0.1:1/v1"}}}`
	harnessSelectionWrite(t, models, definition)

	profiles := filepath.Join(config, "harnesses.yml")
	harnessSelectionWrite(t, profiles, "profiles:\n  writing:\n    harness: pi\n    command: "+filepath.Join(driver, "driver")+"\n    model_provider: fixture\n    model: fixture-model\n    credential_ref: auth.json\n    model_config_ref: models.json\n")

	if _, err := operations.bindGeneration(t.Context(), "resource-recovery", "conversation"); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, profiles, "profiles: {}\n")
	harnessSelectionWrite(t, filepath.Join(config, "auth.json"), `{"fixture":{"type":"api_key","key":"ROTATED_FIXTURE_KEY"}}`)

	fresh := newService(t.Context(), operations.env, operations.projectContext, testAssetsFS())

	bound, err := fresh.bindGeneration(t.Context(), "resource-recovery", "conversation")

	ctx := bound.context(t.Context())
	if err != nil {
		t.Fatal("recovered Pi profile was re-resolved", err)
	}

	harnessSelectionWrite(t, models, `{"providers":{"fixture":{"api":"openai-completions","baseUrl":"http://127.0.0.1:2/v1"}}}`)

	if _, err := bound.conversationModel.InvokeStructuredModel(ctx, piCapacityRequest(bound.projectContext.ProjectDir, "tampered")); err == nil {
		t.Fatal("changed frozen model definition accepted")
	}

	if _, err := os.Stat(filepath.Join(driver, "version-called")); !os.IsNotExist(err) {
		t.Fatal("changed model definition launched Pi", err)
	}

	harnessSelectionWrite(t, models, definition)

	if _, err := bound.conversationModel.InvokeStructuredModel(ctx, piCapacityRequest(bound.projectContext.ProjectDir, "restored")); err != nil {
		t.Fatal("restored frozen resource failed", err)
	}
}
