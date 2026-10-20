package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Opting in with missing private credentials is a real failure before any
// model executable runs. The child receives only synthetic fixture choices.
func TestHarnessLivePiGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("real Git fixture is selected explicitly")
	}

	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}

	marker, driver := filepath.Join(root, "spawned"), filepath.Join(root, "driver")
	harnessSelectionWrite(t, driver, "#!/bin/sh\ntouch '"+marker+"'\nexit 99\n")

	if err := os.Chmod(driver, 0o700); err != nil {
		t.Fatal(err)
	}

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), program, "-test.run=^TestHarnessLivePi$", "-test.timeout=30s")

	command.Env = append(os.Environ(), "CHR_LIVE_PI=1", "CHR_LIVE_PI_ROOT="+root, "CHR_LIVE_PI_PROVIDER=fixture", "CHR_LIVE_PI_MODEL=fixture-model", "CHR_LIVE_PI_NATIVE="+driver, "CHR_LIVE_PI_BRIDGE="+driver, "CHR_LIVE_PI_IMAGE=fixture-runtime", "CHR_LIVE_PI_AUTH="+filepath.Join(root, "missing-auth.json"))

	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "live credential source must be a private regular file") {
		t.Fatal("missing live access was skipped or reported as success", err)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("missing live access still launched a model")
	}
}
