package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestTicketNewCreatesTicket(t *testing.T) {
	dir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"new",
		"--dir",
		dir,
		"--priority",
		"high",
		"--tag",
		"uxui",
		"Do something",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantPath := filepath.Join(dir, ".chc", "tickets", "CHC-1-Do-something.md")
	if !strings.Contains(stdout.String(), "ticket: "+wantPath) {
		t.Fatalf("stdout = %q, want ticket path %q", stdout.String(), wantPath)
	}

	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`key: "CHC-1"`,
		`title: "Do something"`,
		`priority: "high"`,
		`tags: ["uxui"]`,
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestTicketNewUsesConfigKeyPrefix(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, ".chc", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("tickets:\n  key_prefix: PRJ\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run(t.Context(), []string{"ticket", "new", "--dir", dir, "Do something"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, ".chc", "tickets", "PRJ-1-Do-something.md"))
}

func TestTicketNewUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Active project ticket",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket new returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Active-project-ticket.md"))
}

func TestTicketNewFallsBackToCurrentDirectory(t *testing.T) {
	projectDir := newTestWorkspace(t)

	runInDirectory(t, projectDir, func() {
		err := run(t.Context(), []string{
			"ticket",
			"new",
			"--dev-root",
			t.TempDir(),
			"Current project ticket",
		}, &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("ticket new returned error: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Current-project-ticket.md"))
}

func TestProjectContextResolverRejectsMissingContext(t *testing.T) {
	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	_, err = newCommandProjectContextResolver(env).Resolve(t.Context(), commandProjectContextOptions{})
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
}

func TestTicketNewRejectsMissingTitle(t *testing.T) {
	err := run(t.Context(), []string{"ticket", "new"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestProjectAddRegistersProject(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		"example",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want registered project", stdout.String())
	}

	assertPathExists(t, filepath.Join(devRoot, "data", "chc", "chc.db"))
}

func TestProjectAddRejectsProjectWithoutWorkspace(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		t.TempDir(),
		t.TempDir(),
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestProjectListPrintsRegisteredProjects(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		"example",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"list",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "example\t"+projectDir+"\ttrue\tproject-") {
		t.Fatalf("stdout = %q, want listed project", stdout.String())
	}
}

func TestProjectSelectByNameStoresCurrentProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var selectStdout bytes.Buffer

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &selectStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	if !strings.Contains(selectStdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want selected project", selectStdout.String())
	}

	var currentStdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		devRoot,
	}, &currentStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project current returned error: %v", err)
	}

	if !strings.Contains(currentStdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want current project", currentStdout.String())
	}

	assertPathExists(t, filepath.Join(devRoot, "config", "chc", "active-project"))
}

func TestProjectSelectByPathStoresCurrentProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project current returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want current project", stdout.String())
	}
}

func TestProjectCurrentRejectsMissingSelection(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		t.TempDir(),
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project current returned nil error")
	}
}

func TestProjectSelectRejectsUnknownProject(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		t.TempDir(),
		"missing",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func addProject(t *testing.T, devRoot, name, projectDir string) {
	t.Helper()

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		name,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}
}

func newTestWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	return projectDir
}

func chcTestUserEnvironment(t *testing.T) (chc.UserEnvironment, error) {
	t.Helper()

	return chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: t.TempDir(),
	})
}

func runInDirectory(t *testing.T, dir string, runFunc func()) {
	t.Helper()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	runFunc()
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat path %s: %v", path, err)
	}
}
