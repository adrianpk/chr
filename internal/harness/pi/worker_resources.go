package pi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// WorkerResources retains references and credential-free model semantics.
// OAuth requires a dedicated private directory containing only its auth file
// and the adapter's exclusion file. That directory is the refresh owner.
type WorkerResources struct {
	harness.GenerationResources
	OAuth bool
}

// ResolveWorkerResources performs local preparation without authenticating.
func ResolveWorkerResources(ctx context.Context, configDir, credentialRef, modelRef, provider string) (WorkerResources, error) {
	resources := harness.GenerationResources{CredentialFile: resourcePath(configDir, credentialRef)}
	if credentialRef == "" || filepath.Base(resources.CredentialFile) == refreshLockName {
		return WorkerResources{}, errors.New("configure an explicit private Pi worker credential reference")
	}

	credential, err := workerCredential(resources.CredentialFile, provider)
	if err != nil {
		return WorkerResources{}, err
	}

	if credential.oauth {
		if err := validateRefreshDirectory(ctx, resources.CredentialFile); err != nil {
			return WorkerResources{}, err
		}
	}

	if modelRef != "" {
		resources.ModelConfigFile = resourcePath(configDir, modelRef)

		models, err := selectedModels(resources.ModelConfigFile, provider)
		if err != nil {
			return WorkerResources{}, err
		}

		if _, err := workerModelNetwork(models); err != nil {
			return WorkerResources{}, err
		}

		resources.ModelConfigHash = resourceHash(models)
	}

	return WorkerResources{GenerationResources: resources, OAuth: credential.oauth}, nil
}

type workerCredentialSelection struct {
	data  []byte
	oauth bool
}

func workerCredential(path, provider string) (workerCredentialSelection, error) {
	if filepath.Base(path) == refreshLockName {
		return workerCredentialSelection{}, errors.New("pi credential filename is reserved for refresh exclusion")
	}

	data, err := readPrivateResource(path)
	if err != nil {
		return workerCredentialSelection{}, err
	}

	fields, err := workerCredentialFields(data, provider)
	if err != nil {
		return workerCredentialSelection{}, err
	}

	if fields["type"] == "api_key" {
		selected, err := selectedCredential(path, provider)

		return workerCredentialSelection{data: selected.data}, err
	}

	info, err := os.Lstat(path)
	if err != nil || !privateArtifact(info) {
		return workerCredentialSelection{}, errors.New("pi OAuth credential owner requires a private regular file with one link")
	}

	if err := validateOAuthCredential(fields); err != nil {
		return workerCredentialSelection{}, err
	}

	selected, err := json.Marshal(map[string]any{provider: fields})

	return workerCredentialSelection{data: selected, oauth: true}, err
}

func validateOAuthCredential(fields map[string]any) error {
	expires, ok := fields["expires"].(float64)
	if fields["type"] != "oauth" || !ok || expires <= 0 || expires > 9_007_199_254_740_991 {
		return errors.New("pi worker credentials require a static API key or complete OAuth credentials")
	}

	if !literalOAuthToken(fields["access"]) || !literalOAuthToken(fields["refresh"]) {
		return errors.New("pi OAuth credentials require literal access and refresh tokens")
	}

	return validateOAuthMetadata(fields)
}

func validateOAuthMetadata(fields map[string]any) error {
	// Pi providers retain additional credential metadata. Permit only literal
	// scalar values; commands, environments and nested credential sources fail.
	for name, value := range fields {
		if name == "expires" {
			continue
		}

		text, ok := value.(string)
		if !ok || strings.ContainsAny(text, "$\x00\r\n") || strings.HasPrefix(text, "!") {
			return errors.New("unsupported Pi OAuth credential metadata")
		}
	}

	return nil
}

func literalOAuthToken(value any) bool {
	text, ok := value.(string)

	return ok && text != "" && !strings.ContainsAny(text, "$\x00\r\n") && !strings.HasPrefix(text, "!")
}

const refreshLockName = ".chr-refresh.lock"

func validateRefreshDirectory(ctx context.Context, path string) error {
	dir := filepath.Dir(path)

	if err := validateRefreshDirectoryPath(dir); err != nil {
		return err
	}

	publication, err := openCredentialPublication(ctx, path, syscall.LOCK_SH)
	if err != nil {
		return err
	}
	defer func() { _ = publication.Close() }() //nolint:errcheck // Read-only publication guard.

	directory, err := os.Open(dir)
	if err != nil {
		return errors.New("pi OAuth refresh directory is unavailable")
	}

	entries, readErr := directory.ReadDir(3)

	err = errors.Join(directory.Close(), readErr)
	if (err != nil && !errors.Is(err, io.EOF)) || len(entries) > 2 {
		return errors.New("pi OAuth refresh directory is unavailable")
	}

	for _, entry := range entries {
		if entry.Name() != filepath.Base(path) && entry.Name() != refreshLockName {
			return errors.New("pi OAuth directory may contain only the selected auth file and refresh lock")
		}
	}

	return nil
}

