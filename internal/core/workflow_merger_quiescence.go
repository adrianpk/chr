package core

import (
	"context"
	"time"
)

func waitWorkflowMergerProcessGroup(ctx context.Context, process WorkflowProcess, deadline time.Time) error {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		alive, err := workflowProcessGroupAlive(process)
		if err != nil {
			return err
		}

		if !alive {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
