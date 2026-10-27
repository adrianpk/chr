package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Independent processes race against the same complete revision. Exactly one
// publishes both roles; the other conflicts rather than losing the first write.
func TestHarnessSelectionProcesses(t *testing.T) {
	project := t.TempDir()

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	children := []*harnessSelectionTestProcess{}
	for _, profile := range []string{"alpha", "beta"} {
		children = append(children, startHarnessSelectionTestProcess(t, ctx, project, profile))
	}

	for _, process := range children {
		if _, err := process.input.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}

		if err := process.input.Close(); err != nil {
			t.Fatal(err)
		}
	}

	accepted, conflicts := harnessSelectionProcessResults(t, children)

	config, err := LoadWorkspaceConfig(project)
	if err != nil || accepted != 1 || conflicts != 1 || config.Harnesses.Worker == "" || config.Harnesses.Worker != config.Harnesses.Generation {
		t.Fatal("cross-process compare-and-swap failed", accepted, conflicts, config.Harnesses, err)
	}
}

func harnessSelectionProcessResults(t *testing.T, children []*harnessSelectionTestProcess) (int, int) {
	t.Helper()

	accepted, conflicts := 0, 0

	for _, process := range children {
		if err := process.cmd.Wait(); err != nil {
			t.Fatal(process.output.String(), err)
		}

		if bytes.Contains(process.output.Bytes(), []byte("accepted")) {
			accepted++
		}

		if bytes.Contains(process.output.Bytes(), []byte("conflict")) {
			conflicts++
		}
	}

	return accepted, conflicts
}

type harnessSelectionTestProcess struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	output *bytes.Buffer
}

func startHarnessSelectionTestProcess(t *testing.T, ctx context.Context, project, profile string) *harnessSelectionTestProcess {
	t.Helper()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHarnessSelectionProcess$", "-test.v")

	cmd.Env = append(os.Environ(), "CHR_HARNESS_TEST_PROJECT="+project, "CHR_HARNESS_TEST_PROFILE="+profile)
	output := &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = output, output

	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if cmd.ProcessState != nil {
			return
		}

		_ = cmd.Process.Kill() //nolint:errcheck // Best-effort cleanup stops only this owned fixture.
		_ = cmd.Wait()         //nolint:errcheck // The owned fixture may already have exited.
	})

	return &harnessSelectionTestProcess{cmd: cmd, input: input, output: output}
}

// This helper runs only in the owned subprocess; the stdin byte is its start barrier.
func TestHarnessSelectionProcess(t *testing.T) {
	project := os.Getenv("CHR_HARNESS_TEST_PROJECT")
	if project == "" {
		return
	}

	if _, err := io.ReadFull(os.Stdin, make([]byte, 1)); err != nil {
		t.Fatal(err)
	}

	request := harnessSelectionTestRequest(os.Getenv("CHR_HARNESS_TEST_PROFILE"), contenthash.Sum(nil), os.Getenv("CHR_HARNESS_TEST_PROFILE"))

	_, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if errors.Is(err, workflowstate.ErrConflict) {
		t.Log("conflict")

		return
	}

	if err != nil {
		t.Fatal(err)
	}

	t.Log("accepted")
}
