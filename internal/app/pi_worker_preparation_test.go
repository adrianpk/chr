package app

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

// Selecting native Pi, a missing program or a nonregular executable for the
// worker must fail locally without executing it or inspecting a Docker image.
func TestPiWorkerPreparation(t *testing.T) {
	for _, name := range []string{"missing", "native script", "directory", "pipe", "oversize"} {
		t.Run(name, func(t *testing.T) {
			program, marker := piPreparationProgram(t, name)
			profile := core.HarnessProfile{Harness: "pi", Command: program, ModelProvider: "fixture", Model: "model"}

			readiness, image := piProfileReadiness(t.Context(), profile, harnesscore.Worker, "")
			if readiness.Ready() || image != "" || !strings.Contains(readiness.Problem, "chr worker bridge") {
				t.Fatal("unprepared worker executable was accepted", readiness)
			}

			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatal("preparation executed a configured program")
			}
		})
	}
}

func piPreparationProgram(t *testing.T, name string) (string, string) {
	t.Helper()

	program, marker := filepath.Join(t.TempDir(), "program"), filepath.Join(t.TempDir(), "executed")

	var err error

	switch name {
	case "native script":
		err = os.WriteFile(program, []byte("#!/bin/sh\ntouch '"+marker+"'\n"), 0o700)
	case "directory":
		err = os.Mkdir(program, 0o700)
	case "pipe":
		err = syscall.Mkfifo(program, 0o700)
	case "oversize":
		file, createErr := os.OpenFile(program, os.O_CREATE|os.O_WRONLY, 0o700)
		if createErr != nil {
			t.Fatal(createErr)
		}

		err = file.Truncate((512 << 20) + 1)
		if closeErr := file.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}

	if err != nil {
		t.Fatal(err)
	}

	return program, marker
}
