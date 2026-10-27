package quiz

// QuizCuratedNotes describes a package boundary value.
type QuizCuratedNotes struct {
	Scope                      QuizScope
	Status                     string
	Decisions                  []string
	AcceptanceReadiness        []string
	Constraints                []string
	Preferences                []string
	OpenQuestions              []string
	ImplementationImplications []string
	Entries                    []QuizEntry
	UpdatedAt                  string
}
