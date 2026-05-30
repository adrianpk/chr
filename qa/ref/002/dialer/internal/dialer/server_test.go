package dialer

import "testing"

func TestQueueExcludesDoNotCallLeads(t *testing.T) {
	store := NewStore(SeedLeads())

	for _, lead := range store.Queue("agent-1") {
		if lead.Status == LeadDoNotCall {
			t.Fatalf("queue includes do-not-call lead: %#v", lead)
		}
	}
}

func TestRecordCallQualifiesLead(t *testing.T) {
	store := NewStore(SeedLeads())

	_, err := store.RecordCall("lead-100", "agent-1", "qualified")
	if err != nil {
		t.Fatalf("RecordCall returned error: %v", err)
	}

	lead, err := store.Lead("lead-100")
	if err != nil {
		t.Fatalf("Lead returned error: %v", err)
	}

	if lead.Status != LeadQualified {
		t.Fatalf("Status = %q, want %q", lead.Status, LeadQualified)
	}
}
