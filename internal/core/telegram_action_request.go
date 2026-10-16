package core

import (
	"context"
	"slices"
)

// TelegramActionRequest reads one exact peer-owned prompt for diagnostics. It
// cannot claim an answer, refresh guards or authorize an operation.
func (s *TelegramReceiverStore) TelegramActionRequest(ctx context.Context, peer TelegramPeerIdentity, answer TelegramActionAnswer) (TelegramOperationRequest, error) {
	if !validTelegramActionAnswer(answer) {
		return TelegramOperationRequest{}, ErrTelegramActionUnavailable
	}

	action, err := readTelegramAction(ctx, s.db, answer.Token, peer)
	if err != nil {
		return TelegramOperationRequest{}, err
	}

	if !slices.Contains(action.Choices, answer.Choice) {
		return TelegramOperationRequest{}, ErrTelegramActionUnavailable
	}

	return readTelegramOperationID(ctx, s.db, action.RequestID)
}

// TelegramRequestAction reads the original prompt for the request's exact peer.
// It neither extends expiry nor creates another action or answer binding.
func (s *TelegramReceiverStore) TelegramRequestAction(ctx context.Context, request TelegramOperationRequest) (TelegramAction, error) {
	var token string

	err := s.db.QueryRowContext(ctx, `SELECT token FROM telegram_actions WHERE request_id=?`, request.ID).Scan(&token)
	if err != nil {
		return TelegramAction{}, err
	}

	return readTelegramAction(ctx, s.db, token, request.Source.Peer)
}
