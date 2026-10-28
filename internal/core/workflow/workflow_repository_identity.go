package workflow

import (
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Canonical web identity is independent of host transport aliases and access
// credentials. Preserve path case because the forge owns repository semantics.
func workflowCanonicalRepository(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || !validWorkflowRepositoryURL(parsed) {
		return "", fmt.Errorf("canonical repository must be a public HTTPS repository URL")
	}

	parsed.Path = strings.TrimSuffix(strings.TrimSuffix(parsed.Path, "/"), ".git")

	parts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
	if len(parts) != 2 || !validWorkflowRepositoryPart(parts[0]) || !validWorkflowRepositoryPart(parts[1]) {
		return "", workflowstate.ErrConflict
	}

	parsed.Host, err = workflowCanonicalRepositoryHost(parsed)
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}

func workflowCanonicalRepositoryHost(parsed *url.URL) (string, error) {
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || strings.ContainsAny(host, " \\%\x00\r\n\t") || strings.HasSuffix(parsed.Host, ":") {
		return "", workflowstate.ErrConflict
	}

	if strings.Contains(host, ":") {
		ip, err := netip.ParseAddr(host)
		if err != nil {
			return "", workflowstate.ErrConflict
		}

		host = "[" + ip.String() + "]"
	}

	port, err := workflowCanonicalRepositoryPort(parsed.Port())
	if err != nil {
		return "", err
	}

	return host + port, nil
}

func workflowCanonicalRepositoryPort(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return "", workflowstate.ErrConflict
	}

	if port == 443 {
		return "", nil
	}

	return ":" + strconv.Itoa(port), nil
}

func validWorkflowRepositoryURL(parsed *url.URL) bool {
	return parsed.Scheme == "https" && parsed.User == nil && parsed.RawQuery == "" && !parsed.ForceQuery && parsed.Fragment == "" && parsed.RawPath == "" && parsed.Host != ""
}

func validWorkflowRepositoryPart(part string) bool {
	return part != "" && part != "." && part != ".." && len(part) <= 128 && !strings.ContainsAny(part, " \\%\x00\r\n\t")
}
