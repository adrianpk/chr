package tui

type interactionKeyHint struct {
	Active bool
	Text   string
}

var viewKeyHints = map[View]string{
	ViewBoard:     "Keys: F1-F9 or click changes tab, tab changes pane, left/right changes Board column, alt+a toggles attention, alt+e edits source, up/down selects or scrolls, / filters, q quits",
	ViewTickets:   "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, / filters, q quits",
	ViewDocuments: "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, q quits",
	ViewProject:   "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects field, enter edits, s saves, r refreshes, q quits",
	ViewCases:     "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects focused pane, enter opens or edits, g guided use cases, s saves, r refreshes, q quits",
	ViewDumps:     "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects or scrolls, enter edits, alt+e external, c creates, r refreshes, q quits",
	ViewQuizzes:   "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, s starts, enter answers, r refreshes, q quits",
	ViewSpecs:     "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, g generates spec, r refreshes, q quits",
	ViewPlans:     "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, g generates missing plan, G regenerates, r refreshes, q quits",
	ViewDash:      "Keys: F1-F9 or click changes tab, tab changes pane, a runs all ready tickets, enter starts selected ticket, x repairs readiness, s requests stop, q quits",
	ViewAuthoring: "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, q quits",
}
