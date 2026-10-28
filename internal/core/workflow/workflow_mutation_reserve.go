package workflow

import "context"

// ReserveRecordMutation binds the exact arguments before a caller checks an
// external source guard. A confirmed result reconciles before current guards;
// otherwise failed dispatch cannot exchange this identity for a new payload.
func (s *Store) ReserveRecordMutation(ctx context.Context, request MutationRequest) (retained MutationLookup, err error) {
	normalized, err := s.normalizeMutation(request)
	if err != nil {
		return retained, err
	}

	request, digest := normalized.Request, normalized.Digest

	err = s.withLock(ctx, func() error {
		var readErr error

		retained, readErr = s.readMutationReceipt(request, digest)
		if readErr != nil || retained.Found {
			return readErr
		}

		return s.reserveMutationBinding(request, digest)
	})

	return retained, err
}
