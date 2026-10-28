package workflow

import (
	"context"
	"time"
)

func watchWorkflowStop(watched context.Context, store *Store, watch workflowStopWatch) {
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-watched.Done():
				return
			case <-ticker.C:
				if err := store.checkWorkflowClaimStop(watched, watch.Claim); err != nil {
					watch.Cancel()

					return
				}
			}
		}
	}()
}
