package runtime

import (
	"fmt"
	"os"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

// ValidationOptions describes a package boundary value.
type ValidationOptions struct {
	ExtraMounts           []harness.RuntimeMount
	ProjectDir            string
	Config                Config
	Workspace             workspacecore.ExecutionConfig
	ExecutionWorkspaceDir string
	TicketKey             string
	TicketBranch          string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	RunID                 string
	OpsLayout             *opscore.Layout
	ArtifactID            string
}

// ValidationResult describes a package boundary value.
type ValidationResult struct {
	Plan   Plan
	Checks []ValidationCheck
}

// ValidationCheck describes a package boundary value.
type ValidationCheck struct {
	Name    string
	Path    string
	OK      bool
	Message string
}

// Validate validates the corresponding package boundary value.
func Validate(opts ValidationOptions) (ValidationResult, error) {
	plan, err := ResolvePlan(PlanOptions(opts))
	if err != nil {
		return ValidationResult{}, err
	}

	checks := []ValidationCheck{
		checkDirectory("execution_workspace", plan.ExecutionWorkspaceDir),
		checkDirectory("workspace", plan.WorkspaceDir),
		checkFile("dockerfile", plan.DockerfilePath),
		checkDirectory("artifact_dir", plan.ArtifactDir),
	}

	return ValidationResult{
		Plan:   plan,
		Checks: checks,
	}, nil
}

// ValidationPassed executes the corresponding package operation.
func ValidationPassed(checks []ValidationCheck) bool {
	for _, check := range checks {
		if !check.OK {
			return false
		}
	}

	return true
}

func checkDirectory(name, path string) ValidationCheck {
	info, err := os.Stat(path)
	if err != nil {
		return failedCheck(name, path, err)
	}

	if !info.IsDir() {
		return ValidationCheck{
			Name:    name,
			Path:    path,
			OK:      false,
			Message: "path is not a directory",
		}
	}

	return passedCheck(name, path)
}

func checkFile(name, path string) ValidationCheck {
	info, err := os.Stat(path)
	if err != nil {
		return failedCheck(name, path, err)
	}

	if info.IsDir() {
		return ValidationCheck{
			Name:    name,
			Path:    path,
			OK:      false,
			Message: "path is a directory",
		}
	}

	return passedCheck(name, path)
}

func passedCheck(name, path string) ValidationCheck {
	return ValidationCheck{
		Name:    name,
		Path:    path,
		OK:      true,
		Message: "ok",
	}
}

func failedCheck(name, path string, err error) ValidationCheck {
	return ValidationCheck{
		Name:    name,
		Path:    path,
		OK:      false,
		Message: fmt.Sprintf("%v", err),
	}
}
