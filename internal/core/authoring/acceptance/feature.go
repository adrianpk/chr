package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

const (
	// Directory defines a package boundary constant.
	Directory = "acceptance"
)

// Artifact describes a generated Gherkin feature artifact.
type Artifact struct {
	Scope            quizcore.Scope
	SpecSlug         string
	SpecPath         string
	UseCaseSlug      string
	Body             string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// Paths describes the canonical feature artifact paths.
type Paths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// RoundNotes describes persisted feature-round clarification notes.
type RoundNotes struct {
	Scope            quizcore.Scope
	Status           string
	RejectionReason  string
	Prompts          []generationcore.AcceptanceFeaturePrompt
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// RoundPaths describes the canonical feature-round note paths.
type RoundPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

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

// Layout executes the corresponding package operation.
func Layout(projectDir string, scope quizcore.Scope) (Paths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return Paths{}, err
	}

	scopeID, err := quizcore.ScopeID(scope)
	if err != nil {
		return Paths{}, err
	}

	body, err := opscore.NamedArtifactBody(projectDir, "spec", scopeID+"-acceptance-feature")
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

// RoundLayout executes the corresponding package operation.
func RoundLayout(projectDir string, scope quizcore.Scope) (RoundPaths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return RoundPaths{}, err
	}

	scopeID, err := quizcore.ScopeID(scope)
	if err != nil {
		return RoundPaths{}, err
	}

	body, err := opscore.NamedArtifactBody(projectDir, "spec", scopeID+"-acceptance-feature-round")
	if err != nil {
		return RoundPaths{}, err
	}

	return RoundPaths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

// WriteResult retains a confirmed path and the original static-check outcome.
// Failed checks remain observable even when there is no published path.
type WriteResult struct {
	Paths  Paths
	Checks CheckResult
}

// Write executes the corresponding package operation.
func Write(ctx context.Context, projectDir string, feature Artifact) (WriteResult, error) {
	err := ctx.Err()
	if err != nil {
		return WriteResult{}, err
	}

	result, err := Check(feature)
	if err != nil {
		return WriteResult{}, err
	}

	if !result.Valid {
		return WriteResult{Checks: result}, fmt.Errorf("acceptance feature static checks failed")
	}

	paths, err := Layout(projectDir, feature.Scope)
	if err != nil {
		return WriteResult{}, err
	}

	scopeID, err := quizcore.ScopeID(feature.Scope)
	if err != nil {
		return WriteResult{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(projectDir, "spec", scopeID+"-acceptance-feature"); err != nil {
		return WriteResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return WriteResult{}, fmt.Errorf("create acceptance feature directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(Render(feature)), 0o644)
	if err != nil {
		return WriteResult{}, fmt.Errorf("write acceptance feature: %w", err)
	}

	return WriteResult{Paths: paths, Checks: result}, nil
}

// WriteRoundNotes executes the corresponding package operation.
func WriteRoundNotes(ctx context.Context, projectDir string, notes RoundNotes) (RoundPaths, error) {
	err := ctx.Err()
	if err != nil {
		return RoundPaths{}, err
	}

	err = ValidateRoundNotes(notes)
	if err != nil {
		return RoundPaths{}, err
	}

	paths, err := RoundLayout(projectDir, notes.Scope)
	if err != nil {
		return RoundPaths{}, err
	}

	scopeID, err := quizcore.ScopeID(notes.Scope)
	if err != nil {
		return RoundPaths{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(projectDir, "spec", scopeID+"-acceptance-feature-round"); err != nil {
		return RoundPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return RoundPaths{}, fmt.Errorf("create acceptance feature round directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderRoundNotes(paths, notes)), 0o644)
	if err != nil {
		return RoundPaths{}, fmt.Errorf("write acceptance feature round notes: %w", err)
	}

	return paths, nil
}

// Render renders the corresponding package data.
func Render(feature Artifact) string {
	return strings.TrimSpace(feature.Body) + "\n"
}

// RenderRoundNotes renders the corresponding package data.
func RenderRoundNotes(paths RoundPaths, notes RoundNotes) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("title: " + opsmetadata.Quote("Acceptance Feature Round") + "\n")
	output.WriteString("status: " + opsmetadata.Quote(notes.Status) + "\n")
	output.WriteString("type: " + opsmetadata.Quote("acceptance_feature_round") + "\n")
	output.WriteString("case: " + opsmetadata.Quote(notes.Scope.CaseSlug) + "\n")
	output.WriteString("source_request: " + opsmetadata.Quote(notes.SourceRequestID) + "\n")
	output.WriteString("source_response: " + opsmetadata.Quote(notes.SourceResponseID) + "\n")
	output.WriteString("updated_at: " + opsmetadata.Quote(notes.CreatedAt.UTC().Format(time.RFC3339)) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Acceptance Feature Round\n\n")
	output.WriteString("## Missing Information\n\n")
	output.WriteString(strings.TrimSpace(notes.RejectionReason) + "\n\n")
	output.WriteString("## Prompts\n\n")

	for _, prompt := range notes.Prompts {
		output.WriteString("### " + prompt.ID + "\n\n")
		output.WriteString("- Question: " + prompt.Question + "\n")
		output.WriteString("- Default: " + prompt.DefaultAnswer + "\n")

		if prompt.Reason != "" {
			output.WriteString("- Reason: " + prompt.Reason + "\n")
		}

		output.WriteString("- Answer:")

		if prompt.Answer != "" {
			output.WriteString(" " + prompt.Answer)
		}

		output.WriteString("\n\n")
	}

	output.WriteString("## Source\n\n")
	output.WriteString("- feature round: " + paths.RelativePath + "\n")

	return output.String()
}

// MergeRoundAnswers copies existing Markdown answers into matching new prompts.
func MergeRoundAnswers(notes RoundNotes, existingContent string) RoundNotes {
	answers := roundAnswers(existingContent)
	if len(answers) == 0 {
		return notes
	}

	merged := notes

	merged.Prompts = make([]generationcore.AcceptanceFeaturePrompt, len(notes.Prompts))
	for index, prompt := range notes.Prompts {
		merged.Prompts[index] = prompt
		if strings.TrimSpace(prompt.Answer) != "" {
			continue
		}

		answer := answers[prompt.ID]
		if answer != "" {
			merged.Prompts[index].Answer = answer
		}
	}

	return merged
}

func roundAnswers(content string) map[string]string {
	answers := map[string]string{}
	lines := strings.Split(content, "\n")
	currentID := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			currentID = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))

			continue
		}

		if currentID == "" || !strings.HasPrefix(trimmed, "- Answer:") {
			continue
		}

		answer := strings.TrimSpace(strings.TrimPrefix(trimmed, "- Answer:"))
		if answer != "" {
			answers[currentID] = answer
		}
	}

	return answers
}

// Validate validates the corresponding package boundary value.
func Validate(feature Artifact) error {
	err := quizcore.ValidateScope(feature.Scope)
	if err != nil {
		return err
	}

	if strings.TrimSpace(feature.SpecSlug) == "" {
		return fmt.Errorf("acceptance feature spec slug is required")
	}

	if strings.TrimSpace(feature.SpecPath) == "" {
		return fmt.Errorf("acceptance feature spec path is required")
	}

	if strings.TrimSpace(feature.Body) == "" {
		return fmt.Errorf("acceptance feature body is required")
	}

	if feature.CreatedAt.IsZero() {
		return fmt.Errorf("acceptance feature created time is required")
	}

	return nil
}

// ValidateRoundNotes validates the corresponding package boundary value.
func ValidateRoundNotes(notes RoundNotes) error {
	err := quizcore.ValidateScope(notes.Scope)
	if err != nil {
		return err
	}

	if strings.TrimSpace(notes.Status) != generationcore.AcceptanceFeatureStatusNeedsMoreInfo {
		return fmt.Errorf("acceptance feature round status must be needs_more_info")
	}

	if strings.TrimSpace(notes.RejectionReason) == "" {
		return fmt.Errorf("acceptance feature round rejection reason is required")
	}

	if len(generationcore.NormalizeAcceptanceFeaturePrompts(notes.Prompts)) == 0 {
		return fmt.Errorf("acceptance feature round prompts are required")
	}

	if strings.TrimSpace(notes.SourceRequestID) == "" {
		return fmt.Errorf("acceptance feature round source request id is required")
	}

	if strings.TrimSpace(notes.SourceResponseID) == "" {
		return fmt.Errorf("acceptance feature round source response id is required")
	}

	if notes.CreatedAt.IsZero() {
		return fmt.Errorf("acceptance feature round created time is required")
	}

	return nil
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
