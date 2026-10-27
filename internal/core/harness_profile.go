package core

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/configvalue"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

const harnessProfilesLimit = 65_536

// HarnessSelectionConfig stores named references, never executable or credential paths.
type HarnessSelectionConfig struct {
	Worker     string `json:"worker,omitempty"`
	Generation string `json:"generation,omitempty"`
}

// Profile returns the selected reference for one recognized role.
func (c HarnessSelectionConfig) Profile(role harnesscore.Role) (string, error) {
	switch role {
	case harnesscore.Worker:
		return c.Worker, nil
	case harnesscore.Generation:
		return c.Generation, nil
	default:
		return "", errors.New("unsupported harness role")
	}
}

// HarnessProfile is private host configuration; projections omit local references.
type HarnessProfile struct {
	Name           string `json:"name"`
	Harness        string `json:"harness"`
	Command        string `json:"-"`
	ModelProvider  string `json:"model_provider,omitempty"`
	Model          string `json:"model,omitempty"`
	Thinking       string `json:"thinking,omitempty"`
	CredentialRef  string `json:"-"`
	ModelConfigRef string `json:"-"`
	RuntimeRef     string `json:"-"`
	Revision       string `json:"revision"`
}

// LoadHarnessProfiles reads a bounded private file without discovering resources.
// An absent file is valid for installations that use only legacy dispatch.
func LoadHarnessProfiles(configDir string) (map[string]HarnessProfile, error) {
	path := filepath.Join(configDir, "harnesses.yml")

	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]HarnessProfile{}, nil
	}

	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > harnessProfilesLimit {
		return nil, errors.New("harnesses.yml must be a private regular file of at most 64 KiB")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("private harness configuration is unavailable")
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // Read-only; the bounded reader reports IO failures.

	data, err := readHarnessProfileBytes(file)
	if err != nil {
		return nil, err
	}

	return parseHarnessProfiles(data)
}

// HarnessProfileRevision hashes semantic references without reading credentials.
func HarnessProfileRevision(p HarnessProfile) string {
	data := []byte{}
	for _, value := range []string{
		p.Name, p.Harness, p.Command, p.ModelProvider, p.Model,
		p.Thinking, p.CredentialRef, p.ModelConfigRef, p.RuntimeRef,
	} {
		data = binary.AppendUvarint(data, uint64(len(value)))
		data = append(data, value...)
	}

	return contenthash.Sum(data)
}

func validateHarnessProfile(p HarnessProfile) error {
	if !harnesscore.ValidProfileName(p.Name) || !workflowConfigChoice(p.Harness, []string{"codex", "pi"}) || p.Command == "" {
		return errors.New("a harness profile needs a unique name, supported harness and executable")
	}

	if p.Harness == "pi" && (p.ModelProvider == "" || p.Model == "" || p.Model == "cli_default") {
		return errors.New("a Pi profile requires an exact provider and model")
	}

	return nil
}

func applyHarnessSelection(config *WorkspaceConfig, values map[string]string) error {
	for _, field := range []struct {
		key    string
		target *string
	}{
		{"harnesses.worker", &config.Harnesses.Worker},
		{"harnesses.generation", &config.Harnesses.Generation},
	} {
		value, found := values[field.key]
		if !found || value == "null" {
			continue
		}

		value = configvalue.String(value)

		if !harnesscore.ValidProfileName(value) {
			return errors.New("harness selection requires a named profile or null")
		}

		*field.target = value
	}

	return nil
}

func harnessScalar(value string) (string, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, `"`) {
		var decoded string
		if json.Unmarshal([]byte(value), &decoded) != nil || len(decoded) > 4096 || strings.ContainsAny(decoded, "\x00\r\n") {
			return "", errors.New("invalid harness scalar")
		}

		return decoded, nil
	}

	if value == "" || len(value) > 4096 || strings.ContainsAny(value, "\x00\r\n\t{}[]&*!|>'\"#") {
		return "", errors.New("harness configuration requires bounded plain or quoted scalars")
	}

	return value, nil
}
