package app

import (
	"debug/buildinfo"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
)

// Inspect executable metadata without running a configured program. A worker
// profile selects the static Linux CHR bridge; generation selects native Pi.
// Image and executable hashes are frozen separately at admission.
func piWorkerBridgePreparation(program string) (chr, prepared bool) {
	info := piBridgeBuildInfo(program)
	if info == nil || info.Path != "forge.adrianpk.com/chr/cmd/chr" {
		return false, false
	}

	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}

	return true, settings["GOOS"] == "linux" && settings["GOARCH"] == runtime.GOARCH && settings["CGO_ENABLED"] == "0"
}

func piBridgeBuildInfo(program string) *debug.BuildInfo {
	path, err := exec.LookPath(program)
	if err != nil {
		return nil
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return nil
	}

	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }() //nolint:errcheck // Read-only metadata inspection.

	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() > 512<<20 {
		return nil
	}

	info, err := buildinfo.Read(file)
	if err != nil {
		return nil
	}

	return info
}
