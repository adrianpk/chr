package authoring

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLoadGoGraphContextRunsBuildAndSelectedQueries(t *testing.T) {
	runner := &fakeGoGraphRunner{path: "/usr/local/bin/gograph"}

	ctx, err := LoadGoGraphContext(GoGraphContextOptions{
		ProjectDir: t.TempDir(),
		Terms:      []string{"CompleteTask", "Store", "CompleteTask"},
		Runner:     runner,
	})
	if err != nil {
		t.Fatalf("LoadGoGraphContext returned error: %v", err)
	}

	if !ctx.Available {
		t.Fatal("gograph context was not marked available")
	}

	assertGoGraphCommand(t, runner.commands[0], []string{"/usr/local/bin/gograph", "build", "."})
	assertGoGraphCommand(t, runner.commands[1], []string{"/usr/local/bin/gograph", "changes"})
	assertGoGraphCommand(t, runner.commands[2], []string{"/usr/local/bin/gograph", "hotspot", "--top", "20"})

	rendered := RenderGoGraphContext(ctx)
	for _, want := range []string{"Go Graph Context", "gograph build .", "output: context"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered gograph context missing %q:\n%s", want, rendered)
		}
	}
}

func TestLoadGoGraphContextSupportsPreciseBuild(t *testing.T) {
	runner := &fakeGoGraphRunner{path: "/bin/gograph"}

	_, err := LoadGoGraphContext(GoGraphContextOptions{
		ProjectDir: t.TempDir(),
		Precise:    true,
		Runner:     runner,
	})
	if err != nil {
		t.Fatalf("LoadGoGraphContext returned error: %v", err)
	}

	assertGoGraphCommand(t, runner.commands[0], []string{"/bin/gograph", "build", ".", "--precise"})
}

func TestLoadGoGraphContextDegradesWhenUnavailable(t *testing.T) {
	ctx, err := LoadGoGraphContext(GoGraphContextOptions{
		ProjectDir: t.TempDir(),
		Runner: &fakeGoGraphRunner{
			err: errors.New("not found"),
		},
	})
	if err != nil {
		t.Fatalf("LoadGoGraphContext returned error: %v", err)
	}

	if ctx.Available {
		t.Fatal("unavailable gograph context was marked available")
	}

	if len(ctx.Warnings) != 1 {
		t.Fatalf("Warnings = %d, want 1", len(ctx.Warnings))
	}

	if ctx.Sources[0].Included {
		t.Fatal("unavailable gograph source was marked included")
	}
}

type fakeGoGraphRunner struct {
	path     string
	err      error
	commands [][]string
}

func (r *fakeGoGraphRunner) LookPath(file string) (string, error) {
	if r.err != nil {
		return "", r.err
	}

	if r.path != "" {
		return r.path, nil
	}

	return file, nil
}

func (r *fakeGoGraphRunner) Run(
	_ context.Context,
	_ string,
	command string,
	args ...string,
) GoGraphCommandResult {
	fullCommand := append([]string{command}, args...)
	r.commands = append(r.commands, fullCommand)

	name := "build"
	if len(args) > 0 {
		name = args[0]
	}

	return GoGraphCommandResult{
		Command: fullCommand,
		Stdout:  "output: " + name,
	}
}

func assertGoGraphCommand(t *testing.T, got, want []string) {
	t.Helper()

	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}
