package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// The frozen frames protect both artifact kinds, empty/missing/generated states,
// busy messages, selection and every maximized pane while rendering is shared.
func TestDocumentCatalogFrames(t *testing.T) {
	content, err := os.ReadFile("testdata/quality-catalogs/screens.json")
	if err != nil {
		t.Fatal(err)
	}

	var expected map[string]string
	if err := json.Unmarshal(content, &expected); err != nil {
		t.Fatal(err)
	}

	actual := documentCatalogFrames()
	if !reflect.DeepEqual(actual, expected) {
		for name, frame := range actual {
			if expected[name] != frame {
				t.Errorf("frame %s changed:\n%s", name, frame)
			}
		}
	}
}

func documentCatalogFrames() map[string]string {
	frames := make(map[string]string)

	for _, kind := range []View{ViewSpecs, ViewPlans} {
		for _, state := range []string{"empty", "missing", "generated", "busy"} {
			model := NewModel(Options{ProjectDir: "/project"})
			model.selectView(kind)

			if state != "empty" {
				content := ""
				if state != "missing" {
					content = "# Artifact\n\nA retained document.\n\n- Bounded work\n"
				}

				model.specs = []SpecView{{Index: 1, Title: "Example spec", UseCaseTitle: "Example case", Content: content, Status: state, Scope: QuizScope{CaseSlug: "example"}}}
				model.plans = []PlanView{{Index: 1, Title: "Example plan", UseCaseTitle: "Example case", RelativePath: "plans/example.md", Content: content, Status: state, Scope: QuizScope{CaseSlug: "example"}}}
				model.selectedSpec, model.selectedPlan = 1, 1
			}

			model.specBusy, model.planBusy = state == "busy", state == "busy"
			model.statusMessage = "Generating the selected document."
			frames[fmt.Sprintf("%s/%s/plain", kind, state)] = model.View()

			model.width, model.height = 100, 28
			for _, pane := range []int{-1, 0, 1} {
				model.maximizedPane = pane
				frames[fmt.Sprintf("%s/%s/pane=%d", kind, state, pane)] = model.View()
			}
		}
	}

	return frames
}
