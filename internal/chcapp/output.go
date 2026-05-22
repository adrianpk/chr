package chcapp

import (
	"fmt"
	"io"
	"path/filepath"
)

func displayPath(projectDir, path string) string {
	if projectDir == "" || path == "" {
		return path
	}

	relative, err := filepath.Rel(projectDir, path)
	if err != nil {
		return path
	}

	if relative == "." || relative == "" {
		return "."
	}

	if relative == ".." || startsWithParent(relative) {
		return path
	}

	return relative
}

func startsWithParent(path string) bool {
	return len(path) >= 3 && path[:3] == "../"
}

func writeLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return err
		}
	}

	return nil
}
