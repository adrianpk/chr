package pi

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// RuntimeImage identifies an explicitly prepared, immutable Linux Pi image.
// The reference is private setup input; only ID belongs in a frozen binding.
type RuntimeImage struct {
	ID           string
	Architecture string
}

// InspectRuntime reads local Docker metadata without building or pulling an image.
// It is called only for an explicitly selected Pi worker runtime.
func InspectRuntime(ctx context.Context, executor harness.AgentCommandExecutor, reference string) (RuntimeImage, error) {
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/:@-]{0,254}$`).MatchString(reference) {
		return RuntimeImage{}, errors.New("prepare a Pi Docker image and set runtime_ref to its local image reference")
	}

	if executor == nil {
		executor = process.Executor{StdoutLimit: 16_384, StderrLimit: 16_384}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	format := `{"id":{{json .Id}},"os":{{json .Os}},"architecture":{{json .Architecture}},"labels":{{json .Config.Labels}}}`

	result, err := executor.RunAgentCommand(ctx, harness.AgentCommand{Program: "docker", Args: []string{"image", "inspect", "--format", format, reference}}, "", time.Now())
	if err != nil {
		return RuntimeImage{}, errors.New("pi runtime unavailable: enable Docker access, install missing assets with chr runtime init and explicitly build the selected Pi image")
	}

	return parseRuntimeImage(result.Stdout)
}

type runtimeImageMetadata struct {
	ID           string            `json:"id"`
	OS           string            `json:"os"`
	Architecture string            `json:"architecture"`
	Labels       map[string]string `json:"labels"`
}

func parseRuntimeImage(data []byte) (RuntimeImage, error) {
	var metadata runtimeImageMetadata

	if json.Unmarshal(data, &metadata) != nil || !regexp.MustCompile(`^sha256:[a-f0-9]{64}$`).MatchString(metadata.ID) || !validRuntimePlatform(metadata) {
		return RuntimeImage{}, errors.New("pi runtime requires one inspectable Linux amd64 or arm64 image")
	}

	if metadata.Labels["org.chr.harness"] != Identity || metadata.Labels["org.chr.pi.version"] != "1.3.0" || metadata.Labels["org.chr.pi.protocol"] != ProtocolRevision || metadata.Labels["org.chr.node.version"] != "22.23.3" {
		return RuntimeImage{}, errors.New("selected image is not a prepared Pi 1.3.0 runtime; build the sibling runtime/agent/pi Dockerfile")
	}

	return RuntimeImage{ID: metadata.ID, Architecture: metadata.Architecture}, nil
}

func validRuntimePlatform(metadata runtimeImageMetadata) bool {
	return metadata.OS == "linux" && (metadata.Architecture == "amd64" || metadata.Architecture == "arm64")
}
