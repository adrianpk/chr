// Package runtime owns host runtime plans, launch commands and filesystem validation.
package runtime

// Config describes a package boundary value.
type Config struct {
	Image             string
	DockerfilePath    string
	Workdir           string
	ArtifactDir       string
	ValidationCommand string
}

// DefaultConfig executes the corresponding package operation.
func DefaultConfig() Config {
	return Config{
		Image:             DefaultImage,
		DockerfilePath:    DefaultDockerfilePath,
		Workdir:           DefaultWorkdir,
		ArtifactDir:       DefaultArtifactDir,
		ValidationCommand: DefaultValidationCommand,
	}
}

const (
	// DefaultImage defines a package boundary constant.
	DefaultImage = "chr-agent-runtime:local"
	// DefaultDockerfilePath defines a package boundary constant.
	DefaultDockerfilePath = ".chr/default/runtime/agent/Dockerfile"
	// DefaultWorkdir defines a package boundary constant.
	DefaultWorkdir = "/workspace"
	// DefaultArtifactDir defines the zero-configuration shared run artifact root.
	DefaultArtifactDir = "ops/default/run"
	// PrivateArtifactDir is the retired private run artifact root accepted only for migration to ops layout.
	PrivateArtifactDir = ".chr/default/runtime/runs"
	// DefaultValidationCommand defines a package boundary constant.
	DefaultValidationCommand = "make verify"
)
