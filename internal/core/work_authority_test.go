package core

import (
	"testing"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestWorkAuthority keeps absence transitional and validates explicit routing.
func TestWorkAuthority(t *testing.T) {
	cases := []workAuthorityCase{
		{"absent", "", workspaceconfig.WorkStateAuthorityTrackerCompat, true, true},
		{"empty", "", "", false, false},
		{"tickets", "tickets", workspaceconfig.WorkStateAuthorityTickets, true, true},
		{"compat", "tracker_compat", workspaceconfig.WorkStateAuthorityTrackerCompat, true, true},
		{"unknown", "both", "", false, false},
		{"padded", " tickets ", " tickets ", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkWorkAuthority(t, tc) })
	}
}

type workAuthorityCase struct {
	name, value string
	want        workspaceconfig.WorkStateAuthority
	valid       bool
	loadValid   bool
}

func checkWorkAuthority(t *testing.T, tc workAuthorityCase) {
	t.Helper()

	project := t.TempDir()
	if tc.name != "absent" {
		writeWorkspaceConfig(t, project, "work:\n  state_authority: "+tc.value+"\n")
	}

	config, err := workspaceconfig.Load(project)
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
		config.Work.StateAuthority = workspaceconfig.WorkStateAuthority(tc.value)
	}

	if got := workspaceconfig.WorkAuthority(config.Work); got != tc.want || (workspaceconfig.ValidateWorkConfig(config.Work) == nil) != tc.valid {
		t.Fatalf("authority = %q, validation = %v", got, workspaceconfig.ValidateWorkConfig(config.Work))
	}
}
