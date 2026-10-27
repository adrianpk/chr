package config

import "strings"

// SetValue replaces or adds one root-section scalar while preserving the other configuration lines.
func SetValue(data []byte, section, key, value string) []byte {
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	start, end := -1, len(lines)

	for i, line := range lines {
		if line == section+":" {
			start = i

			continue
		}

		if start >= 0 && line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			end = i

			break
		}
	}

	replacement := "  " + key + ": " + value
	if start < 0 {
		lines = append(lines, section+":", replacement)
	} else {
		lines = replaceWorkflowUpgradeValue(lines, start, end, key, replacement)
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}

func replaceWorkflowUpgradeValue(lines []string, start, end int, key, replacement string) []string {
	for i := start + 1; i < end; i++ {
		observed := parseWorkspaceConfigLine(lines[i], "")
		if observed.Parsed && observed.Key == key {
			lines[i] = replacement

			return lines
		}
	}

	return append(lines[:end], append([]string{replacement}, lines[end:]...)...)
}
