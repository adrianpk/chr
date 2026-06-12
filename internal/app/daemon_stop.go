package app

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

type daemonStopRequestResult struct {
	RequestedAt    time.Time
	AlreadyStopped bool
}

func requestDaemonStopState(
	ctx context.Context,
	env core.UserEnvironment,
	assets fs.FS,
	requestedAt time.Time,
) (daemonStopRequestResult, error) {
	repository, err := core.NewDaemonStateRepository(ctx, env, assets)
	if err != nil {
		return daemonStopRequestResult{}, err
	}

	defer func() {
		//nolint:errcheck // Stop result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.RequestDaemonStop(ctx, requestedAt)
	if errors.Is(err, core.ErrDaemonStateNotFound) {
		return daemonStopRequestResult{AlreadyStopped: true}, nil
	}

	if err != nil {
		return daemonStopRequestResult{}, err
	}

	return daemonStopRequestResult{RequestedAt: requestedAt}, nil
}
