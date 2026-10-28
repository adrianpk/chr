package workflow

import "context"

// ReserveRecordMutation binds the exact arguments before a caller checks an
// external source guard. A confirmed result reconciles before current guards;
// otherwise failed dispatch cannot exchange this identity for a new payload.
func (s *Store) ReserveRecordMutation(ctx context.Context, request MutationRequest) (result MutationResult, found bool, err error) {
	request, digest, err := s.normalizeMutation(request)
	if err != nil {
		return result, false, err
	}

	err = s.withLock(ctx, func() error {
		var readErr error

		result, found, readErr = s.readMutationReceipt(request, digest)
		if readErr != nil || found {
			return readErr
		}

		return s.reserveMutationBinding(request, digest)
	})

	return result, found, err
}
