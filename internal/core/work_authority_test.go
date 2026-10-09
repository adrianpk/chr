package core

import (
	"testing"
)

// TestWorkAuthority keeps absence transitional and validates explicit routing.
func TestWorkAuthority(t *testing.T) {
	cases := []workAuthorityCase{
		{"absent", "", WorkStateAuthorityTrackerCompat, true, true},
		{"empty", "", "", false, false},
		{"tickets", "tickets", WorkStateAuthorityTickets, true, true},
		{"compat", "tracker_compat", WorkStateAuthorityTrackerCompat, true, true},
		{"unknown", "both", "", false, false},
		{"padded", " tickets ", " tickets ", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkWorkAuthority(t, tc) })
	}
}

type workAuthorityCase struct {
	name, value string
	want        WorkStateAuthority
	valid       bool
	loadValid   bool
}

func checkWorkAuthority(t *testing.T, tc workAuthorityCase) {
	t.Helper()

	project := t.TempDir()
	if tc.name != "absent" {
		writeWorkspaceConfig(t, project, "work:\n  state_authority: "+tc.value+"\n")
	}

	config, err := LoadWorkspaceConfig(project)
	if !tc.loadValid {
		if err == nil {
			t.Fatal("invalid authority used a fallback")
		}

		return
	}

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
}
