package core

import (
	"fmt"
	"os"
)

// RuntimeValidationOptions describes a package boundary value.
type RuntimeValidationOptions struct {
	ProjectDir            string
	Config                RuntimeConfig
	Workspace             WorkWorkspaceConfig
	ExecutionWorkspaceDir string
	TicketKey             string
	TicketBranch          string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	RunID                 string
}

// RuntimeValidationResult describes a package boundary value.
type RuntimeValidationResult struct {
	Plan   RuntimePlan
	Checks []RuntimeValidationCheck
}

// RuntimeValidationCheck describes a package boundary value.
type RuntimeValidationCheck struct {
	Name    string
	Path    string
	OK      bool
	Message string
}

// ValidateRuntime validates the corresponding package boundary value.
func ValidateRuntime(opts RuntimeValidationOptions) (RuntimeValidationResult, error) {
	plan, err := PlanRuntime(RuntimePlanOptions(opts))
	if err != nil {
		return RuntimeValidationResult{}, err
	}

	checks := []RuntimeValidationCheck{
		checkDirectory("workspace", plan.WorkspaceDir),
		checkFile("dockerfile", plan.DockerfilePath),
		checkDirectory("artifact_dir", plan.ArtifactDir),
	}

	return RuntimeValidationResult{
		Plan:   plan,
		Checks: checks,
	}, nil
}

// RuntimeValidationPassed executes the corresponding package operation.
func RuntimeValidationPassed(checks []RuntimeValidationCheck) bool {
	for _, check := range checks {
		if !check.OK {
			return false
		}
	}

	return true
}

func checkDirectory(name, path string) RuntimeValidationCheck {
	info, err := os.Stat(path)
	if err != nil {
		return failedRuntimeCheck(name, path, err)
	}

	if !info.IsDir() {
		return RuntimeValidationCheck{
			Name:    name,
			Path:    path,
			OK:      false,
			Message: "path is not a directory",
		}
	}

	return passedRuntimeCheck(name, path)
}

func checkFile(name, path string) RuntimeValidationCheck {
	info, err := os.Stat(path)
	if err != nil {
		return failedRuntimeCheck(name, path, err)
	}

	if info.IsDir() {
		return RuntimeValidationCheck{
			Name:    name,
			Path:    path,
			OK:      false,
			Message: "path is a directory",
		}
	}

	return passedRuntimeCheck(name, path)
}

func passedRuntimeCheck(name, path string) RuntimeValidationCheck {
	return RuntimeValidationCheck{
		Name:    name,
		Path:    path,
		OK:      true,
		Message: "ok",
	}
}

func failedRuntimeCheck(name, path string, err error) RuntimeValidationCheck {
	return RuntimeValidationCheck{
		Name:    name,
		Path:    path,
		OK:      false,
		Message: fmt.Sprintf("%v", err),
	}
}
