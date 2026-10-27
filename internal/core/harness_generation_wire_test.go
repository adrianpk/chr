package core

import (
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

type generationBindingWire struct {
	Schema       string                                 `json:"schema"`
	Operation    harnesshost.HarnessGenerationOperation `json:"operation"`
	Program      string                                 `json:"program,omitempty"`
	Resources    *harness.GenerationResources           `json:"resources,omitempty"`
	ConfigDir    string                                 `json:"config_dir"`
	ExecutionDir string                                 `json:"execution_dir"`
	Checksum     string                                 `json:"checksum"`
}

func generationBindingOperationPath(project, id string) (string, error) {
	return workspacecore.ConfigPath(project, filepath.Join("tmp", "agent-runs", "operations", contenthash.Sum([]byte(id))+".json"))
}

func generationBindingMetadataPath(request harness.AgentRunRequest) string {
	root := request.RunDir
	if root == "" {
		root = workspacecore.UserPath(request.ProjectDir, "tmp", "agent-runs", request.ID)
	}

	return filepath.Join(root, "harness-run.json")
}
