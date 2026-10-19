package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
)

const harnessExecutableLimit = 512 << 20

// FreezeHarnessExecutable resolves a host-selected program once and binds its
// regular executable file and argv. It never launches the executable.
func FreezeHarnessExecutable(command []string) (HarnessExecutableRef, []string, error) {
	if err := validateWorkflowCommand(command); err != nil {
		return HarnessExecutableRef{}, nil, err
	}

	program, err := exec.LookPath(command[0])
	if err != nil {
		return HarnessExecutableRef{}, nil, errors.New("selected harness executable is unavailable")
	}

	program, err = filepath.Abs(program)
	if err != nil {
		return HarnessExecutableRef{}, nil, err
	}

	program, err = filepath.EvalSymlinks(program)
	if err != nil {
		return HarnessExecutableRef{}, nil, errors.New("selected harness executable cannot be resolved")
	}

	retained := slices.Clone(command)
	retained[0] = program

	hash, err := harnessExecutableHash(program)
	if err != nil {
		return HarnessExecutableRef{}, nil, err
	}

	data, err := json.Marshal(retained)

	return HarnessExecutableRef{Hash: hash, CommandHash: workflowHash(data)}, retained, err
}

func harnessExecutableHash(program string) (string, error) {
	file, err := openHarnessExecutable(program)
	if err != nil {
		return "", errors.New("retained harness executable is unavailable")
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // Read-only; file/hash IO errors are returned.

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 || info.Size() > harnessExecutableLimit {
		return "", errors.New("harness executable requires a regular executable file within 512 MiB")
	}

	hash := sha256.New()

	count, err := io.Copy(hash, io.LimitReader(file, harnessExecutableLimit+1))
	if err != nil || count > harnessExecutableLimit {
		return "", errors.New("harness executable could not be fingerprinted within bounds")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func openHarnessExecutable(program string) (*os.File, error) {
	info, err := os.Lstat(program)
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("harness executable must be a regular file")
	}
	// Nonblocking/no-follow also protects the type check if a path is replaced
	// before open; a configured FIFO must never suspend admission.
	return os.OpenFile(program, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
}

func checkHarnessCommandImage(binding HarnessBinding, command []string) error {
	if err := validateWorkflowCommand(command); err != nil {
		return err
	}

	if !filepath.IsAbs(command[0]) || filepath.Clean(command[0]) != command[0] {
		return ErrWorkflowConflict
	}

	data, err := json.Marshal(command)
	if err != nil || workflowHash(data) != binding.ExecutableRef.CommandHash {
		return ErrWorkflowConflict
	}

	return nil
}

func checkHarnessExecutable(binding *HarnessBinding, command []string) error {
	if binding == nil {
		return nil
	}

	if err := validateWorkflowHarnessPair(binding, command); err != nil {
		return err
	}

	resolved, err := filepath.EvalSymlinks(command[0])
	if err != nil || resolved != command[0] {
		return ErrWorkflowConflict
	}

	hash, err := harnessExecutableHash(resolved)
	if err != nil {
		return err
	}

	if hash != binding.ExecutableRef.Hash {
		return errors.New("retained harness executable changed; restore the admitted executable")
	}

	return nil
}
