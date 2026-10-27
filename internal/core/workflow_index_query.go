package core

import (
	"context"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowQuerySQL(q WorkflowQuery) (string, error) {
	switch q.Kind {
	case "ticket":
		return `SELECT representation_ref FROM indexed_tickets WHERE project_id = ? AND namespace = ? AND workflow_revision > 0 AND (? = '' OR slice_id = ?) AND (? = '' OR use_case = ?) AND (? = 0 OR slice_set_id = '') ORDER BY slice_set_id, slice_id, task_order, key LIMIT ? OFFSET ?`, nil
	case "plan":
		return `SELECT path FROM workflow_sets WHERE project_id = ? AND namespace = ? AND (? = '' OR id = ?) AND (? = '' OR use_case = ?) AND ? = 0 ORDER BY use_case_order, created_at, id LIMIT ? OFFSET ?`, nil
	case "slice":
		return `SELECT id FROM workflow_slices WHERE project_id = ? AND namespace = ? AND (? = '' OR set_id = ?) AND ? = '' AND ? = '' AND ? = 0 ORDER BY set_id, slice_order, id LIMIT ? OFFSET ?`, nil
	case "receipt":
		return `SELECT path FROM workflow_receipts WHERE project_id = ? AND namespace = ? AND (? = '' OR unit_id = ?) AND ? = '' AND ? = '' AND ? = 0 ORDER BY confirmed_at, id LIMIT ? OFFSET ?`, nil
	}

	return "", fmt.Errorf("unsupported workflow query kind: %s", q.Kind)
}

func (i *workflowTicketIndex) queryWorkflowPage(ctx context.Context, store *WorkflowStore, snapshot WorkflowSnapshot, q WorkflowQuery) (WorkflowPage, error) {
	query, err := workflowQuerySQL(q)
	if err != nil {
		return WorkflowPage{}, err
	}

	rows, err := i.db.QueryContext(ctx, query, store.projectID, store.layout.Namespace(), q.OwnerID, q.OwnerID, q.UseCase, q.UseCase, q.LooseOnly, q.Limit+1, q.Offset)
	if err != nil {
		return WorkflowPage{}, err
	}

	defer func() {
		//nolint:errcheck // rows.Err is checked after iteration.
		_ = rows.Close()
	}()

	page := WorkflowPage{NextOffset: -1}
	count := 0

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return WorkflowPage{}, err
		}

		if count == q.Limit {
			page.NextOffset = q.Offset + q.Limit

			break
		}

		if err := appendWorkflowQueryResult(&page, snapshot, q.Kind, key); err != nil {
			return WorkflowPage{}, err
		}

		count++
	}

	return page, rows.Err()
}

func appendWorkflowQueryResult(page *WorkflowPage, snapshot WorkflowSnapshot, kind, key string) error {
	for _, r := range snapshot.Records {
		if kind != "slice" && r.Path == key {
			page.Records = append(page.Records, r)

			return nil
		}

		if kind == "slice" && r.Set != nil {
			for _, slice := range r.Set.Slices {
				if slice.ID == key {
					page.Slices = append(page.Slices, slice)

					return nil
				}
			}
		}
	}

	return fmt.Errorf("%w: indexed workflow reference is absent", workflowstate.ErrConflict)
}
