package acceptance

import (
	"strings"
)

// CheckResult describes static feature validation output.
type CheckResult struct {
	Valid    bool
	Findings []Finding
}

// Finding describes one static feature validation finding.
type Finding struct {
	Code    string
	Message string
	Line    int
}

type scenario struct {
	name    string
	line    int
	tags    []string
	text    []string
	given   bool
	when    bool
	then    bool
	hasStep bool
}

// Check executes deterministic static feature checks.
func Check(feature Artifact) (CheckResult, error) {
	err := Validate(feature)
	if err != nil {
		return CheckResult{}, err
	}

	checker := checker{
		lines: strings.Split(feature.Body, "\n"),
	}
	checker.check()

	return CheckResult{
		Valid:    len(checker.findings) == 0,
		Findings: checker.findings,
	}, nil
}

type checker struct {
	lines        []string
	findings     []Finding
	pendingTags  []string
	featureTags  []string
	featureSeen  bool
	inBackground bool
	background   scenario
	scenarios    []scenario
	currentIndex int
}

func (c *checker) check() {
	c.currentIndex = -1

	for index, line := range c.lines {
		c.checkLine(index+1, strings.TrimSpace(line))
	}

	c.checkDocument()
}

func (c *checker) checkLine(lineNumber int, line string) {
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	if strings.HasPrefix(line, "@") {
		c.pendingTags = append(c.pendingTags, strings.Fields(line)...)

		return
	}

	lower := strings.ToLower(line)
	c.checkStructuralLine(lineNumber, line)
	c.checkUnsafeLineContent(lineNumber, lower)
}

func (c *checker) checkStructuralLine(lineNumber int, line string) {
	switch {
	case strings.HasPrefix(line, "Feature:"):
		c.featureSeen = true
		c.featureTags = append([]string(nil), c.pendingTags...)
		c.pendingTags = nil
		c.inBackground = false
	case strings.HasPrefix(line, "Background:"):
		c.inBackground = true
		c.currentIndex = -1
	case strings.HasPrefix(line, "Scenario:") || strings.HasPrefix(line, "Scenario Outline:"):
		c.inBackground = false
		c.addScenario(lineNumber, line)
	case stepKind(line) != "":
		c.addStep(lineNumber, line)
	}
}

func (c *checker) checkUnsafeLineContent(lineNumber int, lower string) {
	if hasOpenQuestion(lower) {
		c.addFinding("open_question", "feature content must not depend on open questions", lineNumber)
	}

	if hasVagueStep(lower) {
		c.addFinding("vague_step", "feature steps must describe concrete observable behavior", lineNumber)
	}

	if hasImplementationDetail(lower) {
		c.addFinding("implementation_detail", "feature steps must avoid internal implementation details", lineNumber)
	}
}

func (c *checker) addScenario(lineNumber int, line string) {
	name := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "Scenario Outline:"), "Scenario:"))
	c.scenarios = append(c.scenarios, scenario{
		name: name,
		line: lineNumber,
		tags: append([]string(nil), c.pendingTags...),
		text: []string{name},
	})
	c.currentIndex = len(c.scenarios) - 1
	c.pendingTags = nil
}

func (c *checker) addStep(lineNumber int, line string) {
	if c.currentIndex < 0 {
		if c.inBackground {
			c.addBackgroundStep(line)

			return
		}

		c.addFinding("step_without_scenario", "feature step must belong to a scenario", lineNumber)

		return
	}

	scenario := &c.scenarios[c.currentIndex]
	scenario.hasStep = true
	scenario.text = append(scenario.text, line)

	switch stepKind(line) {
	case "given":
		scenario.given = true
	case "when":
		scenario.when = true
	case "then":
		scenario.then = true
	}
}

func (c *checker) addBackgroundStep(line string) {
	c.background.hasStep = true

	switch stepKind(line) {
	case "given":
		c.background.given = true
	case "when":
		c.background.when = true
	case "then":
		c.background.then = true
	}
}

