package tui

type interactionKeyHint struct {
	Active bool
	Text   string
}

var viewKeyHints = map[View]string{
	ViewCases:     "Keys: enter opens case  g guided use cases  n new  s save  r refresh  / filter",
	ViewDumps:     "Keys: enter edit  alt+e external  c create  r refresh",
	ViewQuizzes:   "Keys: s start  enter answer  r refresh  alt+e source",
	ViewSpecs:     "Keys: g generate spec  f generate feature  r refresh  alt+e source",
	ViewPlans:     "Keys: g generate missing plan  G regenerate  r refresh  alt+e source",
	ViewBoard:     "Keys: left/right column  alt+a attention  up/down select  / filter  alt+e source",
	ViewDash:      "Keys: enter start selected  v preview  w workflow  [/] page  a run all ready  p priority  o order  x repair  s stop",
	ViewProject:   "Keys: enter edit  s save  r refresh  h Harness defaults",
	ViewHelp:      "Keys: esc return  up/down scroll  : commands",
	ViewTickets:   "Keys: up/down select  / filter  alt+e source",
	ViewDocuments: "Keys: up/down select  alt+e source",
	ViewAuthoring: "Keys: up/down select  alt+e source",
	ViewAgent:     "Keys: tab next pane  up/down scroll  : commands",
	ViewRuntime:   "Keys: tab next pane  up/down scroll  : commands",
}
