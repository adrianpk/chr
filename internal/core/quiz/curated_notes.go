package quiz

// CuratedNotes describes a package boundary value.
type CuratedNotes struct {
	Scope                      Scope
	Status                     string
	Decisions                  []string
	AcceptanceReadiness        []string
	Constraints                []string
	Preferences                []string
	OpenQuestions              []string
	ImplementationImplications []string
	Entries                    []Entry
	UpdatedAt                  string
}
