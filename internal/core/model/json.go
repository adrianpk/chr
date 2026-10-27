package model

import "strings"

// ExtractJSONObject retains the original model-output object extraction rule.
func ExtractJSONObject(output []byte) []byte {
	text := strings.TrimSpace(string(output))
	start := strings.Index(text, "{")

	end := strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return output
	}

	return []byte(text[start : end+1])
}
