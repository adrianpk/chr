package todo

import (
	"errors"
	"testing"
)

func TestStoreAddsAndListsTasks(t *testing.T) {
	store := NewStore()

	first := store.AddTask("buy milk")
	second := store.AddTask("write tests")

	if first.ID != 1 {
		t.Fatalf("first.ID = %d, want 1", first.ID)
	}

	if second.ID != 2 {
		t.Fatalf("second.ID = %d, want 2", second.ID)
	}

	tasks := store.ListTasks()
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
}

func TestStoreDoesNotReuseDeletedTaskIDs(t *testing.T) {
	store := NewStore()

	first := store.AddTask("first")
	second := store.AddTask("second")

	if err := store.DeleteTask(first.ID); err != nil {
		t.Fatalf("DeleteTask returned error: %v", err)
	}

	third := store.AddTask("third")

	if third.ID == second.ID {
		t.Fatalf("third.ID = %d reused existing task id", third.ID)
	}

	if third.ID != 3 {
		t.Fatalf("third.ID = %d, want 3", third.ID)
	}
}

func TestStoreCompletesTask(t *testing.T) {
	store := NewStore()
	task := store.AddTask("ship feature")

	if err := store.CompleteTask(task.ID); err != nil {
		t.Fatalf("CompleteTask returned error: %v", err)
	}

	tasks := store.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}

	if !tasks[0].Completed {
		t.Fatal("task is not completed")
	}
}

func TestStoreCompleteTaskReturnsNotFound(t *testing.T) {
	store := NewStore()

	err := store.CompleteTask(99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("CompleteTask error = %v, want ErrNotFound", err)
	}
}
