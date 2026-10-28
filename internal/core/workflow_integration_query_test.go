package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func nativeQueueImages(t *testing.T, root string) map[string]string {
	t.Helper()

	images := map[string]string{}

	err := filepath.WalkDir(root, func(path string, file os.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		if err != nil {
			return err
		}

		if file.Type().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			images[path] = contenthash.Sum(data)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return images
}
