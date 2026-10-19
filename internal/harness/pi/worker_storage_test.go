package pi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Concurrent containers use the actual pinned Pi AuthStorage.modify API on
// private copies. The host owner must serialize and preserve both synthetic
// rotations across namespaces, without network access or a real account.
func TestPiCredentialRefreshStorage(t *testing.T) {
	image := os.Getenv("CHR_TEST_PI_RUNTIME_IMAGE")
	if image == "" {
		t.Skip("set CHR_TEST_PI_RUNTIME_IMAGE to the prepared Pi image")
	}

	binary, script := prepareStorageFixture(t)

	owner, control := newWorkerPrivateDirectory(t), newWorkerPrivateDirectory(t)
	writeStorageCredential(t, filepath.Join(owner, "auth.json"))

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	done := make(chan error, 2)
	for _, id := range []string{"first", "second"} {
		startStorageContainer(t, ctx, image, binary, script, owner, control, id, done)
	}

	waitStorageFiles(t, control, "first", "second")

	if err := os.WriteFile(filepath.Join(control, "release"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}

	checkStorageCredential(t, owner)
}

func checkStorageCredential(t *testing.T, owner string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(owner, "auth.json"))
	if err != nil {
		t.Fatal(err)
	}

	var document map[string]struct {
		Access, Refresh string
		Expires         int
	}
	if json.Unmarshal(data, &document) != nil || document["fixture"].Expires != 3 || document["fixture"].Access != "fixture-access-3" || document["fixture"].Refresh != "fixture-refresh-3" {
		t.Fatal("pinned storage lost a credential rotation")
	}

	if _, err := ResolveWorkerResources(t.Context(), "", filepath.Join(owner, "auth.json"), "", "fixture"); err != nil {
		t.Fatal("pinned storage left an invalid credential owner", err)
	}
}

func prepareStorageFixture(t *testing.T) (string, string) {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "fixture.test")
	build := exec.CommandContext(t.Context(), "go", "test", "-c", "-o", binary, ".")

	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build storage fixture: %v: %s", err, output)
	}

	script, err := filepath.Abs("testdata/worker-refresh.mjs")
	if err != nil {
		t.Fatal(err)
	}

	return binary, script
}

func writeStorageCredential(t *testing.T, path string) {
	t.Helper()

	data := `{"fixture":{"type":"oauth","access":"fixture-access-1","refresh":"fixture-refresh-1","expires":1}}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func startStorageContainer(t *testing.T, ctx context.Context, image, binary, script, owner, control, id string, done chan<- error) {
	t.Helper()

	name := fmt.Sprintf("chr-pi-storage-%s-%d", id, time.Now().UnixNano())
	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	private := fmt.Sprintf("/private:rw,nosuid,nodev,mode=0700,uid=%d,gid=%d", os.Getuid(), os.Getgid())
	args := []string{
		"run", "--rm", "--pull=never", "--network=none", "--read-only", "--name", name, "--user", user, "--tmpfs", private,
		"--mount", "type=bind,src=" + binary + ",dst=/fixture.test,readonly",
		"--mount", "type=bind,src=" + script + ",dst=/fixture.mjs,readonly",
		"--mount", "type=bind,src=" + owner + ",dst=/owner",
		"--mount", "type=bind,src=" + control + ",dst=/control",
		"--env", "CHR_PI_STORAGE_WORKER=" + id, image, "/fixture.test", "-test.run=^TestPiCredentialRefreshHelper$", "-test.timeout=45s",
	}
	command := exec.CommandContext(ctx, "docker", args...)

	var output bytes.Buffer

	command.Stdout, command.Stderr = &output, &output
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		// Cleanup also covers a killed Docker CLI. Expected --rm absence is
		// handled by the same owned-container cleanup used in production.
		if err := removeWorkerContainer(cleanup, "docker", name); err != nil {
			t.Error(err)
		}
	})

	go func() {
		err := command.Wait()
		if err != nil {
			err = fmt.Errorf("pinned storage container: %w: %s", err, output.String())
		}

		done <- err
	}()
}

func waitStorageFiles(t *testing.T, directory string, names ...string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	for _, name := range names {
		for {
			if _, err := os.Stat(filepath.Join(directory, name)); err == nil {
				break
			}

			select {
			case <-ctx.Done():
				t.Fatal("pinned storage fixture barrier timed out")
			case <-time.After(20 * time.Millisecond):
			}
		}
	}
}

// The independently owned test process runs only inside a selected fixture
// container. Both containers reach the barrier before either refresh begins.
func TestPiCredentialRefreshHelper(t *testing.T) {
	id := os.Getenv("CHR_PI_STORAGE_WORKER")
	if id == "" {
		t.Skip("owned pinned storage helper")
	}

	if id != "first" && id != "second" {
		t.Fatal("invalid storage fixture identity")
	}

	if err := os.WriteFile(filepath.Join("/control", id), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	waitStorageFiles(t, "/control", "release")

	err := WithWorkerCredential(t.Context(), "/owner/auth.json", "/private/auth.json", "fixture", func() error {
		output, err := exec.CommandContext(t.Context(), "node", "/fixture.mjs").CombinedOutput()
		if err != nil {
			return fmt.Errorf("pinned storage action: %w: %s", err, output)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
