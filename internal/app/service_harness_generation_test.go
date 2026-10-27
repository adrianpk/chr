package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Actual authoring admission, provider failure and a rebuilt service retain the
// first profile. A later independent request observes the live replacement.
func TestHarnessGenerationBindingAuthoring(t *testing.T) {
	generationBindingAuthoringCase(t, harnessGenerationFixture)
}

func generationBindingAuthoringCase(t *testing.T, fixture func(*testing.T, string) (service, string)) {
	operations, driverDir := fixture(t, `{"body":"Preserve the supplied scope and retained retry ceiling.","question":"","rationale":""}`)
	project := operations.projectContext.ProjectDir
	request := telegramAuthoringServiceRequest()
	request.ProjectID = operations.projectContext.ProjectID

	harnessSelectionWrite(t, filepath.Join(driverDir, "fail-next"), "owned failure")

	if _, err := operations.ContinueAuthoring(t.Context(), request); !errors.Is(err, authoringcore.ErrProviderFailure) {
		t.Fatal("expected retained provider failure", err)
	}

	switchGenerationFixture(t, operations, "other")

	profiles := filepath.Join(operations.env.ConfigDir(), "harnesses.yml")

	body, err := os.ReadFile(profiles)
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, profiles, strings.Replace(string(body), "command: "+filepath.Join(driverDir, "driver"), "command: /unavailable-fixture", 1))
	// The original shared authoring owner retains its one-second retry delay.
	time.Sleep(1100 * time.Millisecond)

	restarted := newService(t.Context(), operations.env, operations.projectContext, testAssetsFS())
	if _, err := restarted.ContinueAuthoring(t.Context(), request); err != nil {
		t.Fatal("restart rebound to unavailable original profile setup", err)
	}

	request.Binding.RequestID = "independent-authoring"
	if _, err := restarted.ContinueAuthoring(t.Context(), request); err != nil {
		t.Fatal("fresh operation did not select replacement", err)
	}

	checkGenerationFixtureModels(t, project, map[string]string{"authoring-service-capture": "fixture-model", "authoring-independent-authoring": "other-model"})
	checkGenerationAuthoringAttempts(t, project, "service-capture", 2)
}

func harnessGenerationFixture(t *testing.T, output string) (service, string) {
	t.Helper()

	operations, project, configDir, _ := harnessSelectionFixture(t)
	driverDir := t.TempDir()
	driver := filepath.Join(driverDir, "driver")
	script := `#!/bin/sh
driver_dir=$(dirname "$0")
printf '%s\n' "$@" >> "$driver_dir/argv.log"
while [ "$#" -gt 0 ]; do
 if [ "$1" = --output-last-message ]; then shift; output=$1; fi
 shift
done
cat > "$driver_dir/prompt.log"
if [ -f "$driver_dir/fail-next" ]; then
 rm "$driver_dir/fail-next"
 printf 'owned provider failure' >&2
 exit 42
fi
cat "$driver_dir/result.json" > "$output"
`
	harnessSelectionWrite(t, driver, script)

	if err := os.Chmod(driver, 0o700); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(driverDir, "result.json"), output)

	profiles := "profiles:\n  writing:\n    harness: codex\n    command: " + driver + "\n    model: fixture-model\n  other:\n    harness: codex\n    command: " + driver + "\n    model: other-model\n"
	harnessSelectionWrite(t, filepath.Join(configDir, "harnesses.yml"), profiles)
	harnessSelectionWrite(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: writing\n")

	return operations, driverDir
}

func switchGenerationFixture(t *testing.T, operations service, profile string) {
	t.Helper()

	if err := changeGenerationFixture(t.Context(), operations, profile); err != nil {
		t.Fatal(err)
	}
}

func changeGenerationFixture(ctx context.Context, operations service, profile string) error {
	config, err := core.LoadWorkspaceConfig(operations.projectContext.ProjectDir)
	if err != nil {
		return err
	}

	request := core.HarnessSelectionChangeRequest{
		Binding:    workflowstate.MutationBinding{RequestID: "switch-" + profile, Operator: "maintainer"},
		ProjectRef: operations.projectContext.ProjectID, ExpectedRevision: config.Revision(), Generation: &profile,
	}
	_, err = operations.ChangeHarnessSelections(ctx, request, func(context.Context) error { return nil })

	return err
}

func generationFixtureMetadata(t *testing.T, project string) []harnesshost.HarnessRunMetadata {
	t.Helper()

	paths, err := filepath.Glob(workspacecore.UserPath(project, "tmp", "agent-runs", "*", "harness-run.json"))
	if err != nil {
		t.Fatal(err)
	}

	var records []harnesshost.HarnessRunMetadata

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var record harnesshost.HarnessRunMetadata
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}

		records = append(records, record)
	}

	return records
}

func checkGenerationFixtureModels(t *testing.T, project string, expected map[string]string) {
	t.Helper()

	counts := map[string]int{}

	for _, record := range generationFixtureMetadata(t, project) {
		if record.Selection.Selection.Model != expected[record.OperationID] || record.ProcessOutcome != "completed" || record.OutputValidation != "valid" {
			t.Fatal("caller lost retained selection or parser evidence", record)
		}

		counts[record.OperationID]++
	}

	if len(counts) != len(expected) {
		t.Fatal("missing operation metadata", counts)
	}
}

func checkGenerationAuthoringAttempts(t *testing.T, project, id string, attempts int) {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(project, workspacecore.DirectoryName, "authoring", "requests", "*.json"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		var request struct {
			Attempts int
			Request  authoringcore.ContinuationRequest
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(data, &request); err != nil {
			t.Fatal(err)
		}

		if request.Request.Binding.RequestID == id && request.Attempts == attempts {
			return
		}
	}

	t.Fatal("original authoring attempt count was not retained")
}
