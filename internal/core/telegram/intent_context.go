package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntentSnapshot binds interpretation to one peer's observed context.
type IntentSnapshot struct {
	Selected                string
	Focused                 string
	LastWorkConversation    string
	RepliedWorkConversation string
	Sessions                []Conversation
	Digest                  string
}

// IntentContext reads selection and sessions in one transaction. Its
// digest can guard target resolution after a provider call without holding a TX.
func (s *ReceiverStore) IntentContext(ctx context.Context, input ConversationInput) (snapshot IntentSnapshot, err error) {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return snapshot, workflowstate.ErrConflict
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		var err error

		snapshot, err = op.intentSnapshot()

		return err
	})

	return snapshot, err
}

func (op conversationTransaction) intentSnapshot() (IntentSnapshot, error) {
	var (
		snapshot IntentSnapshot
		err      error
	)

	selection, selectionError := op.peerSelection()

	snapshot.Selected, snapshot.Focused, err = selection.ProjectID, selection.RootDigest, selectionError
	if err != nil {
		return snapshot, err
	}

	snapshot.Sessions, err = readConversations(op.ctx, op.tx, op.input.Update.Peer)
	if err != nil {
		return snapshot, err
	}

	snapshot.LastWorkConversation, err = readWorkMessageConversation(op.ctx, op.tx, op.input.Update.Peer, 0)
	if err != nil {
		return snapshot, err
	}

	if op.input.Update.ReplyToMessageID > 0 {
		snapshot.RepliedWorkConversation, err = readWorkMessageConversation(op.ctx, op.tx, op.input.Update.Peer, op.input.Update.ReplyToMessageID)
		if err != nil {
			return snapshot, err
		}
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return snapshot, err
	}

	snapshot.Digest = contenthash.Sum(encoded)

	return snapshot, nil
}

// ResolveConversationTarget freezes the same exact/focused/sole routing
// used by capture operations, before an interpreted decision is retained.
func (s *ReceiverStore) ResolveConversationTarget(ctx context.Context, input ConversationInput) (OperationTarget, error) {
	var target OperationTarget
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return target, workflowstate.ErrConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		conversation, err := op.target()
		if err != nil {
			return err
		}

		target, err = op.operationTarget(conversation)

		return err
	})

	return target, err
}
