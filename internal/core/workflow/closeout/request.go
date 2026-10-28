package closeout

// Request requires the observed set and canonical target branch.
// InvocationID is optional for an explicit standalone decision, including an
// empty set. It cannot take ownership from another unfinished invocation.
type Request struct {
	SetPath            string
	ExpectedRevision   uint64
	ExpectedSourceHash string
	Repository         string
	Branch             string
	InvocationID       string
}
