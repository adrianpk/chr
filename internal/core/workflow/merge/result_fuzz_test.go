package merge

import (
	"errors"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// FuzzNativeMergerJSON exercises untrusted structured output within the file
// and nesting limits; malformed data must fail without panics or authority.
func FuzzNativeMergerJSON(f *testing.F) {
	for _, seed := range []string{`{"protocol_version":1,"outcome":"failed","failure":"Unresolved"}`, `{"outcome":"failed","outcome":"resolved"}`, `{"usage":{"input":9223372036854775808}}`, strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34), string([]byte{0xff})} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var result workflowstate.MergerResult
		if err := decodeWorkflowMergerResult(data, &result); err != nil && !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal("untrusted decoder escaped its error contract", err)
		}
	})
}
