package chc

import (
	"context"
	"errors"
	"testing"
)

func TestDetectReviewProviderReference(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		provider ReviewProvider
		host     string
		owner    string
		repo     string
		number   int
	}{
		{
			name:     "forgejo",
			rawURL:   "https://forge.adrianpk.com/adrianpk/chc/pulls/275",
			provider: ReviewProviderForgejo,
			host:     "forge.adrianpk.com",
			owner:    "adrianpk",
			repo:     "chc",
			number:   275,
		},
		{
			name:     "codeberg",
			rawURL:   "https://codeberg.org/example/repo/pulls/12",
			provider: ReviewProviderForgejo,
			host:     "codeberg.org",
			owner:    "example",
			repo:     "repo",
			number:   12,
		},
		{
			name:     "github",
			rawURL:   "https://github.com/example/repo/pull/34",
			provider: ReviewProviderGitHub,
			host:     "github.com",
			owner:    "example",
			repo:     "repo",
			number:   34,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reference, err := DetectReviewProviderReference(test.rawURL)
			if err != nil {
				t.Fatalf("DetectReviewProviderReference returned error: %v", err)
			}

			assertReviewProviderReference(t, reference, test.provider, test.host, test.owner, test.repo, test.number)
		})
	}
}

type reviewProviderReferenceAssertion struct {
	provider ReviewProvider
	host     string
	owner    string
	repo     string
	number   int
}

func assertReviewProviderReference(
	t *testing.T,
	reference ReviewProviderReference,
	provider ReviewProvider,
	host string,
	owner string,
	repo string,
	number int,
) {
	t.Helper()

	got := reviewProviderReferenceAssertion{
		provider: reference.Provider,
		host:     reference.Host,
		owner:    reference.Owner,
		repo:     reference.Repo,
		number:   reference.Number,
	}

	want := reviewProviderReferenceAssertion{
		provider: provider,
		host:     host,
		owner:    owner,
		repo:     repo,
		number:   number,
	}
	if got != want {
		t.Fatalf("reference = %#v", reference)
	}
}

func TestRunReviewProviderDiagnosticsReady(t *testing.T) {
	runner := fakeReviewProviderDiagnosticRunner{
		paths: map[string]string{"tea": "/usr/bin/tea"},
		env:   map[string]string{"FORGEJO_TOKEN": "secret"},
	}

	diagnostic := RunReviewProviderDiagnostics(t.Context(), ReviewProviderDiagnosticOptions{
		Config: DefaultReviewProviderConfig(),
		URL:    "https://forge.adrianpk.com/adrianpk/chc/pulls/275",
		Runner: runner,
	})

	if diagnostic.Status != ReviewProviderDiagnosticStatusReady {
		t.Fatalf("Status = %q, want %q", diagnostic.Status, ReviewProviderDiagnosticStatusReady)
	}

	assertReviewProviderCheck(t, diagnostic, "provider_url", ReviewProviderDiagnosticCheckOK, "detected")
	assertReviewProviderCheck(t, diagnostic, "cli_available", ReviewProviderDiagnosticCheckOK, "found")
	assertReviewProviderCheck(t, diagnostic, "token_available", ReviewProviderDiagnosticCheckOK, "found")
	assertReviewProviderCheck(t, diagnostic, "pr_readable", ReviewProviderDiagnosticCheckOK, "readable")
}

func TestRunReviewProviderDiagnosticsReportsMissingCLI(t *testing.T) {
	diagnostic := RunReviewProviderDiagnostics(t.Context(), ReviewProviderDiagnosticOptions{
		Config: DefaultReviewProviderConfig(),
		URL:    "https://github.com/example/repo/pull/34",
		Runner: fakeReviewProviderDiagnosticRunner{},
	})

	if diagnostic.Status != ReviewProviderDiagnosticStatusBlocked {
		t.Fatalf("Status = %q, want %q", diagnostic.Status, ReviewProviderDiagnosticStatusBlocked)
	}

	assertReviewProviderCheck(t, diagnostic, "cli_available", ReviewProviderDiagnosticCheckFail, "missing_cli")
}

