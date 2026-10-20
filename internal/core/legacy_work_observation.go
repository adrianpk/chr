package core

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"slices"
)

// LegacyWorkObservation retains a bounded page of legacy work records.
type LegacyWorkObservation struct {
	Executions []ExecutionRecord
	Loops      []LoopRunRecord
}

// ReadLegacyWork opens existing state in SQLite read-only mode. An absent
// database means no retained work; this operation never initializes a schema.
func ReadLegacyWork(ctx context.Context, env UserEnvironment, projectID string) (result LegacyWorkObservation, err error) {
	path := UserDatabasePath(env)
	if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return result, nil
	}

	if err != nil {
		return result, err
	}

	uri := url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}

	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return result, err
	}

	db.SetMaxOpenConns(1)
	defer func() { err = errors.Join(err, db.Close()) }()

	result.Executions, err = readObservedExecutions(ctx, db, projectID)
	if err == nil {
		result.Loops, err = readObservedLoops(ctx, db, projectID)
	}

	return result, err
}

func readObservedExecutions(ctx context.Context, db *sql.DB, projectID string) ([]ExecutionRecord, error) {
	rows, err := db.QueryContext(ctx, selectExecutionSQL()+" WHERE project_id = ? ORDER BY CASE WHEN status IN ('pending','running') THEN 0 ELSE 1 END, updated_at DESC, id LIMIT ?", projectID, WorkflowPageLimit)
	if err != nil {
		return nil, err
	}

	records, readErr := scanExecutionRows(rows)

	return records, errors.Join(readErr, rows.Close())
}

func readObservedLoops(ctx context.Context, db *sql.DB, projectID string) ([]LoopRunRecord, error) {
	rows, err := db.QueryContext(ctx, selectLoopRunSQL()+" WHERE project_id = ? ORDER BY CASE WHEN status IN ('pending','running') THEN 0 ELSE 1 END, updated_at DESC, id LIMIT ?", projectID, WorkflowPageLimit)
	if err != nil {
		return nil, err
	}

	var records []LoopRunRecord

	for rows.Next() {
		record, readErr := scanLoopRun(rows)
		if readErr != nil {
			return nil, errors.Join(readErr, rows.Close())
		}

		records = append(records, record)
	}

	return records, errors.Join(rows.Err(), rows.Close())
}

// LegacyWorkQueue separates shared eligible order from display metadata.
type LegacyWorkQueue struct {
	Tickets, Eligible []IndexedTicket
	Excluded          int
}

// ReadLegacyWorkQueue reads authoritative files without refreshing SQL.
func ReadLegacyWorkQueue(ctx context.Context, layout OpsLayout, projectID string) (result LegacyWorkQueue, err error) {
	result.Tickets, err = NewWorkflowStore(layout, projectID).readBoardTickets(ctx, false)
	if err != nil {
		return result, err
	}

	tickets, excluded, err := FilterLegacyWorkTickets(ctx, result.Tickets)
	if err != nil {
		return result, err
	}

	result.Excluded = excluded

	for _, ticket := range tickets {
		if slices.Contains(defaultLoopEligibleStatuses, ticket.Status) {
			result.Eligible = append(result.Eligible, ticket)
		}
	}

	order, err := normalizeTicketLoopOrder(TicketLoopOrderCreatedAt)
	if err != nil {
		return result, err
	}

	slices.SortFunc(result.Eligible, order.compare)

	return result, nil
}
