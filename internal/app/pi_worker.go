package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

func newPiWorkerCommand(stdin io.Reader, stdout io.Writer) *cli.Command {
	return &cli.Command{Name: "_pi-worker", Hidden: true, Flags: []cli.Flag{&cli.StringFlag{Name: "runtime", Required: true}}, Action: func(ctx context.Context, cmd *cli.Command) error {
		if runtime.GOOS != "linux" || cmd.NArg() != 0 || len(cmd.String("runtime")) > 16_384 {
			return errors.New("pi workers require the prepared Linux chr bridge")
		}

		var options pi.WorkerRuntimeOptions

		decoder := json.NewDecoder(bytes.NewBufferString(cmd.String("runtime")))
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&options); err != nil {
			return errors.New("invalid retained Pi worker runtime")
		}

		return runPiWorker(ctx, stdin, stdout, options)
	}}
}

func runPiWorker(ctx context.Context, stdin io.Reader, stdout io.Writer, options pi.WorkerRuntimeOptions) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	data, err := io.ReadAll(io.LimitReader(stdin, harness.MaxInput+1))
	if err != nil || len(data) > harness.MaxInput {
		return errors.New("pi worker input exceeds its host boundary")
	}

	prompt, err := piWorkerPrompt(data, options)
	if err != nil {
		return err
	}

	directory, err := os.Getwd()
	if err != nil {
		return err
	}

	mounts, err := core.WorkerRuntimeMounts(ctx, directory)
	if err != nil {
		return err
	}

	instructions, err := core.WorkerRuntimeInstructions(directory)
	if err != nil {
		return err
	}

	contextData, err := json.Marshal(map[string]any{"task": json.RawMessage(prompt), "project_instructions": string(instructions)})
	if err != nil {
		return err
	}

	prompt = string(contextData)

	reservation, err := core.DelegateWorkerRuntimeReservation(ctx, directory)
	if err != nil {
		return err
	}

	output, err := pi.RunWorkerContainer(ctx, pi.WorkerLaunch{Options: options, Directory: directory, Mounts: mounts, Slot: reservation.Slot, Guard: reservation.Guard, Prompt: prompt})
	if err != nil {
		return err
	}

	_, err = stdout.Write(output)

	return err
}

func piWorkerPrompt(data []byte, options pi.WorkerRuntimeOptions) (string, error) {
	var input core.WorkflowTaskInput
	if json.Unmarshal(data, &input) != nil || input.Schema != "chr-task-input/v2" || input.HarnessBinding == nil {
		return "", errors.New("pi worker requires a retained v2 task input")
	}

	binding := *input.HarnessBinding
	if err := core.ValidateHarnessBinding(binding); err != nil {
		return "", err
	}

	if !validPiWorkerRole(input, options) {
		return "", errors.New("pi worker input conflicts with its frozen runtime or role")
	}

	// Retained host argv contains private file references. They are not model
	// instructions; the original input and its hash remain with the host.
	var prompt map[string]json.RawMessage
	if err := json.Unmarshal(data, &prompt); err != nil {
		return "", err
	}

	delete(prompt, "worker")

	clean, err := json.Marshal(prompt)

	return string(clean), err
}

func newPiBridgeCommand(stdin io.Reader, stdout io.Writer) *cli.Command {
	return &cli.Command{Name: "_pi-bridge", Hidden: true, Flags: []cli.Flag{
		&cli.StringFlag{Name: "provider", Required: true}, &cli.StringFlag{Name: "model", Required: true}, &cli.StringFlag{Name: "thinking"},
		&cli.StringFlag{Name: "credential-file", Required: true}, &cli.StringFlag{Name: "model-file"}, &cli.StringFlag{Name: "slot", Required: true}, &cli.StringFlag{Name: "guard", Required: true},
	}, Action: func(ctx context.Context, cmd *cli.Command) error {
		if os.Getenv("CHR_PI_WORKER_CONTAINER") != "1" || cmd.NArg() != 0 {
			return errors.New("pi RPC worker bridge requires its owned container")
		}

		if _, err := os.Stat("/.dockerenv"); err != nil {
			return errors.New("pi RPC worker bridge must execute inside Docker")
		}

		return runPiBridge(ctx, stdin, stdout, cmd)
	}}
}

func runPiBridge(ctx context.Context, stdin io.Reader, stdout io.Writer, cmd *cli.Command) (err error) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	prompt, ctx, cancel, err := pi.ReadWorkerInput(ctx, stdin)
	if err != nil {
		return err
	}
	defer cancel()

	release, err := core.HoldWorkerRuntimeReservation(ctx, core.WorkerRuntimeReservation{Slot: cmd.String("slot"), Guard: cmd.String("guard")})
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, release()) }()

	private, err := os.MkdirTemp("/chr-call", "call-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(private)) }()

	result, err := pi.RunWorkerBridge(ctx, pi.WorkerBridgeRequest{
		Provider: cmd.String("provider"), Model: cmd.String("model"), Thinking: cmd.String("thinking"),
		CredentialFile: cmd.String("credential-file"), ModelFile: cmd.String("model-file"), StateDir: filepath.Clean(private), Prompt: prompt,
	})
	if err != nil {
		return err
	}

	_, err = stdout.Write(result.Output)

	return err
}

func validPiWorkerRole(input core.WorkflowTaskInput, options pi.WorkerRuntimeOptions) bool {
	binding := input.HarnessBinding

	return binding.Harness == pi.Identity && binding.RuntimeRef == options.Image && binding.ModelProvider == options.Provider && binding.Model == options.Model && binding.Thinking == options.Thinking && input.Merger == nil && input.AgentCall == nil
}
