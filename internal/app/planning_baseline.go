//nolint:wsl_v5 // Baseline extraction keeps small parsing steps intentionally adjacent.
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const projectWidePlanningMaxFiles = 80

type projectWidePlanningBaseline struct {
	ProjectDir              string   `json:"project_dir"`
	WorkspaceDirExists      bool     `json:"workspace_dir_exists"`
	NonWorkspaceFileCount   int      `json:"non_workspace_file_count"`
	AppearsEmpty            bool     `json:"appears_empty"`
	GoModuleExists          bool     `json:"go_module_exists"`
	GoModulePath            string   `json:"go_module_path,omitempty"`
	GoVersion               string   `json:"go_version,omitempty"`
	GoPackageDirs           []string `json:"go_package_dirs,omitempty"`
	Entrypoints             []string `json:"entrypoints,omitempty"`
	MakefileExists          bool     `json:"makefile_exists"`
	ReadmeExists            bool     `json:"readme_exists"`
	TestFileCount           int      `json:"test_file_count"`
	ExistingImplementation  bool     `json:"existing_implementation"`
	DetectedProjectPlanKind string   `json:"detected_project_plan_kind"`
	Files                   []string `json:"files,omitempty"`
}

type projectWidePlanningReadiness struct {
	Status                  string   `json:"status"`
	PlanKind                string   `json:"plan_kind"`
	BlockingSlots           []string `json:"blocking_slots"`
	HighImpactSlots         []string `json:"high_impact_slots"`
	OptionalSlots           []string `json:"optional_slots"`
	GenerationGuidance      string   `json:"generation_guidance"`
	ShouldGeneratePlan      bool     `json:"should_generate_plan"`
	RequireProvisionalLabel bool     `json:"require_provisional_label"`
}

func (s service) projectWidePlanningSections() ([]core.QuizModelContextSection, error) {
	baseline, err := s.projectWidePlanningBaseline()
	if err != nil {
		return nil, err
	}

	readiness := projectWidePlanningReadinessFor(baseline)

	baselineContent, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project-wide planning baseline: %w", err)
	}

	readinessContent, err := json.MarshalIndent(readiness, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project-wide planning readiness: %w", err)
	}

	slotsContent, err := json.MarshalIndent(projectWidePlanningSlotCatalog(), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project-wide planning slots: %w", err)
	}

	return []core.QuizModelContextSection{
		{
			Name:      "project_wide_planning_baseline",
			Layer:     "project",
			Authority: "deterministic_baseline_scan",
			Content:   string(baselineContent),
			Metadata: map[string]string{
				"plan_scope": "project_wide",
			},
		},
		{
			Name:      "project_wide_planning_readiness",
			Layer:     "project",
			Authority: "deterministic_readiness_gate",
			Content:   string(readinessContent),
			Metadata: map[string]string{
				"plan_scope": "project_wide",
				"status":     readiness.Status,
				"plan_kind":  readiness.PlanKind,
			},
		},
		{
			Name:      "project_wide_planning_slots",
			Layer:     "project",
			Authority: "semantic_slot_catalog",
			Content:   string(slotsContent),
			Metadata: map[string]string{
				"plan_scope": "project_wide",
			},
		},
	}, nil
}

func (s service) projectWidePlanningBaseline() (projectWidePlanningBaseline, error) {
	baseline := projectWidePlanningBaseline{
		ProjectDir:         s.projectContext.ProjectDir,
		WorkspaceDirExists: pathExists(filepath.Join(s.projectContext.ProjectDir, workspacecore.DirectoryName)),
		MakefileExists:     pathExists(filepath.Join(s.projectContext.ProjectDir, "Makefile")),
		ReadmeExists:       pathExists(filepath.Join(s.projectContext.ProjectDir, "README.md")),
	}

	goModPath := filepath.Join(s.projectContext.ProjectDir, "go.mod")
	goMod, err := readGoModSummary(goModPath)
	if err != nil {
		return projectWidePlanningBaseline{}, err
	}
	baseline.GoModuleExists = goMod.exists
	baseline.GoModulePath = goMod.modulePath
	baseline.GoVersion = goMod.goVersion

	files, err := planningBaselineFiles(s.projectContext.ProjectDir)
	if err != nil {
		return projectWidePlanningBaseline{}, err
	}

	baseline.Files = files
	baseline.NonWorkspaceFileCount = nonWorkspaceFileCount(files)
	baseline.AppearsEmpty = baseline.NonWorkspaceFileCount == 0
	baseline.GoPackageDirs = goPackageDirs(files)
	baseline.Entrypoints = entrypointFiles(files)
	baseline.TestFileCount = testFileCount(files)
	baseline.ExistingImplementation = baseline.GoModuleExists ||
		len(baseline.GoPackageDirs) > 0 ||
		len(baseline.Entrypoints) > 0 ||
		baseline.TestFileCount > 0
	baseline.DetectedProjectPlanKind = detectedProjectPlanKind(baseline)

	return baseline, nil
}

type goModSummary struct {
	exists     bool
	modulePath string
	goVersion  string
}

func readGoModSummary(path string) (goModSummary, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return goModSummary{}, nil
	}
	if err != nil {
		return goModSummary{}, fmt.Errorf("read go.mod: %w", err)
	}

	summary := goModSummary{exists: true}
	for _, rawLine := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(rawLine)
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "module":
			summary.modulePath = fields[1]
		case "go":
			summary.goVersion = fields[1]
		}
	}

	return summary, nil
}

