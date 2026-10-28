package workflow

import "testing"

// TestNativeRepositoryAliases resolves real Git remote names into one public
// identity. Credentials and private transport aliases never become queue keys;
// another port, path case or repository retains its distinct explicit identity.
func TestNativeRepositoryAliases(t *testing.T) {
	store, _, _ := workflowClaimFixture(t)
	for _, tc := range []struct {
		name, raw, want string
	}{
		{"canonical", "https://forge.example/test/repo", "https://forge.example/test/repo"},
		{"suffix", "https://forge.example/test/repo.git/", "https://forge.example/test/repo"},
		{"host-case", "https://FORGE.EXAMPLE/test/repo", "https://forge.example/test/repo"},
		{"default-port", "https://forge.example:443/test/repo", "https://forge.example/test/repo"},
		{"port-spelling", "https://forge.example:00443/test/repo", "https://forge.example/test/repo"},
		{"dns-root", "https://forge.example.:443/test/repo", "https://forge.example/test/repo"},
		{"other-port", "https://forge.example:8443/test/repo", "https://forge.example:8443/test/repo"},
		{"path-case", "https://forge.example/Test/Repo", "https://forge.example/Test/Repo"},
		{"ipv6", "https://[2001:db8:0::1]:443/test/repo", "https://[2001:db8::1]/test/repo"},
		{"ipv6-mapped", "https://[::ffff:192.0.2.1]:443/test/repo", "https://[::ffff:192.0.2.1]/test/repo"},
		{"userinfo", "https://worker@forge.example/test/repo", ""},
		{"ssh", "git@forge.example:test/repo.git", ""},
		{"http", "http://forge.example/test/repo", ""},
		{"empty-port", "https://forge.example:/test/repo", ""},
		{"zero-port", "https://forge.example:0/test/repo", ""},
		{"port-overflow", "https://forge.example:65536/test/repo", ""},
		{"encoded-path", "https://forge.example/test/%72epo", ""},
		{"query", "https://forge.example/test/repo?branch=dev", ""},
		{"empty-query", "https://forge.example/test/repo?", ""},
		{"fragment", "https://forge.example/test/repo#dev", ""},
		{"nested-path", "https://forge.example/test/group/repo", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workflowSelectionGit(t, store, "remote", "add", tc.name, tc.raw)

			got, err := EntryRepository(t.Context(), store.layout.ProjectDir(), tc.name)
			if (err != nil) != (tc.want == "") || got != tc.want {
				t.Fatalf("remote identity = %q, %v", got, err)
			}

			if got != "" {
				again, err := workflowCanonicalRepository(got)
				if err != nil || again != got {
					t.Fatal("canonical repository identity was not stable", err)
				}
			}
		})
	}
}
