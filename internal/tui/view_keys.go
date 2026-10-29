package tui

type interactionKeyHint struct {
	Active bool
	Text   string
}

var viewKeyHints = map[View]string{
	ViewCases:     "Keys: enter opens case  g guided use cases  n new  s save  r refresh  / filter",
	ViewDumps:     "Keys: enter edit  a append  alt+e external  c create  r refresh",
	ViewQuizzes:   "Keys: s start  m Deep/Fast  u resume  x cancel  enter answer  r refresh  alt+e source",
	ViewSpecs:     "Keys: c create  e edit  g generate spec  f generate feature  r refresh  alt+e source",
	ViewPlans:     "Keys: g prepare plan + tickets  G regenerate  r refresh  alt+e source",
	ViewBoard:     "Keys: c create  e edit  a comment  v review  r refresh  alt+e source  left/right column  alt+a attention",
	ViewDash:      "Keys: enter start selected  v preview  w workflow  [/] page  a run all ready  p priority  o order  x repair  s stop",
	ViewProject:   "Keys: enter edit  s save  r refresh  h Harness defaults",
	ViewHelp:      "Keys: esc return  up/down scroll  : commands",
	ViewTickets:   "Keys: c create  e edit  a comment  v review  r refresh  up/down select  / filter  alt+e source",
	ViewDocuments: "Keys: up/down select  alt+e source",
	ViewAuthoring: "Keys: up/down select  alt+e source",
	ViewAgent:     "Keys: tab next pane  up/down scroll  : commands",
	ViewRuntime:   "Keys: tab next pane  up/down scroll  : commands",
}
