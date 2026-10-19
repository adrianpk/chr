package pi

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"forge.adrianpk.com/chr/internal/harness"
)

//go:embed readonly.mjs
var readOnlyExtension []byte

func writeReadOnlyProfile(dir string, request harness.AgentRunRequest) error {
	root, err := filepath.EvalSymlinks(request.ProjectDir)
	if err != nil || !filepath.IsAbs(root) {
		return errors.New("pi file inspection requires an existing absolute project root")
	}

	blocked := []string{request.Resources.CredentialFile, request.Resources.ModelConfigFile}
	for i, path := range blocked {
		if path != "" {
			blocked[i], err = filepath.EvalSymlinks(path)
			if err != nil {
				return errors.New("pi inspection resource boundary is unavailable")
			}
		}
	}

	refs, err := inspectionReferences(request)
	if err != nil {
		return err
	}

	config, err := json.Marshal(map[string]any{"root": root, "blocked": blocked, "allowed": refs})
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "inspection.json"), config, 0o600); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "readonly.mjs"), readOnlyExtension, 0o600)
}

func inspectionReferences(request harness.AgentRunRequest) ([]string, error) {
	if len(request.ReadOnlyRefs) > 16 {
		return nil, errors.New("pi inspection accepts at most 16 explicit evidence files")
	}

	refs := make([]string, 0, len(request.ReadOnlyRefs))
	for _, ref := range request.ReadOnlyRefs {
		if ref == "" {
			continue
		}

		if len(ref) > 4096 {
			return nil, errors.New("pi inspection evidence path exceeds its byte ceiling")
		}

		if !filepath.IsAbs(ref) {
			ref = filepath.Join(request.ProjectDir, ref)
		}

		real, err := filepath.EvalSymlinks(ref)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, errors.New("pi inspection evidence reference is unavailable")
		}

		info, err := os.Stat(real)
		if err != nil || !info.Mode().IsRegular() {
			return nil, errors.New("pi inspection evidence must reference a regular file")
		}

		refs = append(refs, real)
	}

	return refs, nil
}

func writeCallArtifact(path string, data []byte) (err error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
	if err != nil {
		return errors.New("pi call artifact cannot be opened privately")
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	info, err := file.Stat()
	if err != nil || !privateArtifact(info) {
		return errors.New("pi call artifact requires a private regular file with one link")
	}

	if err := file.Truncate(0); err != nil {
		return errors.New("pi call artifact cannot be replaced")
	}

	_, err = file.Write(data)

	return err
}

func privateArtifact(info os.FileInfo) bool {
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)

	return ok && stat.Nlink == 1 && stat.Uid == uint32(os.Geteuid())
}

func checkCallDirectoryPath(dir string) error {
	if !filepath.IsAbs(dir) || len(dir) > 4096 {
		return errors.New("pi run directory requires a bounded absolute path")
	}

	for current := filepath.Clean(dir); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return errors.New("pi run directory is unavailable")
		}

		if err == nil && !info.IsDir() {
			return errors.New("pi run directory cannot traverse links or non-directories")
		}

		if filepath.Dir(current) == current {
			return nil
		}
	}
}
