package chc

import "path/filepath"

func DumpDirectory(projectDir string) string {
	return filepath.Join(projectDir, DirectoryName, "dumps")
}
