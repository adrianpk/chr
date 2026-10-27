package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

func projectMetadataStringFields(fields map[string]json.RawMessage) map[string]string {
	values := make(map[string]string, len(fields))

	for key, value := range fields {
		values[key] = unquoteMetadataString(string(value))
	}

	return values
}

func renderProjectMetadataFields(metadata ProjectMetadata) string {
	var lines []string
	if metadata.frontmatter != "" {
		lines = strings.Split(metadata.frontmatter, "\n")
	}

	fields := []struct{ name, value string }{
		{"schema_version", "1"}, {"name", yamlString(metadata.Name)}, {"description", yamlString(metadata.Description)},
	}
	if metadata.ProjectDumpRef != "" || strings.Contains("\n"+metadata.frontmatter, "\nproject_dump_ref:") {
		fields = append(fields, struct{ name, value string }{"project_dump_ref", yamlString(metadata.ProjectDumpRef)})
	}

	for _, field := range fields {
		lines = replaceProjectMetadataField(lines, field.name, field.value)
	}

	return strings.Join(lines, "\n")
}

func replaceProjectMetadataField(lines []string, key, value string) []string {
	for i, line := range lines {
		field := projectDocumentField(line)

		name, found := field.Name, field.Found
		if !found || name != key {
			continue
		}

		end := i + 1
		for end < len(lines) && (strings.HasPrefix(lines[end], " ") || strings.HasPrefix(lines[end], "\t")) {
			end++
		}

		return replaceStringLines(lines, i, end, []string{key + ": " + value})
	}

	return append(lines, key+": "+value)
}

func projectDocumentFields(frontmatter string) (map[string]string, error) {
	fields := map[string]string{}

	for _, line := range strings.Split(frontmatter, "\n") {
		field := projectDocumentField(line)

		name, value, found := field.Name, field.Value, field.Found
		if !found {
			continue
		}

		if _, exists := fields[name]; exists {
			return nil, fmt.Errorf("duplicate document frontmatter field: %s", name)
		}

		fields[name] = unquoteMetadataString(value)
	}

	return fields, nil
}

func projectDocumentField(line string) documentField {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return documentField{Name: "", Value: "", Found: false}
	}

	name, value, found := strings.Cut(line, ":")

	return documentField{Name: strings.TrimSpace(name), Value: strings.TrimSpace(value), Found: found}
}
