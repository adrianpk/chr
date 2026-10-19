package pi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

func newWorkerPrivateDirectory(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	return dir
}

func writeWorkerAuth(t *testing.T, path, access string) {
	t.Helper()

	data := fmt.Sprintf(`{"fixture":{"type":"oauth","access":%q,"refresh":"literal-refresh","expires":1}}`, access)
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readWorkerAccess(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var document map[string]struct{ Access string }
	if err := json.Unmarshal(data, &document); err != nil {
		return "", err
	}

	return document["fixture"].Access, nil
}

// Concurrent callers and an independent host process share the same exclusion
// file. Each gets a private copy and observes the previous atomic refresh.
// Renewal is simulated at the provider IO boundary; no OAuth account is used.
func TestPiCredentialRefresh(t *testing.T) {
	source := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
	writeWorkerAuth(t, source, "0")

	var group sync.WaitGroup

	failures := make(chan error, 9)

	for range 8 {
		private := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")

		group.Go(func() {
			failures <- incrementWorkerCredential(t.Context(), source, private)
		})
	}

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestPiCredentialRefreshChild$")
	command.Env = append(os.Environ(), "CHR_REFRESH_SOURCE="+source, "CHR_REFRESH_PRIVATE="+filepath.Join(newWorkerPrivateDirectory(t), "auth.json"))

	group.Go(func() {
		out, err := command.CombinedOutput()
		if err != nil {
			failures <- fmt.Errorf("refresh child: %w: %s", err, out)
		}
	})
	group.Wait()
	close(failures)

	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}

	access, err := readWorkerAccess(source)
	if err != nil || access != "9" {
		t.Fatalf("lost refresh update: %q, %v", access, err)
	}
}

// A subprocess owns its own file descriptors and must coordinate with the host.
func TestPiCredentialRefreshChild(t *testing.T) {
	source := os.Getenv("CHR_REFRESH_SOURCE")
	if source == "" {
		t.Skip("independent refresh-owner fixture")
	}

	if err := incrementWorkerCredential(t.Context(), source, os.Getenv("CHR_REFRESH_PRIVATE")); err != nil {
		t.Fatal(err)
	}
}

func incrementWorkerCredential(ctx context.Context, source, private string) error {
	return WithWorkerCredential(ctx, source, private, "fixture", func() error {
		access, err := readWorkerAccess(private)
		if err != nil {
			return err
		}

		current, err := strconv.Atoi(access)
		if err != nil {
			return err
		}

		data := fmt.Sprintf(`{"fixture":{"type":"oauth","access":%q,"refresh":"literal-refresh","expires":1}}`, strconv.Itoa(current+1))

		return os.WriteFile(private, []byte(data), 0o600)
	})
}

// A failed model call may already have rotated a refresh token. Publish the
// valid credential update, preserve the call failure and cancel lock waiters.
func TestPiCredentialRefreshFailure(t *testing.T) {
	source := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
	private := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
	writeWorkerAuth(t, source, "old")

	callErr := errors.New("model call failed after refresh")

	err := WithWorkerCredential(t.Context(), source, private, "fixture", func() error {
		writeWorkerAuth(t, private, "new")

		return callErr
	})
	if !errors.Is(err, callErr) {
		t.Fatalf("lost call failure: %v", err)
	}

	access, err := readWorkerAccess(source)
	if err != nil || access != "new" {
		t.Fatalf("lost refreshed credential: %q, %v", access, err)
	}

	lock, err := acquireRefreshLock(t.Context(), filepath.Join(filepath.Dir(source), refreshLockName))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }() //nolint:errcheck // Fixture cleanup.

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = acquireRefreshLock(ctx, filepath.Join(filepath.Dir(source), refreshLockName))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lock ignored cancellation: %v", err)
	}
}

