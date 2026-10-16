package core

import "context"

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
