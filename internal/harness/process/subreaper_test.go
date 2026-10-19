package process

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Adoption is enabled only in an owned subprocess. A detached tool descendant
// must be killed and waited before its bridge can release a capacity slot.
func TestAgentRunProcessReaper(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux worker runtime")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAgentRunProcessReaperDriver$")

	command.Env = append(os.Environ(), "CHR_REAPER_FIXTURE=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("owned child reaping: %v: %s", err, output)
	}
}

// This isolated helper never enables subreaping in the package test host.
func TestAgentRunProcessReaperDriver(t *testing.T) {
	if os.Getenv("CHR_REAPER_FIXTURE") != "1" {
		t.Skip("owned child fixture")
	}

	if err := EnableChildReaping(); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("sh", "-c", "sleep 30 >/dev/null 2>&1 & printf '%s' $!")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = syscall.Kill(pid, syscall.SIGKILL) }() //nolint:errcheck // Failure cleanup in this owned process.

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	if err := ReapChildren(ctx); err != nil {
		t.Fatal(err)
	}

	if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
		t.Fatalf("detached tool child survived reaping: %v", err)
	}
}
