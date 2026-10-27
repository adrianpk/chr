package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigPath preserves the existing guarded configuration filesystem contract.
// Configuration has its own owner; workflow record safePath never gains access
// to parent directories. Every upgrade file stays inside .chr/default.
func ConfigPath(project, relative string) (string, error) {
	if !filepath.IsLocal(relative) || filepath.Clean(relative) != relative {
		return "", fmt.Errorf("invalid workflow upgrade path")
	}

	project, err := filepath.Abs(project)
	if err != nil {
		return "", err
	}

	path := UserPath(project, relative)
	for parent := path; parent != project; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil {
			if err := checkConfigPathEntry(info, parent == path); err != nil {
				return "", err
			}
		}
	}

	return path, nil
}

func checkConfigPathEntry(info os.FileInfo, leaf bool) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("workflow upgrade path cannot traverse a symlink")
	}

	if leaf && !info.Mode().IsRegular() {
		return fmt.Errorf("workflow upgrade requires regular files")
	}

	return nil
}
