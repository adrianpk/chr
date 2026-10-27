// Package runtime owns host runtime plans, launch commands and filesystem validation.
package runtime

// RuntimeConfig describes a package boundary value.
type RuntimeConfig struct {
	Image             string
	DockerfilePath    string
	Workdir           string
	ArtifactDir       string
	ValidationCommand string
}

// DefaultRuntimeConfig executes the corresponding package operation.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Image:             DefaultRuntimeImage,
		DockerfilePath:    DefaultRuntimeDockerfilePath,
		Workdir:           DefaultRuntimeWorkdir,
		ArtifactDir:       DefaultRuntimeArtifactDir,
		ValidationCommand: DefaultRuntimeValidationCommand,
	}
}

const (
	// DefaultRuntimeImage defines a package boundary constant.
	DefaultRuntimeImage = "chr-agent-runtime:local"
	// DefaultRuntimeDockerfilePath defines a package boundary constant.
	DefaultRuntimeDockerfilePath = ".chr/default/runtime/agent/Dockerfile"
	// DefaultRuntimeWorkdir defines a package boundary constant.
	DefaultRuntimeWorkdir = "/workspace"
	// DefaultRuntimeArtifactDir defines the zero-configuration shared run artifact root.
	DefaultRuntimeArtifactDir = "ops/default/run"
	// PrivateRuntimeArtifactDir is the retired private run artifact root accepted only for migration to ops layout.
	PrivateRuntimeArtifactDir = ".chr/default/runtime/runs"
	// DefaultRuntimeValidationCommand defines a package boundary constant.
	DefaultRuntimeValidationCommand = "make verify"
)
