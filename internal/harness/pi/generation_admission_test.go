package pi

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
)

// Failed preparation cannot modify a redirected artifact or reach the executable.
func TestPiGenerationAdmission(t *testing.T) {
	for _, scenario := range []string{"public artifact", "linked artifact", "artifact pipe", "linked directory", "invalid provider", "invalid model", "thinking", "input limit", "schema"} {
		t.Run(scenario, func(t *testing.T) {
			request, marker, retained := piAdmissionFixture(t)
			preparePiAdmissionFault(t, &request, retained, scenario)

			result, err := (Runner{}).RunAgent(t.Context(), request)
			if err == nil || len(result.Output) != 0 {
				t.Fatal("invalid preparation accepted")
			}

			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatal("invalid preparation reached the executable", err)
			}

			data, err := os.ReadFile(retained)
			if err != nil || string(data) != "Retain this file." {
				t.Fatal("invalid preparation changed an existing file", err)
			}
		})
	}
}

func piAdmissionFixture(t *testing.T) (harness.AgentRunRequest, string, string) {
	t.Helper()

	dir := t.TempDir()
	auth, marker := filepath.Join(dir, "auth.json"), filepath.Join(dir, "spawned")
	retained, program := filepath.Join(dir, "retained"), filepath.Join(dir, "driver")

	writePiFixtureFile(t, auth, validPiAuth)
	writePiFixtureFile(t, retained, "Retain this file.")
	writePiFixtureFile(t, program, "#!/bin/sh\ntouch '"+marker+"'\nexit 7\n")

	if err := os.Chmod(program, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(dir, "run"), 0o700); err != nil {
		t.Fatal(err)
	}

	return harness.AgentRunRequest{ID: "admission", Purpose: "test", ProjectDir: dir, RunDir: filepath.Join(dir, "run"), Program: program, FrozenSelection: true, ModelProvider: "fixture", Model: "model", Resources: harness.GenerationResources{CredentialFile: auth}, Prompt: "Return the owned fixture result."}, marker, retained
}

func preparePiAdmissionFault(t *testing.T, request *harness.AgentRunRequest, retained, scenario string) {
	t.Helper()

	switch scenario {
	case "invalid provider":
		request.ModelProvider = "bad\x00identity"
	case "invalid model":
		request.Model = "cli_default"
	case "thinking":
		request.Thinking = "unsupported"
	case "input limit":
		request.Prompt = strings.Repeat("x", harness.MaxRecord)
	case "schema":
		request.OutputSchema = []byte("invalid JSON")
	default:
		if err := preparePiArtifactFault(t, request, retained, scenario); err != nil {
			t.Fatal(err)
		}
	}
}

func preparePiArtifactFault(t *testing.T, request *harness.AgentRunRequest, retained, scenario string) error {
	t.Helper()

	prompt := filepath.Join(request.RunDir, "prompt.md")

	switch scenario {
	case "public artifact":
		writePiFixtureFile(t, prompt, "old")

		return os.Chmod(prompt, 0o644)
	case "linked artifact":
		return os.Link(retained, prompt)
	case "artifact pipe":
		return syscall.Mkfifo(prompt, 0o600)
	case "linked directory":
		link := filepath.Join(request.ProjectDir, "link")
		if err := os.Symlink(t.TempDir(), link); err != nil {
			return err
		}

		request.RunDir = filepath.Join(link, "outside")
	}

	return nil
}
