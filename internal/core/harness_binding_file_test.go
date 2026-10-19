package core

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// Supported CLI distributions can exceed 256 MiB. Fingerprinting remains
// streaming and bounded, while the existing over-ceiling fixture still rejects.
func TestHarnessBindingLargeExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large-driver")

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	resizeErr := file.Truncate(300 << 20)
	if err := errors.Join(resizeErr, file.Close()); err != nil {
		t.Fatal(err)
	}

	if _, _, err := FreezeHarnessExecutable([]string{path}); err != nil {
		t.Fatal("supported CLI footprint was rejected", err)
	}
}

// Host preparation accepts resolved regular executables and rejects non-files
// and oversized images before reading or blocking on an executable candidate.
func TestHarnessBindingExecutableFiles(t *testing.T) {
	for _, kind := range []string{"regular", "symlink", "fifo", "directory", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			path := harnessBindingExecutableFixture(t, kind)
			reference, command, err := FreezeHarnessExecutable([]string{path})

			valid := kind == "regular" || kind == "symlink"
			if (err == nil) != valid {
				t.Fatal("executable type admission disagreed", err)
			}

			if valid && (!harnessSelectionDigest(reference.Hash) || !filepath.IsAbs(command[0]) || len(command) != 1) {
				t.Fatal("executable identity was not retained")
			}
		})
	}
}

func harnessBindingExecutableFixture(t *testing.T, kind string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "driver")

	switch kind {
	case "fifo":
		if err := syscall.Mkfifo(path, 0o700); err != nil {
			t.Fatal(err)
		}
	case "directory":
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	case "oversize":
		writeHarnessBindingOversizeFile(t, path)

	default:
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}

		if kind == "symlink" {
			link := filepath.Join(dir, "selected")
			if err := os.Symlink(path, link); err != nil {
				t.Fatal(err)
			}

			return link
		}
	}

	return path
}

func writeHarnessBindingOversizeFile(t *testing.T, path string) {
	t.Helper()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	resizeErr := file.Truncate(harnessExecutableLimit + 1)

	closeErr := file.Close()
	if resizeErr != nil || closeErr != nil {
		t.Fatal(resizeErr, closeErr)
	}
}
