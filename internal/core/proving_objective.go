package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProvingObjective describes a package boundary value.
type ProvingObjective struct {
	ID             string
	TicketKey      string
	Kind           string
	ArtifactRoot   string
	ContextPath    string
	ResultPath     string
	LogPath        string
	ValidationPath string
	MaxAttempts    int
	Body           string
}

// LoadProvingObjective loads the corresponding package data.
func LoadProvingObjective(path string) (ProvingObjective, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ProvingObjective{}, fmt.Errorf("read proving objective: %w", err)
	}

	objective, err := ParseProvingObjective(string(content))
	if err != nil {
		return ProvingObjective{}, fmt.Errorf("parse proving objective %s: %w", path, err)
	}

	return objective, nil
}

// ParseProvingObjective parses the corresponding package data.
func ParseProvingObjective(content string) (ProvingObjective, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return ProvingObjective{}, fmt.Errorf("objective frontmatter is required")
	}

	rest := strings.TrimPrefix(normalized, "---\n")

	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) != 2 {
		return ProvingObjective{}, fmt.Errorf("objective frontmatter terminator is required")
	}

	fields := parseSimpleFrontmatter(parts[0])

	maxAttempts, err := strconv.Atoi(fields["max_attempts"])
	if err != nil || maxAttempts <= 0 {
		return ProvingObjective{}, fmt.Errorf("max_attempts must be a positive integer")
	}

	objective := ProvingObjective{
		ID:             fields["objective_id"],
		TicketKey:      fields["ticket_key"],
		Kind:           fields["kind"],
		ArtifactRoot:   fields["artifact_root"],
		ContextPath:    fields["context_path"],
		ResultPath:     fields["result_path"],
		LogPath:        fields["log_path"],
		ValidationPath: fields["validation_path"],
		MaxAttempts:    maxAttempts,
		Body:           strings.TrimSpace(parts[1]),
	}

	err = validateProvingObjective(objective)
	if err != nil {
		return ProvingObjective{}, err
	}

	return objective, nil
}

type provingObjectiveSource struct {
	ProjectDir string
	Objective  ProvingObjective
}

func loadProvingObjectiveForProject(projectDir, objectivePath string) (provingObjectiveSource, error) {
	projectDir, err := resolveProjectDir(projectDir)
	if err != nil {
		return provingObjectiveSource{}, err
	}

	objective, err := LoadProvingObjective(resolveProjectPath(projectDir, objectivePath))
	if err != nil {
		return provingObjectiveSource{}, err
	}

	return provingObjectiveSource{ProjectDir: projectDir, Objective: objective}, nil
}

func validateProvingObjective(objective ProvingObjective) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "objective_id", value: objective.ID},
		{name: "ticket_key", value: objective.TicketKey},
		{name: "kind", value: objective.Kind},
		{name: "artifact_root", value: objective.ArtifactRoot},
		{name: "context_path", value: objective.ContextPath},
		{name: "result_path", value: objective.ResultPath},
		{name: "log_path", value: objective.LogPath},
		{name: "validation_path", value: objective.ValidationPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}

	if strings.TrimSpace(objective.Body) == "" {
		return fmt.Errorf("objective body is required")
	}

	return nil
}

func findProvingTicketPath(projectDir, ticketKey string) (string, error) {
	ticketsDir := UserPath(projectDir, "tickets")

	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return "", fmt.Errorf("read proving tickets directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), ticketKey+"-") && strings.HasSuffix(entry.Name(), ".md") {
			return filepath.Join(ticketsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("ticket %s not found", ticketKey)
}
