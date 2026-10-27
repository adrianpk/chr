package telegram

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkReference points to a confirmed admission and its current local
// conversation fence. Canonical workflow files retain execution authority.
type WorkReference struct {
	Request              OperationRequest
	Invocation           InvocationBinding
	ConversationRevision int64
}

// WorkReferences pages at most sixteen bound conversations, without
// admitting or claiming work. The caller cycles an empty page back to the start.
func (s *ReceiverStore) WorkReferences(ctx context.Context, botID int64, after string, limit int) (result []WorkReference, err error) {
	if botID <= 0 || len(after) > 128 || limit < 1 || limit > 16 {
		return nil, workflowstate.ErrConflict
	}

	rows, err := s.db.QueryContext(ctx, `SELECT r.id,c.invocation_id,COALESCE(c.last_observed_revision,0),c.revision FROM telegram_conversations c JOIN telegram_requests r ON r.id=c.authorization_ref WHERE c.bot_id=? AND c.id>? AND c.phase IN ('executing','waiting') AND r.operation_id='work.admit' AND r.state='confirmed' AND r.conversation_id=c.id AND r.domain_receipt_ref='workflow:'||c.invocation_id ORDER BY c.id LIMIT ?`, botID, after, limit)
	if err != nil {
		return nil, err
	}

	page, err := scanWorkReferences(rows)
	result, ids := page.References, page.IDs

	if err != nil {
		return nil, err
	}

	for index, id := range ids {
		result[index].Request, err = readOperationID(ctx, s.db, id)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func scanWorkReferences(rows *sql.Rows) (page workReferencePage, err error) {
	var (
		result []WorkReference
		ids    []string
	)

	defer func() { err = errors.Join(err, rows.Close()) }()

	for rows.Next() {
		var (
			ref WorkReference
			id  string
		)

		if err := rows.Scan(&id, &ref.Invocation.ID, &ref.Invocation.Revision, &ref.ConversationRevision); err != nil {
			return workReferencePage{References: nil, IDs: nil}, err
		}

		result, ids = append(result, ref), append(ids, id)
	}

	return workReferencePage{References: result, IDs: ids}, rows.Err()
}

// ValidateWorkReference checks that the same confirmed admission still
// binds this current conversation fence. Hosts separately check current authority.
func (s *ReceiverStore) ValidateWorkReference(ctx context.Context, ref WorkReference) error {
	var count int

	r, target := ref.Request, ref.Request.Target

	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_conversations c JOIN telegram_requests r ON r.id=c.authorization_ref WHERE c.id=? AND c.revision=? AND c.invocation_id=? AND c.project_id=? AND c.namespace=? AND c.project_binding_digest=? AND c.phase IN ('executing','waiting') AND r.id=? AND r.argument_digest=? AND r.operator_id=? AND r.operation_id='work.admit' AND r.state='confirmed' AND r.domain_receipt_ref='workflow:'||c.invocation_id`, target.ConversationID, ref.ConversationRevision, ref.Invocation.ID, target.ProjectID, target.Namespace, target.ProjectBindingDigest, r.ID, r.ArgumentDigest, r.Operator).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return workflowstate.ErrConflict
	}

	return nil
}

// ObserveWork retains a monotonic read projection after an activation.
// It does not change phase, authorization, native checkpoints or the original result.
func (s *ReceiverStore) ObserveWork(ctx context.Context, ref WorkReference, revision uint64, now time.Time) error {
	if revision == 0 || revision > math.MaxInt64 {
		return workflowstate.ErrConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET last_observed_revision=?,updated_at=? WHERE id=? AND revision=? AND invocation_id=? AND authorization_ref=? AND COALESCE(last_observed_revision,0)<=?`, revision, database.FormatTime(now), ref.Request.Target.ConversationID, ref.ConversationRevision, ref.Invocation.ID, ref.Request.ID, revision)

		return err
	})
}
