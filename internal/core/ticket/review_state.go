package ticket

// ReviewState records the ticket review transition, independently of provider PR status.
type ReviewState string

// Review states preserve the existing ticket metadata spellings.
const (
	ReviewStatePass                 ReviewState = "pass"
	ReviewStateChangesRequested     ReviewState = "changes_requested"
	ReviewStateInsufficientEvidence ReviewState = "insufficient_evidence"
	ReviewStateNeedsHumanDecision   ReviewState = "needs_human_decision"
)

// ReviewResult records the review decision, independently of the resulting ticket status.
type ReviewResult string

// Review results preserve decisions independently of the resulting review state.
const (
	ReviewResultPass                 ReviewResult = "pass"
	ReviewResultNeedsChanges         ReviewResult = "needs_changes"
	ReviewResultInsufficientEvidence ReviewResult = "insufficient_evidence"
	ReviewResultNeedsHumanDecision   ReviewResult = "needs_human_decision"
)
