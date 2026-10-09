package core

import (
	"testing"
)

// TestWorkAuthority keeps absence transitional and validates explicit routing.
func TestWorkAuthority(t *testing.T) {
	cases := []struct {
		name, value string
		want        WorkStateAuthority
		valid       bool
	}{
		{"absent", "", WorkStateAuthorityTrackerCompat, true},
		{"tickets", "tickets", WorkStateAuthorityTickets, true},
		{"compat", "tracker_compat", WorkStateAuthorityTrackerCompat, true},
		{"unknown", "both", "both", false},
		{"padded", " tickets ", " tickets ", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := t.TempDir()
			writeWorkspaceConfig(t, project, "work:\n  state_authority: "+tc.value+"\n")

			config, err := LoadWorkspaceConfig(project)
			if err != nil {
				t.Fatal(err)
			}
			// The parser trims YAML scalars; direct typed values still reject padding.
			if tc.name == "padded" {
				config.Work.StateAuthority = WorkStateAuthority(tc.value)
			}

			if got := WorkAuthority(config.Work); got != tc.want || (ValidateWorkConfig(config.Work) == nil) != tc.valid {
				t.Fatalf("authority = %q, validation = %v", got, ValidateWorkConfig(config.Work))
			}
		})
	}
}
