package state

import (
	"fmt"
	"sort"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TaskInputHash detects membership or contract changes in a slice.
func TaskInputHash(records []Record, sliceID string) string {
	tasks := []Task{}

	for _, record := range records {
		if record.Task != nil && record.Task.SliceID == sliceID {
			tasks = append(tasks, *record.Task)
		}
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Order < tasks[j].Order })

	values := []string{}
	for _, task := range tasks {
		values = append(values, fmt.Sprintf("%d:%q:%q", task.Order, task.Key, task.ContractHash))
	}

	return contenthash.Sum([]byte(strings.Join(values, "\n")))
}
