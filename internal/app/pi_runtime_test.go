package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/pi"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// Every bundled profile and the explicit missing-asset upgrade must install the
// locked Pi runtime without replacing the managed project's Codex runtime.
func TestPiRuntimeAssets(t *testing.T) {
	for _, profile := range []string{"standard", "minimal", "personal"} {
		t.Run(profile, func(t *testing.T) {
			project, agent := initPiRuntimeProject(t, profile)
			assertPiRuntimeUpgrade(t, project)
			assertPiRuntimeFiles(t, agent)
		})
	}
}

func initPiRuntimeProject(t *testing.T, profile string) (string, string) {
	t.Helper()

	project := t.TempDir()
	ctx := context.Background()

	_, err := core.NewWorkspace(testAssetsFS()).Init(ctx, core.InitOptions{ProjectDir: project, Profile: profile})
	if err != nil {
		t.Fatal(err)
	}

	agent := workspacecore.UserPath(project, "runtime", "agent")
	if err := os.WriteFile(filepath.Join(agent, "Dockerfile"), []byte("FROM managed-codex\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	assertPiRuntimeFiles(t, agent)

	if err := os.RemoveAll(filepath.Join(agent, "pi")); err != nil {
		t.Fatal(err)
	}

	return project, agent
}

func assertPiRuntimeUpgrade(t *testing.T, project string) {
	t.Helper()

	ctx := context.Background()
	agent := workspacecore.UserPath(project, "runtime", "agent")

	assets := core.NewRuntimeAssets(testAssetsFS())

	preview, err := assets.Init(ctx, core.RuntimeInitOptions{ProjectDir: project, DryRun: true})
	if err != nil || len(preview.Created) != 3 {
		t.Fatalf("preview = %#v, %v", preview, err)
	}

	if _, err := os.Stat(filepath.Join(agent, "pi")); !os.IsNotExist(err) {
		t.Fatalf("dry-run installed Pi: %v", err)
	}

	result, err := assets.Init(ctx, core.RuntimeInitOptions{ProjectDir: project})
	if err != nil || len(result.Created) != 3 {
		t.Fatalf("upgrade = %#v, %v", result, err)
	}

	repeat, err := assets.Init(ctx, core.RuntimeInitOptions{ProjectDir: project})
	if err != nil || len(repeat.Created) != 0 {
		t.Fatalf("repeated upgrade = %#v, %v", repeat, err)
	}
}

func assertPiRuntimeFiles(t *testing.T, agent string) {
	t.Helper()

	for _, name := range []string{"Dockerfile", "package.json", "package-lock.json"} {
		if info, err := os.Stat(filepath.Join(agent, "pi", name)); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("missing Pi %s: %v", name, err)
		}
	}

	codex, err := os.ReadFile(filepath.Join(agent, "Dockerfile"))
	if err != nil || string(codex) != "FROM managed-codex\n" {
		t.Fatalf("Codex overwritten: %q, %v", codex, err)
	}
}

// Opt in to inspect and execute one explicitly built image. The image metadata
// and its actual binaries must agree; no pull, model call or credentials occur.
func TestPiRuntimeImage(t *testing.T) {
	image := os.Getenv("CHR_TEST_PI_RUNTIME_IMAGE")
	if image == "" {
		t.Skip("set CHR_TEST_PI_RUNTIME_IMAGE to an explicitly built Pi image")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	metadata, err := pi.InspectRuntime(ctx, nil, image)
	if err != nil {
		t.Fatal(err)
	}

	command := harness.AgentCommand{Program: "docker", Args: []string{"run", "--rm", "--pull=never", "--network=none", metadata.ID, "sh", "-c", `test "$(node --version)" = v22.23.3 && test "$(pi --version)" = 1.3.0`}}
	if result, err := (process.Executor{}).RunAgentCommand(ctx, command, "", time.Now()); err != nil {
		t.Fatalf("image binaries: %v: %s", err, result.Stderr)
	}
}
