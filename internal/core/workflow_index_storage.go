package core

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/database"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// WorkflowTicketIndex composes ticket storage with the existing native projections.
// Both projections share the same SQLite handle and original cache invalidation.
type WorkflowTicketIndex interface {
	ticketcore.TicketIndex
	WorkflowIndex
	WorkflowIntegrationIndex
}

type workflowTicketIndex struct {
	ticketcore.TicketIndex
	db *sql.DB
}

// NewWorkflowTicketIndex constructs the corresponding package service or value.
func NewWorkflowTicketIndex(ctx context.Context, env database.DataDirectory, assetsFS fs.FS) (ticketcore.TicketIndex, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), database.UserFileName))
	if err != nil {
		return nil, fmt.Errorf("open ticket index database: %w", err)
	}

	err = database.Configure(ctx, db)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	index, err := NewWorkflowTicketIndexFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	return index, nil
}

// NewWorkflowTicketIndexFromDB assembles ticket and native projections on one handle.
func NewWorkflowTicketIndexFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (WorkflowTicketIndex, error) {
	tickets, err := ticketcore.NewTicketIndexFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, err
	}

	return &workflowTicketIndex{TicketIndex: tickets, db: db}, nil
}
