package tui

// TicketEditorChoice is a canonical metadata choice mapped for the terminal.
type TicketEditorChoice struct{ Value, Label string }

// TicketEditorField carries its stored value, options and canonical restriction.
type TicketEditorField struct {
	Name, Label, Value, ReadOnly string
	Multiline                    bool
	Options                      []TicketEditorChoice
}

// TicketEditorView observes the exact ticket and its existing editing rules.
type TicketEditorView struct {
	New                                                            bool
	Key, Revision, CaseSlug, Description, Discussion, Phase, Slice string
	ReadOnlyReason, ReviewReadOnly                                 string
	Fields                                                         []TicketEditorField
}