func TestRunReviewProviderDiagnosticsReportsMissingToken(t *testing.T) {
	runner := fakeReviewProviderDiagnosticRunner{
		paths: map[string]string{"gh": "/usr/bin/gh"},
	}

	diagnostic := RunReviewProviderDiagnostics(t.Context(), ReviewProviderDiagnosticOptions{
		Config: DefaultReviewProviderConfig(),
		URL:    "https://github.com/example/repo/pull/34",
		Runner: runner,
	})

	if diagnostic.Status != ReviewProviderDiagnosticStatusBlocked {
		t.Fatalf("Status = %q, want %q", diagnostic.Status, ReviewProviderDiagnosticStatusBlocked)
	}

	assertReviewProviderCheck(t, diagnostic, "token_available", ReviewProviderDiagnosticCheckFail, "missing_token")
}

func TestRunReviewProviderDiagnosticsReportsUnsupportedProvider(t *testing.T) {
	diagnostic := RunReviewProviderDiagnostics(t.Context(), ReviewProviderDiagnosticOptions{
		Config: DefaultReviewProviderConfig(),
		URL:    "https://example.com/example/repo/merge-requests/1",
		Runner: fakeReviewProviderDiagnosticRunner{},
	})

	if diagnostic.Status != ReviewProviderDiagnosticStatusUnsupported {
		t.Fatalf("Status = %q, want %q", diagnostic.Status, ReviewProviderDiagnosticStatusUnsupported)
	}

	assertReviewProviderCheck(t, diagnostic, "provider_url", ReviewProviderDiagnosticCheckFail, "unsupported_provider")
}

func TestRunReviewProviderDiagnosticsDryRunSkipsReadability(t *testing.T) {
	runner := fakeReviewProviderDiagnosticRunner{
		paths:  map[string]string{"tea": "/usr/bin/tea"},
		env:    map[string]string{"FORGEJO_TOKEN": "secret"},
		runErr: errors.New("must not run"),
	}

	diagnostic := RunReviewProviderDiagnostics(t.Context(), ReviewProviderDiagnosticOptions{
		Config: DefaultReviewProviderConfig(),
		URL:    "https://forge.adrianpk.com/adrianpk/chc/pulls/275",
		DryRun: true,
		Runner: runner,
	})

	if diagnostic.Status != ReviewProviderDiagnosticStatusReady {
		t.Fatalf("Status = %q, want %q", diagnostic.Status, ReviewProviderDiagnosticStatusReady)
	}

	assertReviewProviderCheck(t, diagnostic, "pr_readable", ReviewProviderDiagnosticCheckSkipped, "dry_run")
}

type fakeReviewProviderDiagnosticRunner struct {
	paths  map[string]string
	env    map[string]string
	runErr error
}

func (runner fakeReviewProviderDiagnosticRunner) LookPath(file string) (string, error) {
	if path, ok := runner.paths[file]; ok {
		return path, nil
	}

	return "", errors.New("not found")
}

func (runner fakeReviewProviderDiagnosticRunner) LookupEnv(key string) string {
	return runner.env[key]
}

func (runner fakeReviewProviderDiagnosticRunner) Run(_ context.Context, _ string, _ ...string) error {
	return runner.runErr
}

func assertReviewProviderCheck(
	t *testing.T,
	diagnostic ReviewProviderDiagnostic,
	name string,
	status string,
	reason string,
) {
	t.Helper()

	for _, check := range diagnostic.Checks {
		if check.Name == name {
			if check.Status != status || check.Reason != reason {
				t.Fatalf("check %s = %#v, want status=%q reason=%q", name, check, status, reason)
			}

			return
		}
	}

	t.Fatalf("missing check %q in %#v", name, diagnostic.Checks)
}