func (c *checker) checkDocument() {
	if !c.featureSeen {
		c.addFinding("missing_feature", "feature must include a Feature heading", 0)
	}

	if len(c.scenarios) == 0 {
		c.addFinding("missing_scenario", "feature must include at least one scenario", 0)
	}

	c.checkBehaviorDepth()

	for _, scenario := range c.scenarios {
		c.checkScenario(scenario)
	}
}

func (c *checker) checkBehaviorDepth() {
	if len(c.scenarios) == 0 {
		return
	}

	if len(c.scenarios) == 1 {
		c.addFinding("insufficient_behavior_paths", "feature must cover more than one observable behavior path", c.scenarios[0].line)

		return
	}

	if !hasContrastPath(c.scenarios) {
		c.addFinding("missing_contrast_path", "feature must include at least one rejection, failure, boundary, terminal, or no-change path", c.scenarios[0].line)
	}
}

func (c *checker) checkScenario(scenario scenario) {
	if strings.TrimSpace(scenario.name) == "" {
		c.addFinding("missing_scenario_name", "scenario must have a name", scenario.line)
	}

	if !scenario.given && !c.background.given {
		c.addFinding("missing_given", "scenario must include a Given step", scenario.line)
	}

	if !scenario.when {
		c.addFinding("missing_when", "scenario must include a When step", scenario.line)
	}

	if !scenario.then {
		c.addFinding("missing_then", "scenario must include a Then step", scenario.line)
	}

	if !scenario.hasStep {
		c.addFinding("missing_steps", "scenario must include executable steps", scenario.line)
	}

	if !hasTraceability(append(c.featureTags, scenario.tags...)) {
		c.addFinding("missing_traceability", "scenario must link to spec behavior", scenario.line)
	}
}

func (c *checker) addFinding(code, message string, line int) {
	c.findings = append(c.findings, Finding{
		Code:    code,
		Message: message,
		Line:    line,
	})
}

func hasContrastPath(scenarios []scenario) bool {
	for _, scenario := range scenarios {
		text := strings.ToLower(strings.Join(scenario.text, " "))
		if textHasContrastSignal(text) {
			return true
		}
	}

	return false
}

func textHasContrastSignal(text string) bool {
	for _, phrase := range []string{
		"already",
		"blocked",
		"boundary",
		"cannot",
		"denied",
		"draw",
		"duplicate",
		"empty",
		"ends",
		"error",
		"exhausted",
		"fail",
		"failure",
		"full",
		"invalid",
		"missing",
		"occupied",
		"reject",
		"rejected",
		"remains",
		"terminal",
		"timeout",
		"unavailable",
		"unchanged",
	} {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	return false
}

func stepKind(line string) string {
	lower := strings.ToLower(strings.TrimSpace(line))
	for _, kind := range []string{"given", "when", "then", "and", "but"} {
		if lower == kind || strings.HasPrefix(lower, kind+" ") {
			return kind
		}
	}

	return ""
}

func hasTraceability(tags []string) bool {
	for _, tag := range tags {
		switch {
		case strings.HasPrefix(tag, "@spec:"),
			strings.HasPrefix(tag, "@requirement:"),
			strings.HasPrefix(tag, "@acceptance-criterion:"),
			strings.HasPrefix(tag, "@behavior:"):
			return true
		}
	}

	return false
}

func hasOpenQuestion(line string) bool {
	for _, phrase := range []string{
		"open question",
		"tbd",
		"todo",
		"unknown",
		"unspecified",
		"decide later",
	} {
		if strings.Contains(line, phrase) {
			return true
		}
	}

	return false
}

func hasVagueStep(line string) bool {
	for _, phrase := range []string{
		"works correctly",
		"works as expected",
		"does the right thing",
		"handles it properly",
		"handles appropriately",
		"successful result",
	} {
		if strings.Contains(line, phrase) {
			return true
		}
	}

	return false
}

func hasImplementationDetail(line string) bool {
	for _, phrase := range []string{
		"database table",
		"sql query",
		"internal package",
		"private method",
		"goroutine",
		"implementation detail",
	} {
		if strings.Contains(line, phrase) {
			return true
		}
	}

	return false
}
