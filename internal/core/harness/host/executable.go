// Package host owns executable admission and host execution policy over retained bindings.
package host

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

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const harnessExecutableLimit = 512 << 20

// FreezeHarnessExecutable resolves a host-selected program once and binds its
// regular executable file and argv. It never launches the executable.
func FreezeHarnessExecutable(command []string) (harnesscore.Executable, error) {
	if err := workflowstate.ValidateCommand(command); err != nil {
		return harnesscore.Executable{}, err
	}

	program, err := exec.LookPath(command[0])
	if err != nil {
		return harnesscore.Executable{}, errors.New("selected harness executable is unavailable")
	}

	program, err = filepath.Abs(program)
	if err != nil {
		return harnesscore.Executable{}, err
	}

	program, err = filepath.EvalSymlinks(program)
	if err != nil {
		return harnesscore.Executable{}, errors.New("selected harness executable cannot be resolved")
	}

	retained := slices.Clone(command)
	retained[0] = program

	hash, err := HarnessExecutableHash(program)
	if err != nil {
		return harnesscore.Executable{}, err
	}

	data, err := json.Marshal(retained)

	return harnesscore.Executable{Reference: harnesscore.ExecutableRef{Hash: hash, CommandHash: contenthash.Sum(data)}, Command: retained}, err
}

// HarnessExecutableHash fingerprints a bounded regular executable without launching it.
func HarnessExecutableHash(program string) (string, error) {
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

// CheckHarnessCommandImage validates the retained absolute argv image without launching it.
func CheckHarnessCommandImage(binding harnesscore.Binding, command []string) error {
	if err := workflowstate.ValidateCommand(command); err != nil {
		return err
	}

	if !filepath.IsAbs(command[0]) || filepath.Clean(command[0]) != command[0] {
		return workflowstate.ErrConflict
	}

	data, err := json.Marshal(command)
	if err != nil || contenthash.Sum(data) != binding.ExecutableRef.CommandHash {
		return workflowstate.ErrConflict
	}

	return nil
}

// CheckHarnessExecutable checks admitted command identity and the current executable image.
func CheckHarnessExecutable(binding *harnesscore.Binding, command []string) error {
	if binding == nil {
		return nil
	}

	if err := ValidateWorkflowHarnessPair(binding, command); err != nil {
		return err
	}

	resolved, err := filepath.EvalSymlinks(command[0])
	if err != nil || resolved != command[0] {
		return workflowstate.ErrConflict
	}

	hash, err := HarnessExecutableHash(resolved)
	if err != nil {
		return err
	}

	if hash != binding.ExecutableRef.Hash {
		return errors.New("retained harness executable changed; restore the admitted executable")
	}

	return nil
}
