package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectWidePlanningBaselineDetectsExistingGoProject(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "project.md"), "# Project\n")
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "go.mod"), "module example.com/blog\n\ngo 1.24\n")
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "Makefile"), "test:\n\tgo test ./...\n")
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "cmd", "blog", "main.go"), "package main\n")
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "internal", "blog", "post.go"), "package blog\n")
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "internal", "blog", "post_test.go"), "package blog\n")

	operations := newTestService(t, projectDir)

	baseline, err := operations.projectWidePlanningBaseline()
	if err != nil {
		t.Fatalf("projectWidePlanningBaseline returned error: %v", err)
	}

	if baseline.DetectedProjectPlanKind != "project_gap_plan" {
		t.Fatalf("DetectedProjectPlanKind = %q, want project_gap_plan", baseline.DetectedProjectPlanKind)
	}

	if !baseline.ExistingImplementation {
		t.Fatal("ExistingImplementation = false, want true")
	}

	if baseline.GoModulePath != "example.com/blog" || baseline.GoVersion != "1.24" {
		t.Fatalf("go module summary = %q %q, want module and go version", baseline.GoModulePath, baseline.GoVersion)
	}

	if !containsString(baseline.Entrypoints, "cmd/blog/main.go") {
		t.Fatalf("Entrypoints = %#v, want cmd/blog/main.go", baseline.Entrypoints)
	}

	readiness := projectWidePlanningReadinessFor(baseline)
	if readiness.PlanKind != "project_gap_plan" || readiness.Status != "ready" {
		t.Fatalf("readiness = %#v, want ready project_gap_plan", readiness)
	}

	if !strings.Contains(readiness.GenerationGuidance, "gap-closure") {
		t.Fatalf("GenerationGuidance = %q, want gap-closure guidance", readiness.GenerationGuidance)
	}
}

func TestProjectWidePlanningBaselineDetectsEmptySetupProject(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "project.md"), "# Project\n")

	operations := newTestService(t, projectDir)

	baseline, err := operations.projectWidePlanningBaseline()
	if err != nil {
		t.Fatalf("projectWidePlanningBaseline returned error: %v", err)
	}

	if baseline.DetectedProjectPlanKind != "project_setup_plan" {
		t.Fatalf("DetectedProjectPlanKind = %q, want project_setup_plan", baseline.DetectedProjectPlanKind)
	}

	if !baseline.AppearsEmpty {
		t.Fatal("AppearsEmpty = false, want true")
	}

	readiness := projectWidePlanningReadinessFor(baseline)
	if readiness.Status != "partial" {
		t.Fatalf("readiness status = %q, want partial", readiness.Status)
	}

	if !containsString(readiness.BlockingSlots, "go_module") || !containsString(readiness.BlockingSlots, "go_version") {
		t.Fatalf("BlockingSlots = %#v, want go_module and go_version", readiness.BlockingSlots)
	}

	if !readiness.RequireProvisionalLabel {
		t.Fatal("RequireProvisionalLabel = false, want true")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}
