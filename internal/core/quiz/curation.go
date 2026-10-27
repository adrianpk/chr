package quiz

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ParseRawCaptureEntries parses the corresponding package data.
func ParseRawCaptureEntries(content string) ([]Entry, error) {
	if len(content) > workflowstate.FileLimit {
		return nil, fmt.Errorf("quiz raw capture exceeds file limit")
	}

	blocks := strings.Split(content, "\n## Turn\n\n")
	if len(blocks) > workflowstate.RecordLimit+1 {
		return nil, fmt.Errorf("quiz entry capacity exceeded")
	}

	entries := []Entry{}

	for index, block := range blocks[1:] {
		entry, err := parseRawTurn(index+1, block)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// CurateEntries executes the corresponding package operation.
func CurateEntries(scope Scope, entries []Entry, now time.Time) CuratedNotes {
	answered := quizAnsweredEntriesByTopic(entries)

	return CuratedNotes{
		Scope:                      scope,
		Status:                     "curated",
		Decisions:                  quizDecisionNotes(answered),
		AcceptanceReadiness:        quizAcceptanceReadinessNotes(entries),
		Constraints:                quizKeywordNotes(entries, quizConstraintKeywords()),
		Preferences:                quizKeywordNotes(entries, quizPreferenceKeywords()),
		OpenQuestions:              quizOpenQuestionNotes(entries),
		ImplementationImplications: quizImplementationImplications(answered, entries),
		Entries:                    entries,
		UpdatedAt:                  now.UTC().Format(time.RFC3339),
	}
}

func quizAcceptanceReadinessNotes(entries []Entry) []string {
	seen := map[string]bool{}
	notes := []string{}

	for _, entry := range entries {
		readiness := quizAcceptanceReadinessFromPromptIntent(entry.PromptIntent)
		if readiness == nil {
			continue
		}

		decision := strings.TrimSpace(readiness["decision"])
		if decision == "" {
			continue
		}

		note := "Decision: " + decision
		if missing := strings.TrimSpace(readiness["missing"]); missing != "" {
			note += "; missing: " + missing
		}

		if inferred := strings.TrimSpace(readiness["inferred"]); inferred != "" {
			note += "; inferred: " + inferred
		}

		if seen[note] {
			continue
		}

		seen[note] = true
		notes = append(notes, note)
	}

	sort.Strings(notes)

	return notes
}

func quizAcceptanceReadinessFromPromptIntent(value string) map[string]string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil
	}

	metadata, ok := payload["metadata"].(map[string]any)
	if !ok {
		return nil
	}

	routingState, ok := metadata["routing_state"].(map[string]any)
	if !ok {
		return nil
	}

	readiness, ok := routingState["acceptance_readiness"].(map[string]any)
	if !ok {
		return nil
	}

	return map[string]string{
		"decision": inferredString(readiness["decision"]),
		"missing":  inferredStringList(readiness["missing"]),
		"inferred": inferredReadinessSlots(readiness["inferred_elements"]),
	}
}

func inferredString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(text)
}

func inferredStringList(value any) string {
	values, ok := value.([]any)
	if !ok {
		return ""
	}

	texts := []string{}

	for _, value := range values {
		text := inferredString(value)
		if text != "" {
			texts = append(texts, text)
		}
	}

	return strings.Join(texts, ", ")
}

func inferredReadinessSlots(value any) string {
	values, ok := value.(map[string]any)
	if !ok {
		return ""
	}

	slots := []string{}

	for slot, slotValues := range values {
		if inferredStringList(slotValues) != "" {
			slots = append(slots, slot)
		}
	}

	sort.Strings(slots)

	return strings.Join(slots, ", ")
}

func parseRawTurn(index int, block string) (Entry, error) {
	entry := Entry{
		ID: fmt.Sprintf("entry-%03d", index),
	}
	turn := splitTurnMetadata(block)
	metadata, body := turn.Metadata, turn.Body

	if encoded, exists := metadata["entry_json"]; exists {
		data, err := base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return entry, err
		}

		if err := json.Unmarshal(data, &entry); err != nil {
			return entry, err
		}

		return entry, ValidateEntry(entry)
	}

	entry.QuestionID = metadata["question_id"]
	entry.Status = EntryStatus(metadata["status"])
	entry.Mode = Mode(metadata["mode"])
	entry.Topic = metadata["topic"]
	entry.Depth = parseTurnDepth(metadata["depth"])
	entry.ParentID = metadata["parent_id"]
	entry.Tags = parseTurnTags(metadata["tags"])
	entry.PromptIntent = sectionAfterLabel(body, "Prompt intent:")
	entry.Question = sectionAfterLabel(body, "Question:")
	entry.Answer = sectionAfterLabel(body, "Answer:")
	entry.CreatedAt = quizRawTurnTime(block)
	entry.AnsweredAt = entry.CreatedAt

	err := ValidateEntry(entry)
	if err != nil {
		return Entry{}, err
	}

	return entry, nil
}

