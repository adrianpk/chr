# Model failures

ErrCapacity and ErrOutput retain the existing provider-neutral structured model failure identities. Callers share these sentinels across provider and retained operation boundaries. This package owns no provider calls, slots, retries, dispatch or storage.

ExtractJSONObject retains the original model-output object extraction rule used
by structured decoding. It does not validate the resulting domain payload or
perform provider IO; the consuming model contract owns semantic validation.
