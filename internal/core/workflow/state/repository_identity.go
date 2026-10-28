package state

import (
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

func validRepositoryPart(part string) bool {
	return part != "" && part != "." && part != ".." && len(part) <= 128 && !strings.ContainsAny(part, " \\%\x00\r\n\t")
}

func validRepositoryURL(parsed *url.URL) bool {
	return parsed.Scheme == "https" && parsed.User == nil && parsed.RawQuery == "" && !parsed.ForceQuery && parsed.Fragment == "" && parsed.RawPath == "" && parsed.Host != ""
}

// CanonicalRepository keeps web identity independent of host transport aliases and access
// credentials. Preserve path case because the forge owns repository semantics.
func CanonicalRepository(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || !validRepositoryURL(parsed) {
		return "", fmt.Errorf("canonical repository must be a public HTTPS repository URL")
	}

	parsed.Path = strings.TrimSuffix(strings.TrimSuffix(parsed.Path, "/"), ".git")

	parts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
	if len(parts) != 2 || !validRepositoryPart(parts[0]) || !validRepositoryPart(parts[1]) {
		return "", ErrConflict
	}

	parsed.Host, err = canonicalRepositoryHost(parsed)
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}

func canonicalRepositoryHost(parsed *url.URL) (string, error) {
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || strings.ContainsAny(host, " \\%\x00\r\n\t") || strings.HasSuffix(parsed.Host, ":") {
		return "", ErrConflict
	}

	if strings.Contains(host, ":") {
		ip, err := netip.ParseAddr(host)
		if err != nil {
			return "", ErrConflict
		}

		host = "[" + ip.String() + "]"
	}

	port, err := canonicalRepositoryPort(parsed.Port())
	if err != nil {
		return "", err
	}

	return host + port, nil
}

func canonicalRepositoryPort(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return "", ErrConflict
	}

	if port == 443 {
		return "", nil
	}

	return ":" + strconv.Itoa(port), nil
}
