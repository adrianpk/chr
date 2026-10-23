package app

import (
	"net/url"
	"testing"
)

// Compare the actual manual and guided screens at the same viewport. Hidden
// workflow bindings may differ; typography, geometry and visible controls may not.
func assertSharedQuizAppearance(t *testing.T, browser *webWorkflowBrowser, serverURL string) {
	t.Helper()

	location := browser.evaluate(t, "location.href")
	quizID := browser.evaluate(t, `document.querySelector('.quiz-response [name=id]').value`)
	guided := quizAppearance(t, browser)
	browser.navigate(t, serverURL+quizFocusURL(quizID))
	browser.wait(t, ".quiz-response textarea")

	manual := quizAppearance(t, browser)
	if guided != manual {
		t.Fatalf("quiz appearance differs: guided %s; manual %s", guided, manual)
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatal(err)
	}

	browser.navigate(t, parsed.String())
	browser.wait(t, ".quiz-response textarea")
}

func quizAppearance(t *testing.T, browser *webWorkflowBrowser) string {
	t.Helper()

	return browser.evaluate(t, `JSON.stringify(['.quiz-heading h1','.quiz-mode-toggle','.quiz-panel','#quiz-answer','.quiz-actions'].map(selector=>{
const element=document.querySelector(selector), rect=element.getBoundingClientRect(), style=getComputedStyle(element);
return {selector,x:rect.x,y:rect.y,width:rect.width,height:rect.height,padding:style.padding,font:style.font,color:style.color,background:style.backgroundColor,controls:Array.from(element.querySelectorAll('button,label')).map(button=>({text:button.textContent,color:getComputedStyle(button).color,background:getComputedStyle(button).backgroundColor,padding:getComputedStyle(button).padding}))};
}))`)
}
