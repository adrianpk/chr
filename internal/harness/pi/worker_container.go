package pi

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// WorkerRuntimeOptions is private retained command configuration. It contains
// references and model semantics, never credentials or mutable session state.
type WorkerRuntimeOptions struct {
	Image          string `json:"image"`
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	Thinking       string `json:"thinking,omitempty"`
	CredentialFile string `json:"credential_file"`
	ModelFile      string `json:"model_file,omitempty"`
	ModelHash      string `json:"model_hash,omitempty"`
	OAuth          bool   `json:"oauth,omitempty"`
}

// WorkerLaunch supplies host-owned mounts and the existing admitted capacity.
type WorkerLaunch struct {
	Options   WorkerRuntimeOptions
	Directory string
	Mounts    []harness.RuntimeMount
	Slot      string
	Guard     string
	Prompt    string
}

// WorkerCommand retains the complete bridge configuration in the host command.
func WorkerCommand(program string, options WorkerRuntimeOptions) ([]string, error) {
	if err := validateWorkerOptions(options); err != nil {
		return nil, err
	}

	data, err := json.Marshal(options)

	return []string{program, "_pi-worker", "--runtime", string(data)}, err
}

// RunWorkerContainer launches the bridge and Pi inside the immutable image.
// The host caller owns capacity, the task lifecycle and all result validation.
func RunWorkerContainer(ctx context.Context, request WorkerLaunch) (output []byte, err error) {
	defer func() {
		if err != nil {
			output = nil
		}
	}()

	if err := validateWorkerOptions(request.Options); err != nil {
		return nil, err
	}

	private, err := os.MkdirTemp("", "chr-pi-worker-")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(private)) }()

	command, name, err := prepareWorkerContainer(ctx, request, private)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, removeWorkerContainer(ctx, command.Program, name)) }()

	response, err := (process.Executor{Secrets: workerCredentialSecrets(request.Options.CredentialFile)}).Exchange(ctx, command, time.Now(), func(stdout io.Reader, stdin io.WriteCloser, _ context.CancelFunc) error {
		return exchangeWorkerInput(ctx, stdout, stdin, request.Prompt, &output)
	})
	if err != nil {
		return nil, &privateCallError{cause: err, message: "pi worker container failed: " + strings.TrimSpace(string(response.Stderr))}
	}

	return output, nil
}

func validateWorkerOptions(options WorkerRuntimeOptions) error {
	if err := ValidateGenerationProfile(options.Provider, options.Model, options.Thinking); err != nil {
		return err
	}

	if !regexp.MustCompile(`^sha256:[a-f0-9]{64}$`).MatchString(options.Image) || !filepath.IsAbs(options.CredentialFile) {
		return errors.New("pi worker requires an immutable prepared image and explicit private credentials")
	}

	if options.ModelFile != "" && (!filepath.IsAbs(options.ModelFile) || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(options.ModelHash)) {
		return errors.New("pi worker requires frozen private model semantics")
	}

	return nil
}

func prepareWorkerContainer(ctx context.Context, request WorkerLaunch, private string) (harness.AgentCommand, string, error) {
	options := request.Options

	resources, err := ResolveWorkerResources(ctx, "", options.CredentialFile, options.ModelFile, options.Provider)
	if err != nil || resources.ModelConfigHash != options.ModelHash || resources.OAuth != options.OAuth {
		return harness.AgentCommand{}, "", errors.New("pi worker resources changed or are unavailable; restore the admitted profile resources")
	}

	prepared, err := prepareWorkerMounts(request, private)
	if err != nil {
		return harness.AgentCommand{}, "", err
	}

	name := filepath.Base(private)

	args := workerContainerArguments(request, name)
	if prepared.gateway {
		args = append(args, "--add-host", "host.docker.internal:host-gateway")
	}

	for _, mount := range prepared.mounts {
		argument, err := workerMountArgument(mount)
		if err != nil {
			return harness.AgentCommand{}, "", err
		}

		args = append(args, "--mount", argument)
	}

	args = append(args, options.Image, "_pi-bridge", "--provider", options.Provider, "--model", options.Model, "--thinking", options.Thinking,
		"--credential-file", prepared.authTarget, "--slot", request.Slot, "--guard", request.Guard)
	if options.ModelFile != "" {
		args = append(args, "--model-file", "/chr-models.json")
	}

	return harness.AgentCommand{Program: "docker", Args: args, Dir: request.Directory}, name, nil
}