// WithWorkerCredential excludes concurrent refreshes across processes and calls.
// Pi writes only its private copy; action must reap its process before returning.
// This owner atomically publishes refreshed credentials before releasing exclusion.
func WithWorkerCredential(ctx context.Context, source, private, provider string, action func() error) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	credential, err := workerCredential(source, provider)
	if err != nil {
		return err
	}

	if credential.oauth {
		lock, err := acquireRefreshLock(ctx, filepath.Join(filepath.Dir(source), refreshLockName))
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, lock.Close()) }()

		if err := validateRefreshDirectory(ctx, source); err != nil {
			return err
		}
	}

	original, err := workerCredential(source, provider)
	if err != nil || original.oauth != credential.oauth {
		return errors.New("pi credential type changed during call preparation")
	}

	if err := writeCallArtifact(private, original.data); err != nil {
		return errors.New("prepare private Pi worker authentication")
	}

	err = action()
	if credential.oauth {
		err = errors.Join(err, publishRefreshedCredential(ctx, source, private, provider, original.data))
	}

	return err
}

func acquireRefreshLock(ctx context.Context, path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
	if err != nil {
		return nil, errors.New("pi OAuth refresh lock is unavailable")
	}

	info, err := file.Stat()
	if err != nil || !privateArtifact(info) {
		_ = file.Close() //nolint:errcheck // Preserve the preparation error.

		return nil, errors.New("pi OAuth refresh lock must be a private regular file")
	}

	if err := waitCredentialLock(ctx, file, syscall.LOCK_EX); err != nil {
		return nil, errors.Join(err, file.Close())
	}

	return file, nil
}

func waitCredentialLock(ctx context.Context, file *os.File, mode int) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := syscall.Flock(int(file.Fd()), mode|syscall.LOCK_NB)
		if err == nil {
			return nil
		}

		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return errors.New("pi OAuth refresh exclusion is unavailable")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func publishRefreshedCredential(ctx context.Context, source, private, provider string, original []byte) error {
	updated, err := workerCredential(private, provider)
	if err != nil || !updated.oauth {
		return errors.New("pi refresh produced invalid credentials; retained credentials were preserved")
	}

	if string(updated.data) == string(original) {
		return nil
	}

	return atomicWorkerCredential(ctx, source, provider, original, updated.data)
}

func atomicWorkerCredential(ctx context.Context, path, provider string, original, data []byte) (err error) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	publication, err := openCredentialPublication(cleanup, path, syscall.LOCK_EX)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, publication.Close()) }()

	current, err := workerCredential(path, provider)
	if err != nil || string(current.data) != string(original) {
		return errors.New("pi credential owner changed during refresh; refusing to overwrite it")
	}

	return writeWorkerCredential(path, data)
}

func writeWorkerCredential(path string, data []byte) (err error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".chr-auth-*")
	if err != nil {
		return errors.New("pi refreshed credentials could not be prepared")
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, removeErr)
		}
	}()

	if _, err := file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := errors.Join(file.Sync(), file.Close()); err != nil {
		return err
	}

	if err := os.Rename(file.Name(), path); err != nil {
		return errors.New("pi refreshed credentials could not be published")
	}

	// The rename consumed the temporary path; sync its owner directory.
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}

	return errors.Join(dir.Sync(), dir.Close())
}

func openCredentialPublication(ctx context.Context, path string, mode int) (*os.File, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		publication, err := openStableCredentialPublication(ctx, path, mode)
		if err != nil {
			return nil, err
		}

		if publication.stable {
			return publication.file, nil
		}
	}
}

type credentialPublication struct {
	file   *os.File
	stable bool
}

func openStableCredentialPublication(ctx context.Context, path string, mode int) (credentialPublication, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return credentialPublication{}, errors.New("pi credential publication owner is unavailable")
	}

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return credentialPublication{}, errors.Join(errors.New("pi credential publication requires a private regular file"), file.Close())
	}

	if err := waitCredentialLock(ctx, file, mode); err != nil {
		return credentialPublication{}, errors.Join(err, file.Close())
	}

	current, err := os.Lstat(path)
	if err != nil {
		return credentialPublication{}, errors.Join(err, file.Close())
	}

	if !os.SameFile(info, current) {
		closeErr := file.Close()
		if mode == syscall.LOCK_EX {
			return credentialPublication{}, errors.Join(errors.New("pi credential owner changed during publication"), closeErr)
		}

		return credentialPublication{}, closeErr
	}

	if !privateArtifact(current) {
		return credentialPublication{}, errors.Join(errors.New("pi credential publication owner is not private"), file.Close())
	}

	return credentialPublication{file: file, stable: true}, nil
}

func validateRefreshDirectoryPath(dir string) error {
	if credentialDirectoryIsHome(dir) {
		return errors.New("pi OAuth requires a dedicated credential directory outside the developer home root")
	}

	real, err := filepath.EvalSymlinks(dir)
	if err != nil || real != dir {
		return errors.New("pi OAuth refresh directory must not traverse links")
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return errors.New("pi OAuth requires a dedicated private credential directory")
	}

	return nil
}

func workerCredentialFields(data []byte, provider string) (map[string]any, error) {
	var entries map[string]json.RawMessage
	if json.Unmarshal(data, &entries) != nil || len(entries) != 1 || len(entries[provider]) == 0 {
		return nil, errors.New("pi workers require a credential file containing only the selected provider")
	}

	var fields map[string]any
	if json.Unmarshal(entries[provider], &fields) != nil {
		return nil, ErrCapability
	}

	return fields, nil
}

func credentialDirectoryIsHome(dir string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	if dir == filepath.Clean(home) {
		return true
	}

	real, err := filepath.EvalSymlinks(home)

	return err == nil && real == dir
}
