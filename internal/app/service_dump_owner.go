package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) CreateDump(request DumpCreateRequest) (result DumpOperationResult, err error) {
	err = core.WithAuthoringMutation(s.ctx, s.projectContext.ProjectDir, func(ctx context.Context) error { s.ctx = ctx; result, err = s.createDump(request); return err })

	return result, err
}

func (s service) AppendDump(request DumpAppendRequest) (result DumpOperationResult, err error) {
	err = core.WithAuthoringMutation(s.ctx, s.projectContext.ProjectDir, func(ctx context.Context) error { s.ctx = ctx; result, err = s.appendDump(request); return err })

	return result, err
}

func (s service) SaveDump(request DumpSaveRequest) (result DumpOperationResult, err error) {
	err = core.WithAuthoringMutation(s.ctx, s.projectContext.ProjectDir, func(ctx context.Context) error { s.ctx = ctx; result, err = s.saveDump(request); return err })

	return result, err
}
