package forge

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Provider identifies the supported forge protocol.
type Provider string

const (
	// Forgejo selects the Forgejo API protocol.
	Forgejo Provider = "forgejo"
	// GitHub selects the GitHub API protocol.
	GitHub Provider = "github"
	// Unsupported retains an unrecognized reference without granting provider IO.
	Unsupported Provider = "unsupported"
)

// Reference retains the parsed provider, canonical URL and PR identity.
type Reference struct {
	Provider Provider
	URL      string
	Host     string
	Owner    string
	Repo     string
	Number   int
}

// ParseReference recognizes existing provider URL forms and preserves partial facts on error.
func ParseReference(rawURL string) (Reference, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return Reference{Provider: Unsupported}, errors.New("pull request URL is required")
	}

	address, err := parseReviewProviderURL(trimmed)

	reference := Reference{
		Provider: Unsupported,
		URL:      trimmed,
	}
	if err != nil {
		return reference, err
	}

	reference.Host = strings.ToLower(address.url.Hostname())
	reference.Owner = address.segments[0]
	reference.Repo = address.segments[1]

	reference.Number, err = parseReviewProviderPRNumber(address.segments[3], trimmed)
	if err != nil {
		return reference, err
	}

	reference.Provider = detectReviewProvider(reference.Host, address.segments[2])
	if reference.Provider == Unsupported {
		return reference, fmt.Errorf("unsupported pull request URL: %s", trimmed)
	}

	return reference, nil
}

func cleanPathSegments(rawPath string) []string {
	parts := strings.Split(strings.Trim(rawPath, "/"), "/")
	segments := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}

	return segments
}

type reviewProviderURL struct {
	url      *url.URL
	segments []string
}

func parseReviewProviderURL(rawURL string) (reviewProviderURL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return reviewProviderURL{}, fmt.Errorf("invalid pull request URL: %s", rawURL)
	}

	segments := cleanPathSegments(parsed.Path)
	if len(segments) < 4 {
		return reviewProviderURL{}, fmt.Errorf("unsupported pull request URL: %s", rawURL)
	}

	return reviewProviderURL{url: parsed, segments: segments}, nil
}

func parseReviewProviderPRNumber(rawNumber, rawURL string) (int, error) {
	number, err := strconv.Atoi(rawNumber)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("unsupported pull request URL: %s", rawURL)
	}

	return number, nil
}

func detectReviewProvider(host, pathMarker string) Provider {
	if strings.EqualFold(host, "github.com") && pathMarker == "pull" {
		return GitHub
	}

	if pathMarker == "pulls" {
		return Forgejo
	}

	return Unsupported
}
