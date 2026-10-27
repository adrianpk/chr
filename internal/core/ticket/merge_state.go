package ticket

// MergeState records the ticket merge lifecycle.
type MergeState string

// Merge states preserve the existing ticket metadata spellings.
const (
	MergeStatePending MergeState = "pending"
	MergeStateMerged  MergeState = "merged"
	MergeStateBlocked MergeState = "blocked"
)

// MergeResult records the merge outcome without conflating it with review decisions.
type MergeResult string

// Merge results preserve outcomes independently of the merge lifecycle state.
const (
	MergeResultMerged       MergeResult = "merged"
	MergeResultBlocked      MergeResult = "blocked"
	MergeResultNotMergeable MergeResult = "not_mergeable"
	MergeActorMerger                    = "merger"
)