// Malformed refreshed credentials or an operator's uncoordinated replacement
// must not overwrite the canonical credential. A shared directory is rejected.
func TestPiCredentialRefreshReject(t *testing.T) {
	for _, scenario := range []string{"malformed", "operator replacement", "unrelated file"} {
		t.Run(scenario, func(t *testing.T) {
			source := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
			private := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
			writeWorkerAuth(t, source, "original")

			if scenario == "unrelated file" {
				writeWorkerAuth(t, filepath.Join(filepath.Dir(source), "other.json"), "unrelated")
			}

			err := WithWorkerCredential(t.Context(), source, private, "fixture", func() error {
				return replaceWorkerCredentialFixture(t, scenario, source, private)
			})
			if err == nil {
				t.Fatal("unsafe refresh accepted")
			}

			access, err := readWorkerAccess(source)

			want := "original"
			if scenario == "operator replacement" {
				want = "operator"
			}

			if err != nil || access != want {
				t.Fatalf("source changed: %q, %v", access, err)
			}
		})
	}
}

// The owner must bound lock waiting by its caller's deadline, independently of
// the provider's refresh duration; there is no refresh retry budget here.
func TestPiCredentialRefreshDeadline(t *testing.T) {
	lockPath := filepath.Join(newWorkerPrivateDirectory(t), refreshLockName)

	lock, err := acquireRefreshLock(t.Context(), lockPath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = lock.Close() }() //nolint:errcheck // Fixture cleanup.

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()

	if _, err := acquireRefreshLock(ctx, lockPath); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("refresh wait ignored deadline: %v", err)
	}
}

func replaceWorkerCredentialFixture(t *testing.T, scenario, source, private string) error {
	t.Helper()

	if scenario == "operator replacement" {
		writeWorkerAuth(t, source, "operator")
		writeWorkerAuth(t, private, "refreshed")

		return nil
	}

	return os.WriteFile(private, []byte("{}"), 0o600)
}

// Workers import only one explicit provider and require a refreshable private
// OAuth owner; native generation's static-only resource policy stays separate.
func TestPiCredentialRefreshResources(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		data      string
		wantOK    bool
		wantOAuth bool
	}{
		{"static", `{"fixture":{"type":"api_key","key":"literal-key"}}`, true, false},
		{"OAuth", `{"fixture":{"type":"oauth","access":"literal-access","refresh":"literal-refresh","expires":1,"accountId":"fixture-account"}}`, true, true},
		{"other provider", `{"other":{"type":"api_key","key":"unselected"}}`, false, false},
		{"multiple providers", `{"fixture":{"type":"api_key","key":"literal-key"},"other":{"type":"api_key","key":"unselected"}}`, false, false},
		{"credential command", `{"fixture":{"type":"api_key","key":"!command"}}`, false, false},
		{"credential environment", `{"fixture":{"type":"oauth","access":"$TOKEN","refresh":"literal-refresh","expires":1}}`, false, false},
		{"missing refresh", `{"fixture":{"type":"oauth","access":"literal-access","expires":1}}`, false, false},
		{"nested metadata", `{"fixture":{"type":"oauth","access":"literal-access","refresh":"literal-refresh","expires":1,"env":{"KEY":"value"}}}`, false, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			path := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
			if err := os.WriteFile(path, []byte(scenario.data), 0o600); err != nil {
				t.Fatal(err)
			}

			resources, err := ResolveWorkerResources("", path, "", "fixture")
			if (err == nil) != scenario.wantOK || resources.OAuth != scenario.wantOAuth {
				t.Fatalf("resources = %#v, %v", resources, err)
			}

			if scenario.wantOAuth {
				if _, err := ResolveResources("", path, "", "fixture"); !errors.Is(err, ErrCapability) {
					t.Fatalf("generation admitted OAuth without its owner: %v", err)
				}
			}
		})
	}
}

// Already canceled calls must not create private credentials or acquire a new
// refresh owner, even when no other process currently holds exclusion.
func TestPiCredentialRefreshCanceled(t *testing.T) {
	source := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
	private := filepath.Join(newWorkerPrivateDirectory(t), "auth.json")
	writeWorkerAuth(t, source, "original")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := WithWorkerCredential(ctx, source, private, "fixture", func() error {
		t.Fatal("canceled owner started a call")

		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored cancellation: %v", err)
	}

	if _, err := os.Stat(private); !os.IsNotExist(err) {
		t.Fatalf("canceled call created private credentials: %v", err)
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(source), refreshLockName)); !os.IsNotExist(err) {
		t.Fatalf("canceled call created a refresh lock: %v", err)
	}
}
