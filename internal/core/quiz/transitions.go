package quiz

import (
	"fmt"
	"strings"
	"time"
)

// QuestionTags maps the question topic to its original model tag list.
func QuestionTags(question Question) []string {
	if strings.TrimSpace(question.Topic) == "" {
		return []string{}
	}

	return []string{question.Topic}
}

// EntryTags retains explicit tags or uses the recorded topic.
func EntryTags(entry Entry) []string {
	if len(entry.Tags) > 0 {
		return entry.Tags
	}

	if strings.TrimSpace(entry.Topic) == "" {
		return []string{}
	}

	return []string{entry.Topic}
}

// EntryFromState records the active question with the original answer and tag rules.
func EntryFromState(state RuntimeState, status EntryStatus, answer string, now time.Time) Entry {
	question := *state.ActiveQuestion

	return Entry{
		ID:           fmt.Sprintf("entry-%03d", state.QuestionsAsked),
		QuestionID:   question.ID,
		Question:     question.Prompt,
		Answer:       strings.TrimSpace(answer),
		Status:       status,
		Topic:        question.Topic,
		Depth:        question.Depth,
		ParentID:     question.ParentID,
		Mode:         state.Mode,
		Tags:         QuestionTags(question),
		PromptIntent: question.Rationale,
		CreatedAt:    question.CreatedAt,
		AnsweredAt:   now,
	}
}

// StateAfterEntry retains counters and bounded branch-return behavior.
func StateAfterEntry(state RuntimeState, entry Entry) RuntimeState {
	if entry.Status == EntryStatusAnswered {
		state.AnswersRecorded++
	}

	if entry.Status == EntryStatusSkipped {
		state.SkipsRecorded++
	}

	state.Status = RuntimeStatusIdle
	state.ActiveQuestion = nil
	state.ActiveTopic = ""

	if entry.Status == EntryStatusSkipped {
		state.Trail = nil
		state.BranchReturnReason = BranchReturnReasonSkipped
	} else if entry.Depth >= MaxFollowUpDepth {
		state.Trail = nil
		state.BranchReturnReason = BranchReturnReasonMaxDepth
	} else {
		state.BranchReturnReason = ""
	}

	if state.Mode == ModeDeep {
		state.Candidates = nil
	}

	return state
}

// ActivateQuestion advances the question identity, trail and question count.
func ActivateQuestion(state RuntimeState, question Question, now time.Time) RuntimeState {
	state.Status = RuntimeStatusActive
	state.ActiveQuestion = &question
	state.ActiveTopic = question.Topic
	state.Trail = trailForQuestion(state.Trail, question)
	state.BranchReturnReason = ""
	state.QuestionsAsked++
	state.UpdatedAt = now

	return state
}

// ActivateCandidate maps a buffered Fast candidate to the active question.
func ActivateCandidate(state RuntimeState, candidate Candidate, now time.Time) RuntimeState {
	question := Question{
		ID:        candidate.ID,
		Prompt:    candidate.Prompt,
		Required:  true,
		Topic:     candidate.Topic,
		Depth:     candidate.Depth,
		ParentID:  candidate.ParentID,
		TrunkID:   candidate.ID,
		RouteID:   candidate.RouteID,
		CreatedAt: candidate.CreatedAt,
	}
	if question.CreatedAt.IsZero() {
		question.CreatedAt = now
	}

	if question.Depth > 1 {
		question.TrunkID = candidate.RouteID
		if strings.TrimSpace(question.TrunkID) == "" {
			question.TrunkID = candidate.ParentID
		}
	}

	state = ActivateQuestion(state, question, now)

	return state
}

func trailForQuestion(trail []TrailNode, question Question) []TrailNode {
	node := TrailNode{
		QuestionID: question.ID,
		Prompt:     question.Prompt,
		Topic:      question.Topic,
		Depth:      question.Depth,
	}

	if question.Depth <= 1 {
		return []TrailNode{node}
	}

	parentIndex := -1

	for index, current := range trail {
		if current.QuestionID == question.ParentID {
			parentIndex = index

			break
		}
	}

	if parentIndex < 0 {
		return []TrailNode{node}
	}

	next := append([]TrailNode(nil), trail[:parentIndex+1]...)
	next = append(next, node)

	return next
}
