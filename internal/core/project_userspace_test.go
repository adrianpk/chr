package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectUserspaceUsesActiveOperatorEnvironment(t *testing.T) {
	dir := t.TempDir()

	userspace, err := ResolveProjectUserspace(ResolveProjectUserspaceOptions{
		ProjectDir: dir,
		LookupEnv: func(key string) string {
			if key != ActiveOperatorIDEnv {
				t.Fatalf("env key = %q, want %s", key, ActiveOperatorIDEnv)
			}

			return "adrian"
		},
	})
	if err != nil {
		t.Fatalf("ResolveProjectUserspace returned error: %v", err)
	}

	if userspace.OperatorID != "adrian" {
		t.Fatalf("OperatorID = %q, want adrian", userspace.OperatorID)
	}

	if userspace.Path != filepath.Join(dir, DirectoryName, "user") {
		t.Fatalf("Path = %q, want .chr/user", userspace.Path)
	}

	if userspace.RelativeDir != ".chr/user" {
		t.Fatalf("RelativeDir = %q, want .chr/user", userspace.RelativeDir)
	}
}

func TestResolveProjectUserspaceRequiresOperatorEnvironment(t *testing.T) {
	_, err := ResolveProjectUserspace(ResolveProjectUserspaceOptions{
		ProjectDir: t.TempDir(),
		LookupEnv:  func(string) string { return "" },
	})
	if err == nil {
		t.Fatal("ResolveProjectUserspace returned nil error")
	}
}

func TestResolveProjectUserspaceRejectsInvalidOperatorID(t *testing.T) {
	_, err := ResolveProjectUserspace(ResolveProjectUserspaceOptions{
		ProjectDir: t.TempDir(),
		LookupEnv:  func(string) string { return "../adrian" },
	})
	if err == nil {
		t.Fatal("ResolveProjectUserspace returned nil error")
	}
}

func TestEnsureProjectUserspaceCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")

	userspace, err := EnsureProjectUserspace(context.Background(), EnsureProjectUserspaceOptions{
		ProjectDir: dir,
	})
	if err != nil {
		t.Fatalf("EnsureProjectUserspace returned error: %v", err)
	}

	info, err := os.Stat(userspace.Path)
	if err != nil {
		t.Fatalf("stat userspace: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("userspace path is not a directory: %s", userspace.Path)
	}
}

func TestIsValidOperatorID(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "adrian", want: true},
		{value: "adrian.pk", want: true},
		{value: "adrian_pk", want: true},
		{value: "adrian-pk", want: true},
		{value: "", want: false},
		{value: ".", want: false},
		{value: "..", want: false},
		{value: "adrian/pk", want: false},
		{value: "adrian pk", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := IsValidOperatorID(tt.value); got != tt.want {
				t.Fatalf("IsValidOperatorID(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
