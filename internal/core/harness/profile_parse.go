package harness

import (
	"errors"
	"io"
	"strings"
)

func readProfileBytes(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, profileLimit+1))
	if err != nil || len(data) > profileLimit {
		return nil, errors.New("private harness configuration exceeds its readable bounds")
	}

	return data, nil
}

type profileParser struct {
	profiles map[string]Profile
	current  string
	fields   map[string]bool
	root     bool
}

func parseProfiles(data []byte) (map[string]Profile, error) {
	if len(data) > profileLimit {
		return nil, errors.New("private harness configuration exceeds 64 KiB")
	}

	parser := profileParser{profiles: map[string]Profile{}}
	for _, line := range strings.Split(string(data), "\n") {
		if err := parser.line(line); err != nil {
			return nil, err
		}
	}

	if !parser.root {
		return nil, errors.New("harness configuration requires a profiles mapping")
	}

	for name, profile := range parser.profiles {
		if err := validateProfile(profile); err != nil {
			return nil, err
		}

		profile.Revision = ProfileRevision(profile)
		parser.profiles[name] = profile
	}

	return parser.profiles, nil
}

func (p *profileParser) line(raw string) error {
	line := strings.TrimSpace(raw)
	if harnessProfileComment(line) {
		return nil
	}

	if strings.ContainsAny(raw, "\t\r") {
		return errors.New("harness configuration requires space indentation")
	}

	if !p.root && raw == "profiles:" {
		p.root = true

		return nil
	}

	indent := len(raw) - len(strings.TrimLeft(raw, " "))

	if !p.root {
		return errors.New("harness configuration requires a profiles mapping")
	}

	if indent == 2 && strings.HasSuffix(line, ":") {
		return p.profile(strings.TrimSuffix(line, ":"))
	}

	if indent != 4 || p.current == "" {
		return errors.New("harness profile fields require four-space indentation")
	}

	return p.field(line)
}

func (p *profileParser) profile(name string) error {
	if !ValidProfileName(name) || len(p.profiles) >= 100 {
		return errors.New("harness profiles require names of 1..128 plain bytes and at most 100 entries")
	}

	if _, found := p.profiles[name]; found {
		return errors.New("duplicate harness profile")
	}

	p.current, p.fields = name, map[string]bool{}
	p.profiles[name] = Profile{Name: name}

	return nil
}

func (p *profileParser) field(line string) error {
	key, raw, ok := strings.Cut(line, ":")
	if !ok || p.fields[key] {
		return errors.New("duplicate or malformed harness profile field")
	}

	value, err := profileScalar(raw)
	if err != nil {
		return err
	}

	profile := p.profiles[p.current]
	fields := map[string]*string{
		"harness": &profile.Harness, "command": &profile.Command,
		"model_provider": &profile.ModelProvider, "model": &profile.Model,
		"thinking": &profile.Thinking, "credential_ref": &profile.CredentialRef,
		"model_config_ref": &profile.ModelConfigRef, "runtime_ref": &profile.RuntimeRef,
	}

	target, found := fields[key]
	if !found {
		return errors.New("unsupported harness profile field")
	}

	*target = value
	p.fields[key], p.profiles[p.current] = true, profile

	return nil
}

func harnessProfileComment(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}
