package closeout

import (
	"fmt"
	"path/filepath"
)

// File names evidence within one retained closing attempt.
func File(id string, generation int, name string) string {
	return filepath.ToSlash(filepath.Join("plan", id, "extension", "chr", "closing", fmt.Sprintf("%04d", generation), name))
}
