package dialer

import (
	"errors"
	"slices"
	"time"
)

type LeadStatus string

const (
	LeadNew       LeadStatus = "new"
	LeadQueued    LeadStatus = "queued"
	LeadContacted LeadStatus = "contacted"
	LeadQualified LeadStatus = "qualified"
	LeadConverted LeadStatus = "converted"
	LeadLost      LeadStatus = "lost"
	LeadDoNotCall LeadStatus = "do_not_call"
)

type Lead struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Company     string     `json:"company"`
	Phone       string     `json:"phone"`
	Status      LeadStatus `json:"status"`
	OwnerAgent  string     `json:"owner_agent,omitempty"`
	Score       int        `json:"score"`
	NextCallAt  *time.Time `json:"next_call_at,omitempty"`
	LastOutcome string     `json:"last_outcome,omitempty"`
}

type Call struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"lead_id"`
	AgentID   string    `json:"agent_id"`
	Outcome   string    `json:"outcome"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	leads map[string]Lead
	calls []Call
}

var ErrLeadNotFound = errors.New("lead not found")

func NewStore(leads []Lead) *Store {
	index := make(map[string]Lead, len(leads))
	for _, lead := range leads {
		index[lead.ID] = lead
	}

	return &Store{leads: index}
}

func (s *Store) Leads() []Lead {
	leads := make([]Lead, 0, len(s.leads))
	for _, lead := range s.leads {
		leads = append(leads, lead)
	}

	slices.SortFunc(leads, func(left, right Lead) int {
		if left.Score != right.Score {
			return right.Score - left.Score
		}

		return compareString(left.ID, right.ID)
	})

	return leads
}

func (s *Store) Lead(id string) (Lead, error) {
	lead, ok := s.leads[id]
	if !ok {
		return Lead{}, ErrLeadNotFound
	}

	return lead, nil
}

func (s *Store) Queue(agentID string) []Lead {
	queue := []Lead{}
	now := time.Now().UTC()

	for _, lead := range s.leads {
		if lead.OwnerAgent != agentID {
			continue
		}
		if lead.Status != LeadQueued && lead.Status != LeadNew {
			continue
		}
		if lead.NextCallAt != nil && lead.NextCallAt.After(now) {
			continue
		}
		queue = append(queue, lead)
	}

	slices.SortFunc(queue, func(left, right Lead) int {
		if left.Score != right.Score {
			return right.Score - left.Score
		}

		return compareString(left.ID, right.ID)
	})

	return queue
}

func (s *Store) RecordCall(leadID, agentID, outcome string) (Call, error) {
	lead, ok := s.leads[leadID]
	if !ok {
		return Call{}, ErrLeadNotFound
	}

	call := Call{
		ID:        "call-" + leadID + "-" + time.Now().UTC().Format("150405"),
		LeadID:    leadID,
		AgentID:   agentID,
		Outcome:   outcome,
		CreatedAt: time.Now().UTC(),
	}
	s.calls = append(s.calls, call)

	lead.LastOutcome = outcome
	lead.Status = LeadContacted
	if outcome == "qualified" {
		lead.Status = LeadQualified
	}
	if outcome == "not_interested" {
		lead.Status = LeadLost
	}
	s.leads[leadID] = lead

	return call, nil
}

func SeedLeads() []Lead {
	tomorrow := time.Now().UTC().Add(24 * time.Hour)

	return []Lead{
		{ID: "lead-100", Name: "Ada Rivera", Company: "Northwind Labs", Phone: "+1-555-0100", Status: LeadQueued, OwnerAgent: "agent-1", Score: 92},
		{ID: "lead-101", Name: "Ben Carter", Company: "Atlas Freight", Phone: "+1-555-0101", Status: LeadNew, OwnerAgent: "agent-1", Score: 78},
		{ID: "lead-102", Name: "Mina Shah", Company: "Beacon Dental", Phone: "+1-555-0102", Status: LeadDoNotCall, OwnerAgent: "agent-1", Score: 99},
		{ID: "lead-103", Name: "Jon Bell", Company: "Clearwater Foods", Phone: "+1-555-0103", Status: LeadQualified, OwnerAgent: "agent-2", Score: 88},
		{ID: "lead-104", Name: "Nora West", Company: "Pioneer HVAC", Phone: "+1-555-0104", Status: LeadQueued, OwnerAgent: "agent-2", Score: 65, NextCallAt: &tomorrow},
	}
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}

	return 0
}
