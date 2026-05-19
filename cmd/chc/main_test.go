package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRuntimeInitCreatesTemplateWithExplicitDir(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))

	if !strings.Contains(stdout.String(), "create .chc/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile create", stdout.String())
	}
}

func TestRuntimeInitDryRunDoesNotCreateTemplate(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathNotExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))

	if !strings.Contains(stdout.String(), "create .chc/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile create", stdout.String())
	}
}

func TestRuntimeInitSkipsExistingTemplate(t *testing.T) {
	projectDir := newTestWorkspace(t)

	dockerfilePath := filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	if err := os.WriteFile(dockerfilePath, []byte("FROM custom\n"), 0o644); err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read runtime template: %v", err)
	}

	if string(content) != "FROM custom\n" {
		t.Fatalf("runtime template content = %q, want custom content unchanged", string(content))
	}

	if !strings.Contains(stdout.String(), "skip existing .chc/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile skip", stdout.String())
	}
}

func TestRuntimeInitUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))
}

func TestRuntimeInitFallsBackToCurrentDirectory(t *testing.T) {
	projectDir := newTestWorkspace(t)

	runInDirectory(t, projectDir, func() {
		err := run(t.Context(), []string{
			"runtime",
			"init",
			"--dev-root",
			t.TempDir(),
		}, &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("runtime init returned error: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))
}

func TestRuntimeValidatePrintsRuntimePlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeRuntimeTemplate(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime validate returned error: %v", err)
	}

	for _, want := range []string{
		"dry_run: true",
		"image: chc-agent-runtime:local",
		"dockerfile: " + filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
		"workdir: /workspace",
		"artifact_dir: " + filepath.Join(projectDir, ".chc", "runtime", "runs"),
		"mount: " + projectDir + " -> /workspace read_only=false",
		"check: ok workspace " + filepath.Join(projectDir, ".chc") + " ok",
		"check: ok dockerfile " + filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile") + " ok",
		"check: ok artifact_dir " + filepath.Join(projectDir, ".chc", "runtime", "runs") + " ok",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestRuntimeValidateFailsWhenRuntimeTemplateIsMissing(t *testing.T) {
	projectDir := newTestWorkspace(t)
	if err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755); err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("runtime validate returned nil error")
	}

	if !strings.Contains(stdout.String(), "check: fail dockerfile") {
		t.Fatalf("stdout = %q, want failed Dockerfile check", stdout.String())
	}
}

func TestRuntimeValidateUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeRuntimeTemplate(t, projectDir)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime validate returned error: %v", err)
	}
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

func TestTicketListPrintsIndexedTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"list",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket list returned error: %v", err)
	}

	want := "CHC-1\tDo something\topen\thigh\tnormal\ttask\tonboarding\tCHC-1-Do-something.md"
	if !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want ticket row %q", stdout.String(), want)
	}
}

func TestTicketShowPrintsIndexedTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket show returned error: %v", err)
	}

	for _, want := range []string{
		"key: CHC-1",
		"title: Do something",
		"status: open",
		"priority: high",
		"severity: normal",
		"type: task",
		"use_case: onboarding",
		"project_area: uxui",
		"path: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestTicketShowRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket show returned nil error")
	}
}

func TestTicketCommentAppendsComment(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"comment",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--author",
		"adrian",
		"CHC-1",
		"Looks ready.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket comment returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: "+ticketPath) {
		t.Fatalf("stdout = %q, want ticket path", stdout.String())
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		"# Comments",
		"adrian",
		"\"Looks ready.\"",
	} {
		if !strings.Contains(string(updatedContent), want) {
			t.Fatalf("updated content does not contain %q:\n%s", want, string(updatedContent))
		}
	}
}

func writeIndexedTicketFile(t *testing.T, devRoot, projectDir string) string {
	t.Helper()

	ticketPath := filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Do-something.md")

	content, err := chc.NewTicketRenderer().RenderTicket(chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(ticketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(ticketPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	indexTicket(t, devRoot, projectDir, chc.IndexedTicket{
		Key:          "CHC-1",
		Title:        "Do something",
		Status:       "open",
		RelativePath: filepath.Join(".chc", "tickets", "CHC-1-Do-something.md"),
		Path:         ticketPath,
	})

	return ticketPath
}

func TestTicketCommentRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"comment",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"Looks ready.",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket comment returned nil error")
	}
}

func TestTicketStatusUpdatesTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"ready",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket status returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: CHC-1 ready") {
		t.Fatalf("stdout = %q, want status result", stdout.String())
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	if !strings.Contains(string(updatedContent), `status: "ready"`) {
		t.Fatalf("updated content does not contain ready status:\n%s", string(updatedContent))
	}
}