func planningBaselineFiles(projectDir string) ([]string, error) {
	files := []string{}

	err := filepath.WalkDir(projectDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == projectDir {
			return nil
		}

		relative, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		if entry.IsDir() && skipPlanningBaselineDir(relative) {
			return filepath.SkipDir
		}

		if entry.IsDir() {
			return nil
		}

		files = append(files, relative)
		if len(files) >= projectWidePlanningMaxFiles {
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan project baseline files: %w", err)
	}

	sort.Strings(files)

	return files, nil
}

func skipPlanningBaselineDir(path string) bool {
	switch path {
	case ".git", ".chr", "vendor", "node_modules", "dist", "build", "tmp":
		return true
	default:
		return strings.HasPrefix(path, ".chr/") ||
			strings.HasPrefix(path, ".git/") ||
			strings.HasPrefix(path, "vendor/") ||
			strings.HasPrefix(path, "node_modules/")
	}
}

func nonWorkspaceFileCount(files []string) int {
	count := 0
	for _, file := range files {
		if strings.HasPrefix(file, ".chr/") || strings.HasPrefix(file, ".git/") {
			continue
		}

		count++
	}

	return count
}

func goPackageDirs(files []string) []string {
	seen := map[string]bool{}
	dirs := []string{}

	for _, file := range files {
		if !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			continue
		}

		dir := filepath.ToSlash(filepath.Dir(file))
		if dir == "." {
			dir = ""
		}
		if seen[dir] {
			continue
		}

		seen[dir] = true
		dirs = append(dirs, dir)
	}

	sort.Strings(dirs)

	return dirs
}

func entrypointFiles(files []string) []string {
	entrypoints := []string{}
	for _, file := range files {
		if strings.HasPrefix(file, "cmd/") && strings.HasSuffix(file, ".go") && !strings.HasSuffix(file, "_test.go") {
			entrypoints = append(entrypoints, file)
		}
	}

	sort.Strings(entrypoints)

	return entrypoints
}

func testFileCount(files []string) int {
	count := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			count++
		}
	}

	return count
}

func detectedProjectPlanKind(baseline projectWidePlanningBaseline) string {
	if baseline.AppearsEmpty || !baseline.ExistingImplementation {
		return "project_setup_plan"
	}

	return "project_gap_plan"
}

func projectWidePlanningReadinessFor(baseline projectWidePlanningBaseline) projectWidePlanningReadiness {
	readiness := projectWidePlanningReadiness{
		Status:             "ready",
		PlanKind:           baseline.DetectedProjectPlanKind,
		ShouldGeneratePlan: true,
		OptionalSlots: []string{
			"validation_command",
			"delivery_unit",
			"runtime_constraints",
		},
	}

	if baseline.DetectedProjectPlanKind == "project_setup_plan" {
		readiness.GenerationGuidance = "Generate a setup-oriented plan. The project appears empty or has no Go implementation baseline; use deterministic Go setup assumptions only when missing slots remain unresolved."
		readiness.BlockingSlots = missingSetupBlockingSlots(baseline)
		readiness.HighImpactSlots = []string{"binary_or_library", "storage_model"}
		if len(readiness.BlockingSlots) > 0 {
			readiness.Status = "partial"
			readiness.RequireProvisionalLabel = true
		}

		return readiness
	}

	readiness.GenerationGuidance = "Generate a gap-closure plan for an existing Go project. Do not propose greenfield scaffold, go.mod creation, Makefile creation, or implementation that baseline shows already exists unless the plan explicitly modifies or replaces it."
	readiness.BlockingSlots = missingExistingProjectBlockingSlots(baseline)
	readiness.HighImpactSlots = []string{"slicing_target", "known_existing_code"}
	if len(readiness.BlockingSlots) > 0 {
		readiness.Status = "partial"
		readiness.RequireProvisionalLabel = true
	}

	return readiness
}

func missingSetupBlockingSlots(baseline projectWidePlanningBaseline) []string {
	missing := []string{}
	if baseline.GoModulePath == "" {
		missing = append(missing, "go_module")
	}
	if baseline.GoVersion == "" {
		missing = append(missing, "go_version")
	}

	return missing
}

func missingExistingProjectBlockingSlots(baseline projectWidePlanningBaseline) []string {
	missing := []string{}
	if baseline.GoModulePath == "" {
		missing = append(missing, "go_module")
	}
	if baseline.GoVersion == "" {
		missing = append(missing, "go_version")
	}
	if len(baseline.Entrypoints) == 0 && len(baseline.GoPackageDirs) == 0 {
		missing = append(missing, "entrypoints")
	}

	return missing
}

func projectWidePlanningSlotCatalog() map[string]any {
	return map[string]any{
		"slot_order": []string{
			"project_goal",
			"project_mode",
			"go_module",
			"go_version",
			"binary_or_library",
			"entrypoints",
			"storage_model",
			"runtime_constraints",
			"validation_command",
			"delivery_unit",
			"known_existing_code",
			"slicing_target",
		},
		"blocking_slots": []string{
			"project_goal",
			"project_mode",
			"go_module",
			"go_version",
		},
		"high_impact_slots": []string{
			"binary_or_library",
			"entrypoints",
			"known_existing_code",
			"slicing_target",
		},
		"optional_slots": []string{
			"storage_model",
			"runtime_constraints",
			"validation_command",
			"delivery_unit",
		},
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
