package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// AcceptanceFeatureDirectory defines a package boundary constant.
	AcceptanceFeatureDirectory = "acceptance"
)

const (
	// AcceptanceFeatureStatusGenerated defines a package boundary constant.
	AcceptanceFeatureStatusGenerated = "generated"
	// AcceptanceFeatureStatusNeedsMoreInfo defines a package boundary constant.
	AcceptanceFeatureStatusNeedsMoreInfo = "needs_more_info"
	// AcceptanceFeatureStatusRejected defines a package boundary constant.
	AcceptanceFeatureStatusRejected = "rejected"
	// AcceptanceFeatureStatusNotApplicable defines a package boundary constant.
	AcceptanceFeatureStatusNotApplicable = "not_applicable"
)

// AcceptanceFeatureGeneration describes optional model-generated feature content.
type AcceptanceFeatureGeneration struct {
	Status               string
	Body                 string
	RejectionReason      string
	ClarificationPrompts []AcceptanceFeaturePrompt
}

// AcceptanceFeaturePrompt describes one default-oriented feature clarification prompt.
type AcceptanceFeaturePrompt struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	DefaultAnswer string `json:"default_answer"`
	Reason        string `json:"reason"`
	Answer        string `json:"answer,omitempty"`
}

// AcceptanceFeatureArtifact describes a generated Gherkin feature artifact.
type AcceptanceFeatureArtifact struct {
	Scope            QuizScope
	SpecSlug         string
	SpecPath         string
	UseCaseSlug      string
	Body             string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// AcceptanceFeaturePaths describes the canonical feature artifact paths.
type AcceptanceFeaturePaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// AcceptanceFeatureRoundNotes describes persisted feature-round clarification notes.
type AcceptanceFeatureRoundNotes struct {
	Scope            QuizScope
	Status           string
	RejectionReason  string
	Prompts          []AcceptanceFeaturePrompt
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// AcceptanceFeatureRoundPaths describes the canonical feature-round note paths.
type AcceptanceFeatureRoundPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// AcceptanceFeatureCheckResult describes static feature validation output.
type AcceptanceFeatureCheckResult struct {
	Valid    bool
	Findings []AcceptanceFeatureCheckFinding
}

// AcceptanceFeatureCheckFinding describes one static feature validation finding.
type AcceptanceFeatureCheckFinding struct {
	Code    string
	Message string
	Line    int
}

type acceptanceFeatureScenario struct {
	name    string
	line    int
	tags    []string
	given   bool
	when    bool
	then    bool
	hasStep bool
}

// HasContent reports whether generated feature content is present.
func (g AcceptanceFeatureGeneration) HasContent() bool {
	return strings.TrimSpace(g.Status) != "" ||
		strings.TrimSpace(g.Body) != "" ||
		strings.TrimSpace(g.RejectionReason) != ""
}

// ValidateAcceptanceFeatureGeneration validates the corresponding package boundary value.
func ValidateAcceptanceFeatureGeneration(feature AcceptanceFeatureGeneration) (AcceptanceFeatureGeneration, error) {
	normalized := AcceptanceFeatureGeneration{
		Status:               strings.TrimSpace(feature.Status),
		Body:                 strings.TrimSpace(feature.Body),
		RejectionReason:      strings.TrimSpace(feature.RejectionReason),
		ClarificationPrompts: normalizeAcceptanceFeaturePrompts(feature.ClarificationPrompts),
	}

	switch normalized.Status {
	case AcceptanceFeatureStatusGenerated:
		if normalized.Body == "" {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature generated status requires body")
		}

		normalized.ClarificationPrompts = nil
	case AcceptanceFeatureStatusNeedsMoreInfo:
		if normalized.RejectionReason == "" {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature needs_more_info status requires reason")
		}

		if len(normalized.ClarificationPrompts) == 0 {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature needs_more_info status requires clarification prompts")
		}

		normalized.Body = ""
	case AcceptanceFeatureStatusRejected:
		if normalized.RejectionReason == "" {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature rejected status requires reason")
		}

		normalized.Body = ""
		normalized.ClarificationPrompts = nil
	case AcceptanceFeatureStatusNotApplicable:
		normalized.Body = ""
		normalized.ClarificationPrompts = nil
	default:
		return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature status is invalid: %s", normalized.Status)
	}

	return normalized, nil
}

func normalizeAcceptanceFeaturePrompts(prompts []AcceptanceFeaturePrompt) []AcceptanceFeaturePrompt {
	normalized := make([]AcceptanceFeaturePrompt, 0, len(prompts))
	seen := map[string]bool{}

	for _, prompt := range prompts {
		cleaned := AcceptanceFeaturePrompt{
			ID:            strings.TrimSpace(prompt.ID),
			Question:      strings.TrimSpace(prompt.Question),
			DefaultAnswer: strings.TrimSpace(prompt.DefaultAnswer),
			Reason:        strings.TrimSpace(prompt.Reason),
			Answer:        strings.TrimSpace(prompt.Answer),
		}

		if cleaned.ID == "" || cleaned.Question == "" || cleaned.DefaultAnswer == "" {
			continue
		}

		if seen[cleaned.ID] {
			continue
		}

		seen[cleaned.ID] = true
		normalized = append(normalized, cleaned)
	}

	return normalized
}

// AcceptanceFeatureLayout executes the corresponding package operation.
func AcceptanceFeatureLayout(projectDir string, scope QuizScope) (AcceptanceFeaturePaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return AcceptanceFeaturePaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return AcceptanceFeaturePaths{}, err
	}

	relativePath := UserRelativePath(AcceptanceFeatureDirectory, scopeID+".feature")

	return AcceptanceFeaturePaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

// AcceptanceFeatureRoundLayout executes the corresponding package operation.
func AcceptanceFeatureRoundLayout(projectDir string, scope QuizScope) (AcceptanceFeatureRoundPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, err
	}

	relativePath := UserRelativePath(AcceptanceFeatureDirectory, scopeID+"-feature-round.md")

	return AcceptanceFeatureRoundPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

// WriteAcceptanceFeature executes the corresponding package operation.
func WriteAcceptanceFeature(
	ctx context.Context,
	projectDir string,
	feature AcceptanceFeatureArtifact,
) (AcceptanceFeaturePaths, AcceptanceFeatureCheckResult, error) {
	err := ctx.Err()
	if err != nil {
		return AcceptanceFeaturePaths{}, AcceptanceFeatureCheckResult{}, err
	}

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		return AcceptanceFeaturePaths{}, AcceptanceFeatureCheckResult{}, err
	}

	if !result.Valid {
		return AcceptanceFeaturePaths{}, result, fmt.Errorf("acceptance feature static checks failed")
	}

	paths, err := AcceptanceFeatureLayout(projectDir, feature.Scope)
	if err != nil {
		return AcceptanceFeaturePaths{}, AcceptanceFeatureCheckResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return AcceptanceFeaturePaths{}, AcceptanceFeatureCheckResult{}, fmt.Errorf("create acceptance feature directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderAcceptanceFeature(feature)), 0o644)
	if err != nil {
		return AcceptanceFeaturePaths{}, AcceptanceFeatureCheckResult{}, fmt.Errorf("write acceptance feature: %w", err)
	}

	return paths, result, nil
}

// WriteAcceptanceFeatureRoundNotes executes the corresponding package operation.
func WriteAcceptanceFeatureRoundNotes(ctx context.Context, projectDir string, notes AcceptanceFeatureRoundNotes) (AcceptanceFeatureRoundPaths, error) {
	err := ctx.Err()
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, err
	}

	err = ValidateAcceptanceFeatureRoundNotes(notes)
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, err
	}

	paths, err := AcceptanceFeatureRoundLayout(projectDir, notes.Scope)
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, fmt.Errorf("create acceptance feature round directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderAcceptanceFeatureRoundNotes(paths, notes)), 0o644)
	if err != nil {
		return AcceptanceFeatureRoundPaths{}, fmt.Errorf("write acceptance feature round notes: %w", err)
	}

	return paths, nil
}

// RenderAcceptanceFeature renders the corresponding package data.
func RenderAcceptanceFeature(feature AcceptanceFeatureArtifact) string {
	return strings.TrimSpace(feature.Body) + "\n"
}

// RenderAcceptanceFeatureRoundNotes renders the corresponding package data.
func RenderAcceptanceFeatureRoundNotes(paths AcceptanceFeatureRoundPaths, notes AcceptanceFeatureRoundNotes) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("title: " + yamlString("Acceptance Feature Round") + "\n")
	output.WriteString("status: " + yamlString(notes.Status) + "\n")
	output.WriteString("type: " + yamlString("acceptance_feature_round") + "\n")
	output.WriteString("case: " + yamlString(notes.Scope.CaseSlug) + "\n")
	output.WriteString("source_request: " + yamlString(notes.SourceRequestID) + "\n")
	output.WriteString("source_response: " + yamlString(notes.SourceResponseID) + "\n")
	output.WriteString("updated_at: " + yamlString(notes.CreatedAt.UTC().Format(time.RFC3339)) + "\n")
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

// MergeAcceptanceFeatureRoundAnswers copies existing Markdown answers into matching new prompts.
func MergeAcceptanceFeatureRoundAnswers(notes AcceptanceFeatureRoundNotes, existingContent string) AcceptanceFeatureRoundNotes {
	answers := acceptanceFeatureRoundAnswers(existingContent)
	if len(answers) == 0 {
		return notes
	}

	merged := notes

	merged.Prompts = make([]AcceptanceFeaturePrompt, len(notes.Prompts))
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

func acceptanceFeatureRoundAnswers(content string) map[string]string {
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

// ValidateAcceptanceFeatureArtifact validates the corresponding package boundary value.
func ValidateAcceptanceFeatureArtifact(feature AcceptanceFeatureArtifact) error {
	err := ValidateQuizScope(feature.Scope)
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

// ValidateAcceptanceFeatureRoundNotes validates the corresponding package boundary value.
func ValidateAcceptanceFeatureRoundNotes(notes AcceptanceFeatureRoundNotes) error {
	err := ValidateQuizScope(notes.Scope)
	if err != nil {
		return err
	}

	if strings.TrimSpace(notes.Status) != AcceptanceFeatureStatusNeedsMoreInfo {
		return fmt.Errorf("acceptance feature round status must be needs_more_info")
	}

	if strings.TrimSpace(notes.RejectionReason) == "" {
		return fmt.Errorf("acceptance feature round rejection reason is required")
	}

	if len(normalizeAcceptanceFeaturePrompts(notes.Prompts)) == 0 {
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

// CheckAcceptanceFeatureArtifact executes deterministic static feature checks.
func CheckAcceptanceFeatureArtifact(feature AcceptanceFeatureArtifact) (AcceptanceFeatureCheckResult, error) {
	err := ValidateAcceptanceFeatureArtifact(feature)
	if err != nil {
		return AcceptanceFeatureCheckResult{}, err
	}

	checker := acceptanceFeatureChecker{
		lines: strings.Split(feature.Body, "\n"),
	}
	checker.check()

	return AcceptanceFeatureCheckResult{
		Valid:    len(checker.findings) == 0,
		Findings: checker.findings,
	}, nil
}

type acceptanceFeatureChecker struct {
	lines        []string
	findings     []AcceptanceFeatureCheckFinding
	pendingTags  []string
	featureTags  []string
	featureSeen  bool
	inBackground bool
	background   acceptanceFeatureScenario
	scenarios    []acceptanceFeatureScenario
	currentIndex int
}

func (c *acceptanceFeatureChecker) check() {
	c.currentIndex = -1

	for index, line := range c.lines {
		c.checkLine(index+1, strings.TrimSpace(line))
	}

	c.checkDocument()
}

func (c *acceptanceFeatureChecker) checkLine(lineNumber int, line string) {
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

func (c *acceptanceFeatureChecker) checkStructuralLine(lineNumber int, line string) {
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
	case acceptanceFeatureStepKind(line) != "":
		c.addStep(lineNumber, line)
	}
}

func (c *acceptanceFeatureChecker) checkUnsafeLineContent(lineNumber int, lower string) {
	if acceptanceFeatureHasOpenQuestion(lower) {
		c.addFinding("open_question", "feature content must not depend on open questions", lineNumber)
	}

	if acceptanceFeatureHasVagueStep(lower) {
		c.addFinding("vague_step", "feature steps must describe concrete observable behavior", lineNumber)
	}

	if acceptanceFeatureHasImplementationDetail(lower) {
		c.addFinding("implementation_detail", "feature steps must avoid internal implementation details", lineNumber)
	}
}

func (c *acceptanceFeatureChecker) addScenario(lineNumber int, line string) {
	name := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "Scenario Outline:"), "Scenario:"))
	c.scenarios = append(c.scenarios, acceptanceFeatureScenario{
		name: name,
		line: lineNumber,
		tags: append([]string(nil), c.pendingTags...),
	})
	c.currentIndex = len(c.scenarios) - 1
	c.pendingTags = nil
}

func (c *acceptanceFeatureChecker) addStep(lineNumber int, line string) {
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

	switch acceptanceFeatureStepKind(line) {
	case "given":
		scenario.given = true
	case "when":
		scenario.when = true
	case "then":
		scenario.then = true
	}
}

func (c *acceptanceFeatureChecker) addBackgroundStep(line string) {
	c.background.hasStep = true

	switch acceptanceFeatureStepKind(line) {
	case "given":
		c.background.given = true
	case "when":
		c.background.when = true
	case "then":
		c.background.then = true
	}
}

func (c *acceptanceFeatureChecker) checkDocument() {
	if !c.featureSeen {
		c.addFinding("missing_feature", "feature must include a Feature heading", 0)
	}

	if len(c.scenarios) == 0 {
		c.addFinding("missing_scenario", "feature must include at least one scenario", 0)
	}

	for _, scenario := range c.scenarios {
		c.checkScenario(scenario)
	}
}

func (c *acceptanceFeatureChecker) checkScenario(scenario acceptanceFeatureScenario) {
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

	if !acceptanceFeatureHasTraceability(append(c.featureTags, scenario.tags...)) {
		c.addFinding("missing_traceability", "scenario must link to spec behavior", scenario.line)
	}
}

func (c *acceptanceFeatureChecker) addFinding(code, message string, line int) {
	c.findings = append(c.findings, AcceptanceFeatureCheckFinding{
		Code:    code,
		Message: message,
		Line:    line,
	})
}

func acceptanceFeatureStepKind(line string) string {
	lower := strings.ToLower(strings.TrimSpace(line))
	for _, kind := range []string{"given", "when", "then", "and", "but"} {
		if lower == kind || strings.HasPrefix(lower, kind+" ") {
			return kind
		}
	}

	return ""
}

func acceptanceFeatureHasTraceability(tags []string) bool {
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

func acceptanceFeatureHasOpenQuestion(line string) bool {
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

func acceptanceFeatureHasVagueStep(line string) bool {
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

func acceptanceFeatureHasImplementationDetail(line string) bool {
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
