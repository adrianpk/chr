package main

import (
	"fmt"

	"example.com/todolist/internal/todo"
)

func main() {
	store := todo.NewStore()
	task := store.AddTask("write proving ground")

	fmt.Printf("%d %s\n", task.ID, task.Title)
}