func TestTicketStatusRejectsInvalidStatus(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeIndexedTicketFile(t, devRoot, projectDir)
	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"unknown",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestTicketStatusRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"ready",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestIndexRebuildIndexesTicketFiles(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	indexTicket(t, devRoot, projectDir, chc.IndexedTicket{
		Key:          "CHC-9",
		Title:        "Old",
		Status:       "open",
		RelativePath: "CHC-9-Old.md",
		Path:         filepath.Join(projectDir, ".chc", "tickets", "CHC-9-Old.md"),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	for _, want := range []string{
		"indexed: 1",
		"parse_errors: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}

	tickets := listProjectTickets(t, devRoot, projectDir)
	if len(tickets) != 1 {
		t.Fatalf("len(tickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHC-1" || tickets[0].Title != "Do something" {
		t.Fatalf("ticket = %#v, want rebuilt CHC-1", tickets[0])
	}
}

func TestIndexRebuildReportsParseErrors(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	badTicketPath := filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Bad.md")
	if err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644); err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	for _, want := range []string{
		"indexed: 0",
		"parse_errors: 1",
		"parse_error: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsCleanIndex(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	if err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index status returned error: %v", err)
	}

	for _, want := range []string{
		"missing: 0",
		"stale: 0",
		"orphaned: 0",
		"invalid: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsMissingTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"missing: 1",
		"missing_ticket: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsStaleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	if err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something else",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"stale: 1",
		"stale_ticket: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsInvalidTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	badTicketPath := filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Bad.md")
	if err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644); err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"invalid: 1",
		"invalid_ticket: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
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

func TestProjectSelectRejectsMissingRegisteredPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestProjectSelectRejectsRegisteredPathWithoutWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	if err := os.RemoveAll(filepath.Join(projectDir, ".chc")); err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestTicketNewRejectsMissingActiveProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err := run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket new returned nil error")
	}
}

func TestTicketNewRejectsActiveProjectWithoutWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	if err := os.RemoveAll(filepath.Join(projectDir, ".chc")); err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err := run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket new returned nil error")
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

func selectProject(t *testing.T, devRoot, selector string) {
	t.Helper()

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		selector,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}
}

func indexTicket(t *testing.T, devRoot, projectDir string, ticket chc.IndexedTicket) {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	index, err := chc.NewTicketIndex(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewTicketIndex returned error: %v", err)
	}

	defer func() {
		if err := index.Close(); err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	ticket.ProjectID = projectID(projectDir)
	ticket.FileMTime = now
	ticket.FileSize = 100
	ticket.ContentHash = "hash"
	ticket.IndexedAt = now

	if err := index.UpsertTicket(t.Context(), ticket); err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}
}

func listProjectTickets(t *testing.T, devRoot, projectDir string) []chc.IndexedTicket {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	index, err := chc.NewTicketIndex(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewTicketIndex returned error: %v", err)
	}

	defer func() {
		if err := index.Close(); err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	tickets, err := index.ListProjectTickets(t.Context(), projectID(projectDir))
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	return tickets
}

func writeTicketMarkdownOnly(t *testing.T, projectDir, relativePath string, ticket chc.Ticket) string {
	t.Helper()

	ticketPath := filepath.Join(projectDir, ".chc", "tickets", filepath.FromSlash(relativePath))

	content, err := chc.NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(ticketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(ticketPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return ticketPath
}

func newTestWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	return projectDir
}

func writeRuntimeTemplate(t *testing.T, projectDir string) {
	t.Helper()

	dockerfilePath := filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	if err := os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644); err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755); err != nil {
		t.Fatalf("create runtime runs directory: %v", err)
	}
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

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("path %s exists or returned unexpected stat error: %v", path, err)
	}
}
