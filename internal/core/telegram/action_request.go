package telegram

import (
	"context"
	"slices"
)

// ActionRequest reads one exact peer-owned prompt for diagnostics. It
// cannot claim an answer, refresh guards or authorize an operation.
func (s *ReceiverStore) ActionRequest(ctx context.Context, peer PeerIdentity, answer ActionAnswer) (OperationRequest, error) {
	if !validActionAnswer(answer) {
		return OperationRequest{}, ErrActionUnavailable
	}

	action, err := readAction(ctx, s.db, answer.Token, peer)
	if err != nil {
		return OperationRequest{}, err
	}

	if !slices.Contains(action.Choices, answer.Choice) {
		return OperationRequest{}, ErrActionUnavailable
	}

	return readOperationID(ctx, s.db, action.RequestID)
}

// RequestAction reads the original prompt for the request's exact peer.
// It neither extends expiry nor creates another action or answer binding.
func (s *ReceiverStore) RequestAction(ctx context.Context, request OperationRequest) (Action, error) {
	var token string

	err := s.db.QueryRowContext(ctx, `SELECT token FROM telegram_actions WHERE request_id=?`, request.ID).Scan(&token)
	if err != nil {
		return Action{}, err
	}

	return readAction(ctx, s.db, token, request.Source.Peer)
}
