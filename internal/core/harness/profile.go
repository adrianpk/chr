package harness

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

const profileLimit = 65_536

// SelectionConfig stores named references, never executable or credential paths.
type SelectionConfig struct {
	Worker     string `json:"worker,omitempty"`
	Generation string `json:"generation,omitempty"`
}

// Profile returns the selected reference for one recognized role.
func (c SelectionConfig) Profile(role Role) (string, error) {
	switch role {
	case Worker:
		return c.Worker, nil
	case Generation:
		return c.Generation, nil
	default:
		return "", errors.New("unsupported harness role")
	}
}

// Profile is private host configuration; projections omit local references.
type Profile struct {
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

// LoadProfiles reads a bounded private file without discovering resources.
// An absent file is valid for installations that use only legacy dispatch.
func LoadProfiles(configDir string) (map[string]Profile, error) {
	path := filepath.Join(configDir, "harnesses.yml")

	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]Profile{}, nil
	}

	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > profileLimit {
		return nil, errors.New("harnesses.yml must be a private regular file of at most 64 KiB")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("private harness configuration is unavailable")
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // Read-only; the bounded reader reports IO failures.

	data, err := readProfileBytes(file)
	if err != nil {
		return nil, err
	}

	return parseProfiles(data)
}

// ProfileRevision hashes semantic references without reading credentials.
func ProfileRevision(p Profile) string {
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

func validateProfile(p Profile) error {
	if !ValidProfileName(p.Name) || !slices.Contains([]string{"codex", "pi"}, p.Harness) || p.Command == "" {
		return errors.New("a harness profile needs a unique name, supported harness and executable")
	}

	if p.Harness == "pi" && (p.ModelProvider == "" || p.Model == "" || p.Model == "cli_default") {
		return errors.New("a Pi profile requires an exact provider and model")
	}

	return nil
}

func profileScalar(value string) (string, error) {
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