type turnMetadata struct {
	Metadata map[string]string
	Body     string
}

func splitTurnMetadata(block string) turnMetadata {
	parts := strings.SplitN(block, "\n\n", 2)
	metadata := map[string]string{}

	for _, line := range strings.Split(parts[0], "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "- "))

		key, value, ok := strings.Cut(line, ":")
		if ok {
			metadata[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}

	if len(parts) == 1 {
		return turnMetadata{Metadata: metadata}
	}

	return turnMetadata{Metadata: metadata, Body: parts[1]}
}

func sectionAfterLabel(body, label string) string {
	_, rest, ok := strings.Cut(body, label)
	if !ok {
		return ""
	}

	rest = strings.TrimLeft(rest, "\n")
	value, _, _ := strings.Cut(rest, "\n\n")

	return strings.TrimSpace(value)
}

func parseTurnDepth(value string) int {
	depth, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 1
	}

	return depth
}

func parseTurnTags(value string) []string {
	tags := []string{}

	for _, tag := range strings.Split(value, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}

func quizRawTurnTime(block string) time.Time {
	for _, line := range strings.Split(block, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok || key != "updated_at" {
			continue
		}

		parsed, err := time.Parse(time.RFC3339, strings.Trim(strings.TrimSpace(value), `"`))
		if err == nil {
			return parsed.UTC()
		}
	}

	return time.Unix(1, 0).UTC()
}

func quizAnsweredEntriesByTopic(entries []Entry) map[string][]Entry {
	answered := map[string][]Entry{}

	for _, entry := range entries {
		if entry.Status != EntryStatusAnswered {
			continue
		}

		topic := strings.TrimSpace(entry.Topic)
		answered[topic] = append(answered[topic], entry)
	}

	return answered
}

func quizDecisionNotes(answered map[string][]Entry) []string {
	notes := []string{}
	for _, topic := range sortedTopics(answered) {
		notes = append(notes, topic+": "+joinAnswerInsights(answered[topic]))
	}

	return notes
}

func quizKeywordNotes(entries []Entry, keywords []string) []string {
	notes := []string{}

	for _, entry := range entries {
		if entry.Status != EntryStatusAnswered {
			continue
		}

		answer := strings.TrimSpace(entry.Answer)
		if containsKeyword(answer, keywords) {
			notes = append(notes, entry.Topic+": "+firstSentence(answer))
		}
	}

	return compactUniqueStrings(notes)
}

func quizOpenQuestionNotes(entries []Entry) []string {
	notes := []string{}

	for _, entry := range entries {
		if entry.Status == EntryStatusSkipped {
			notes = append(notes, entry.Topic+": "+strings.TrimSpace(entry.Question))
		}
	}

	return notes
}

func quizImplementationImplications(answered map[string][]Entry, entries []Entry) []string {
	if len(entries) == 0 {
		return []string{"No quiz answers have been captured yet."}
	}

	return []string{
		fmt.Sprintf("Use %d answered quiz topic groups as source material for draft specification behavior.", len(answered)),
		"Preserve unresolved skipped questions as review prompts before accepting the generated spec.",
	}
}

func sortedTopics(entries map[string][]Entry) []string {
	topics := make([]string, 0, len(entries))
	for topic := range entries {
		topics = append(topics, topic)
	}

	sort.Strings(topics)

	return topics
}

func joinAnswerInsights(entries []Entry) string {
	insights := []string{}

	for _, entry := range entries {
		insights = append(insights, firstSentence(entry.Answer))
		if len(insights) == 3 {
			break
		}
	}

	return strings.Join(compactUniqueStrings(insights), "; ")
}

func firstSentence(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	for _, marker := range []string{". ", "! ", "? "} {
		if before, _, ok := strings.Cut(value, marker); ok {
			return strings.TrimSpace(before) + strings.TrimSpace(marker)
		}
	}

	return value
}

func containsKeyword(value string, keywords []string) bool {
	value = strings.ToLower(value)
	for _, keyword := range keywords {
		if strings.Contains(value, keyword) {
			return true
		}
	}

	return false
}

func compactUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	compact := []string{}

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		compact = append(compact, value)
	}

	return compact
}

func quizConstraintKeywords() []string {
	return []string{"must", "need", "required", "cannot", "do not", "never"}
}

func quizPreferenceKeywords() []string {
	return []string{"prefer", "want", "like", "should"}
}
