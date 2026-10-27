package app

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	daemoncore "forge.adrianpk.com/chr/internal/core/harness/daemon"
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
	repository, err := daemoncore.NewRepository(ctx, env, assets)
	if err != nil {
		return daemonStopRequestResult{}, err
	}

	defer func() {
		//nolint:errcheck // Stop result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.RequestStop(ctx, requestedAt)
	if errors.Is(err, daemoncore.ErrStateNotFound) {
		return daemonStopRequestResult{AlreadyStopped: true}, nil
	}

	if err != nil {
		return daemonStopRequestResult{}, err
	}

	return daemonStopRequestResult{RequestedAt: requestedAt}, nil
}
