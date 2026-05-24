package chc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ParseQuizRawCaptureEntries(content string) ([]QuizEntry, error) {
	blocks := strings.Split(content, "\n## Turn\n\n")
	entries := []QuizEntry{}

	for index, block := range blocks[1:] {
		entry, err := parseQuizRawTurn(index+1, block)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func CurateQuizEntries(scope QuizScope, entries []QuizEntry, now time.Time) QuizCuratedNotes {
	answered := quizAnsweredEntriesByTopic(entries)

	return QuizCuratedNotes{
		Scope:                      scope,
		Status:                     "curated",
		Decisions:                  quizDecisionNotes(answered),
		Constraints:                quizKeywordNotes(entries, quizConstraintKeywords()),
		Preferences:                quizKeywordNotes(entries, quizPreferenceKeywords()),
		OpenQuestions:              quizOpenQuestionNotes(entries),
		ImplementationImplications: quizImplementationImplications(answered, entries),
		Entries:                    entries,
		UpdatedAt:                  now.UTC().Format(time.RFC3339),
	}
}

func parseQuizRawTurn(index int, block string) (QuizEntry, error) {
	entry := QuizEntry{
		ID: fmt.Sprintf("entry-%03d", index),
	}
	metadata, body := splitQuizTurnMetadata(block)

	entry.QuestionID = metadata["question_id"]
	entry.Status = QuizEntryStatus(metadata["status"])
	entry.Mode = QuizMode(metadata["mode"])
	entry.Topic = metadata["topic"]
	entry.Depth = parseQuizTurnDepth(metadata["depth"])
	entry.Question = sectionAfterLabel(body, "Question:")
	entry.Answer = sectionAfterLabel(body, "Answer:")
	entry.CreatedAt = quizRawTurnTime(block)
	entry.AnsweredAt = entry.CreatedAt

	err := ValidateQuizEntry(entry)
	if err != nil {
		return QuizEntry{}, err
	}

	return entry, nil
}

func splitQuizTurnMetadata(block string) (map[string]string, string) {
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
		return metadata, ""
	}

	return metadata, parts[1]
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

func parseQuizTurnDepth(value string) int {
	depth, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 1
	}

	return depth
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

func quizAnsweredEntriesByTopic(entries []QuizEntry) map[string][]QuizEntry {
	answered := map[string][]QuizEntry{}

	for _, entry := range entries {
		if entry.Status != QuizEntryStatusAnswered {
			continue
		}

		topic := strings.TrimSpace(entry.Topic)
		answered[topic] = append(answered[topic], entry)
	}

	return answered
}

func quizDecisionNotes(answered map[string][]QuizEntry) []string {
	notes := []string{}
	for _, topic := range sortedQuizTopics(answered) {
		notes = append(notes, topic+": "+joinQuizAnswerInsights(answered[topic]))
	}

	return notes
}

func quizKeywordNotes(entries []QuizEntry, keywords []string) []string {
	notes := []string{}

	for _, entry := range entries {
		if entry.Status != QuizEntryStatusAnswered {
			continue
		}

		answer := strings.TrimSpace(entry.Answer)
		if containsQuizKeyword(answer, keywords) {
			notes = append(notes, entry.Topic+": "+firstQuizSentence(answer))
		}
	}

	return compactUniqueStrings(notes)
}

func quizOpenQuestionNotes(entries []QuizEntry) []string {
	notes := []string{}

	for _, entry := range entries {
		if entry.Status == QuizEntryStatusSkipped {
			notes = append(notes, entry.Topic+": "+strings.TrimSpace(entry.Question))
		}
	}

	return notes
}

func quizImplementationImplications(answered map[string][]QuizEntry, entries []QuizEntry) []string {
	if len(entries) == 0 {
		return []string{"No quiz answers have been captured yet."}
	}

	return []string{
		fmt.Sprintf("Use %d answered quiz topic groups as source material for draft specification behavior.", len(answered)),
		"Preserve unresolved skipped questions as review prompts before accepting the generated spec.",
	}
}

func sortedQuizTopics(entries map[string][]QuizEntry) []string {
	topics := make([]string, 0, len(entries))
	for topic := range entries {
		topics = append(topics, topic)
	}

	sort.Strings(topics)

	return topics
}

func joinQuizAnswerInsights(entries []QuizEntry) string {
	insights := []string{}

	for _, entry := range entries {
		insights = append(insights, firstQuizSentence(entry.Answer))
		if len(insights) == 3 {
			break
		}
	}

	return strings.Join(compactUniqueStrings(insights), "; ")
}

func firstQuizSentence(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	for _, marker := range []string{". ", "! ", "? "} {
		if before, _, ok := strings.Cut(value, marker); ok {
			return strings.TrimSpace(before) + strings.TrimSpace(marker)
		}
	}

	return value
}

func containsQuizKeyword(value string, keywords []string) bool {
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
