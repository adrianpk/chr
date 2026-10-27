package core

import (
	"os"
	"reflect"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

// Legacy reads neither initialize an absent database nor mutate existing
// records; project filtering preserves the execution/loop ownership boundary.
func TestLegacyWorkObservation(t *testing.T) {
	env, err := NewUserEnvironment(UserEnvironmentOptions{DevMode: true, DevRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	absent, err := executioncore.ReadLegacyWork(t.Context(), env, "project-1")
	if err != nil || len(absent.Executions) != 0 {
		t.Fatal(absent, err)
	}

	if _, err = os.Stat(env.DataDir()); !os.IsNotExist(err) {
		t.Fatal("observation initialized user data", err)
	}

	seedObservedLegacyWork(t, env)

	before := nativeQueueImages(t, env.DataDir())

	observed, err := executioncore.ReadLegacyWork(t.Context(), env, "project-1")
	if err != nil || len(observed.Executions) != 1 || observed.Executions[0].ProjectID != "project-1" {
		t.Fatal(observed, err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, env.DataDir())) {
		t.Fatal("read changed database images")
	}
}

func seedObservedLegacyWork(t *testing.T, env UserEnvironment) {
	t.Helper()

	executions, err := executioncore.NewExecutionRepository(t.Context(), env, testMigrationFS())
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	for _, project := range []string{"project-1", "other-project"} {
		record := testExecutionRecord(now)

		record.ID, record.ProjectID = "exec-"+project, project
		if err = executions.CreateExecution(t.Context(), record); err != nil {
			t.Fatal(err)
		}
	}

	if err = executions.Close(); err != nil {
		t.Fatal(err)
	}
}
