package workspace

import "path/filepath"

const (
	// UserDirectoryName is the default pseudo-user workspace directory name.
	UserDirectoryName = "default"
)

// UserRelativePath executes the corresponding package operation.
func UserRelativePath(elements ...string) string {
	parts := append([]string{DirectoryName, UserDirectoryName}, elements...)

	return filepath.ToSlash(filepath.Join(parts...))
}

// UserPath executes the corresponding package operation.
func UserPath(projectDir string, elements ...string) string {
	return filepath.Join(projectDir, filepath.FromSlash(UserRelativePath(elements...)))
}
