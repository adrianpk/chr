package todo

import (
	"errors"
	"strings"
)

var (
	ErrEmptyTitle = errors.New("task title is required")
	ErrNotFound   = errors.New("task not found")
)

type Task struct {
	ID        int
	Title     string
	Completed bool
}

type Store struct {
	tasks []Task
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) AddTask(title string) Task {
	task := Task{
		ID:    len(s.tasks) + 1,
		Title: strings.TrimSpace(title),
	}

	s.tasks = append(s.tasks, task)

	return task
}

func (s *Store) ListTasks() []Task {
	tasks := make([]Task, len(s.tasks))
	copy(tasks, s.tasks)

	return tasks
}

func (s *Store) DeleteTask(id int) error {
	for index, task := range s.tasks {
		if task.ID == id {
			s.tasks = append(s.tasks[:index], s.tasks[index+1:]...)

			return nil
		}
	}

	return ErrNotFound
}

func (s *Store) CompleteTask(id int) error {
	return errors.New("complete task is not implemented")
}