func workerContainerArguments(request WorkerLaunch, name string) []string {
	uid, gid := strconv.Itoa(os.Getuid()), strconv.Itoa(os.Getgid())

	return []string{
		"run", "--rm", "--pull=never", "--interactive", "--init", "--name", name, "--log-driver=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges",
		"--entrypoint", "/usr/local/bin/chr-worker-bridge",
		"--user", uid + ":" + gid, "--workdir", request.Directory,
		"--tmpfs", "/chr-call:rw,nosuid,nodev,mode=0700,uid=" + uid + ",gid=" + gid,
		"--tmpfs", "/tmp:rw,nosuid,nodev,mode=1777", "--env", "CHR_PI_WORKER_CONTAINER=1",
	}
}

func workerMountArgument(mount harness.RuntimeMount) (string, error) {
	if !filepath.IsAbs(mount.Source) || !filepath.IsAbs(mount.Target) || strings.ContainsAny(mount.Source+mount.Target, "\x00\r\n") {
		return "", errors.New("pi worker requires explicit absolute runtime mounts")
	}

	fields := []string{"type=bind", "src=" + mount.Source, "dst=" + mount.Target}
	if mount.ReadOnly {
		fields = append(fields, "readonly")
	}

	var text strings.Builder

	writer := csv.NewWriter(&text)
	if err := writer.Write(fields); err != nil {
		return "", err
	}

	writer.Flush()

	return strings.TrimSuffix(text.String(), "\n"), writer.Error()
}

func copyWorkerBridge(target string) (err error) {
	file, err := os.Open("/proc/self/exe")
	if err != nil {
		return errors.New("pi workers require a prepared Linux chr bridge")
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	copy, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o500)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, copy.Close()) }()

	n, err := io.Copy(copy, io.LimitReader(file, (512<<20)+1))
	if err != nil || n > 512<<20 {
		return errors.New("pi bridge exceeds executable copy bounds")
	}

	return nil
}

func removeWorkerContainer(ctx context.Context, program, name string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	// This exact random name is owned by the current call. An already removed
	// container is expected after --rm, including failed startup.
	result, err := (process.Executor{}).RunAgentCommand(ctx, harness.AgentCommand{Program: program, Args: []string{"rm", "--force", name}}, "", time.Now())
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err != nil && !strings.Contains(string(result.Stderr), "No such container") && !strings.Contains(string(result.Stderr), "No such object") {
		return errors.New("owned Pi container cleanup failed; verify Docker access")
	}

	return nil
}

func workerModelMount(options WorkerRuntimeOptions, private string) (*harness.RuntimeMount, bool, error) {
	if options.ModelFile == "" {
		return nil, false, nil
	}

	models, err := selectedModels(options.ModelFile, options.Provider)
	if err != nil || resourceHash(models) != options.ModelHash {
		return nil, false, errors.New("pi worker model semantics changed during preparation")
	}

	path := filepath.Join(private, "models.json")
	if err := os.WriteFile(path, models, 0o600); err != nil {
		return nil, false, err
	}

	gateway, err := workerModelNetwork(models)
	if err != nil {
		return nil, false, err
	}

	return &harness.RuntimeMount{Source: path, Target: "/chr-models.json", ReadOnly: true}, gateway, nil
}

type workerMountSet struct {
	mounts     []harness.RuntimeMount
	authTarget string
	gateway    bool
}

func prepareWorkerMounts(request WorkerLaunch, private string) (workerMountSet, error) {
	options := request.Options

	bridge := filepath.Join(private, "bridge")
	if err := copyWorkerBridge(bridge); err != nil {
		return workerMountSet{}, err
	}

	mounts := append([]harness.RuntimeMount(nil), request.Mounts...)
	mounts = append(mounts, harness.RuntimeMount{Source: bridge, Target: "/usr/local/bin/chr-worker-bridge", ReadOnly: true})
	mounts = append(mounts, harness.RuntimeMount{Source: request.Slot, Target: request.Slot, ReadOnly: true}, harness.RuntimeMount{Source: request.Guard, Target: request.Guard, ReadOnly: true})
	authTarget := "/chr-auth/auth.json"

	if options.OAuth {
		mounts = append(mounts, harness.RuntimeMount{Source: filepath.Dir(options.CredentialFile), Target: "/chr-auth"})
		authTarget = "/chr-auth/" + filepath.Base(options.CredentialFile)
	} else {
		mounts = append(mounts, harness.RuntimeMount{Source: options.CredentialFile, Target: authTarget, ReadOnly: true})
	}

	modelMount, gateway, err := workerModelMount(options, private)
	if err != nil {
		return workerMountSet{}, err
	}

	if modelMount != nil {
		mounts = append(mounts, *modelMount)
	}

	return workerMountSet{mounts: mounts, authTarget: authTarget, gateway: gateway}, nil
}
