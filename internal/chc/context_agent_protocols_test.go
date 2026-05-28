package chc

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAgentProtocolContextLoadsApplicableAgentsFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Root Protocol\n")
	writeTestFile(t, filepath.Join(dir, "internal", "AGENTS.md"), "# Internal Protocol\n")
	writeTestFile(t, filepath.Join(dir, "internal", "todo", "store.go"), "package todo\n")

	ctx, err := LoadAgentProtocolContext(AgentProtocolContextOptions{
		ProjectDir: dir,
		TargetPath: filepath.Join(dir, "internal", "todo", "store.go"),
	})
	if err != nil {
		t.Fatalf("LoadAgentProtocolContext returned error: %v", err)
	}

	if len(ctx.AgentFiles) != 2 {
		t.Fatalf("AgentFiles = %d, want 2", len(ctx.AgentFiles))
	}

	if ctx.AgentFiles[0].RelativePath != "AGENTS.md" {
		t.Fatalf("first AGENTS path = %q, want AGENTS.md", ctx.AgentFiles[0].RelativePath)
	}

	if ctx.AgentFiles[1].RelativePath != filepath.ToSlash(filepath.Join("internal", "AGENTS.md")) {
		t.Fatalf("second AGENTS path = %q, want internal/AGENTS.md", ctx.AgentFiles[1].RelativePath)
	}

	rendered := RenderProtocolsContext(ctx)
	for _, want := range []string{"# Root Protocol", "# Internal Protocol"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered protocols missing %q:\n%s", want, rendered)
		}
	}
}

func TestLoadAgentProtocolContextLoadsSelectedChcProtocolFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Root Protocol\n")
	writeTestFile(t, filepath.Join(dir, ".chc", "user", "manifest.md"), "# Manifest\n")
	writeTestFile(t, filepath.Join(dir, ".chc", "user", "config.yml"), "version: 1\n")
	writeTestFile(t, filepath.Join(dir, ".chc", "user", "agents", "CONVENTIONS.md"), "# Conventions\n")
	writeTestFile(t, filepath.Join(dir, ".chc", "user", "runtime", "agent", "Dockerfile"), "FROM golang:1.22\n")
	writeTestFile(t, filepath.Join(dir, ".chc", "user", "runtime", "runs", "exec-1", "runtime.log"), "skip\n")
	writeTestFile(t, filepath.Join(dir, ".chc", "user", "tickets", "CHC-1.md"), "# Ticket\n")

	ctx, err := LoadAgentProtocolContext(AgentProtocolContextOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("LoadAgentProtocolContext returned error: %v", err)
	}

	rendered := RenderProtocolsContext(ctx)
	for _, want := range []string{
		".chc/user/agents/CONVENTIONS.md",
		".chc/user/config.yml",
		".chc/user/manifest.md",
		".chc/user/runtime/agent/Dockerfile",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered protocols missing %q:\n%s", want, rendered)
		}
	}

	for _, unwanted := range []string{"runtime.log", "CHC-1.md"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("rendered protocols included %q:\n%s", unwanted, rendered)
		}
	}
}

func TestLoadAgentProtocolContextRecordsMissingRootAgentsSource(t *testing.T) {
	dir := t.TempDir()

	ctx, err := LoadAgentProtocolContext(AgentProtocolContextOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("LoadAgentProtocolContext returned error: %v", err)
	}

	var found bool

	for _, source := range ctx.Sources {
		if source.ID == "agents:AGENTS.md" {
			found = true

			if source.Included {
				t.Fatal("missing AGENTS.md source was marked included")
			}

			if !source.Required {
				t.Fatal("root AGENTS.md source was not marked required")
			}
		}
	}

	if !found {
		t.Fatal("missing root AGENTS.md source was not recorded")
	}
}

func TestRenderAgentContextIncludesDefaultGoInfrastructureIdentity(t *testing.T) {
	rendered := RenderAgentContext(DefaultAgentIdentity())

	for _, want := range []string{
		"Senior software engineer focused on Go",
		"Containerized agent execution",
		"Do not rely on implicit conversation history",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered agent context missing %q:\n%s", want, rendered)
		}
	}
}
