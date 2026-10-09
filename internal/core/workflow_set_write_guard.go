package core

import (
	"fmt"
	"reflect"
)

func guardWorkflowSetChange(old, next WorkflowSet) error {
	if old.Status == "completed" {
		return fmt.Errorf("completed set requires an explicit approved reopening")
	}

	if old.Status == "closing" || !reflect.DeepEqual(old.Closing, next.Closing) {
		return fmt.Errorf("set closing ownership requires a fenced operation")
	}

	if old.ClosureRef != next.ClosureRef || !reflect.DeepEqual(old.ClosureHistory, next.ClosureHistory) || !reflect.DeepEqual(old.Amendments, next.Amendments) {
		return fmt.Errorf("set closure and amendment evidence require an owned operation")
	}

	return nil
}
