package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

type webWorkflowBrowser struct {
	input, output *os.File
	reader        *bufio.Reader
	sequence      int
	session       string
	loaded        bool
}

func newWebWorkflowBrowser(t *testing.T, path string) *webWorkflowBrowser {
	t.Helper()

	browserInput, writer := webWorkflowBrowserPipe(t)
	reader, browserOutput := webWorkflowBrowserPipe(t)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	cmd := exec.CommandContext(ctx, path, "--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage", "--remote-debugging-pipe", "--user-data-dir="+t.TempDir(), "about:blank")
	cmd.ExtraFiles = []*os.File{browserInput, browserOutput}

	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cancel()

		if err := cmd.Wait(); err != nil && ctx.Err() == nil {
			t.Error(err)
		} // Owned browser checks have finished.
	})

	browser := &webWorkflowBrowser{input: writer, output: reader, reader: bufio.NewReader(reader)}
	result := browser.call(t, "Target.createTarget", map[string]any{"url": "about:blank"})

	var target struct{ TargetID string }
	if err := json.Unmarshal(result, &target); err != nil {
		t.Fatal(err)
	}

	result = browser.call(t, "Target.attachToTarget", map[string]any{"targetId": target.TargetID, "flatten": true})

	var session struct{ SessionID string }
	if err := json.Unmarshal(result, &session); err != nil {
		t.Fatal(err)
	}

	browser.session = session.SessionID
	browser.call(t, "Page.enable", map[string]any{})

	return browser
}

func (b *webWorkflowBrowser) callResult(t *testing.T, method string, params map[string]any) (json.RawMessage, error) {
	t.Helper()

	b.sequence++

	message := map[string]any{"id": b.sequence, "method": method, "params": params}
	if b.session != "" {
		message["sessionId"] = b.session
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := b.input.Write(append(data, 0)); err != nil {
		t.Fatal(err)
	}

	if err := b.output.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		t.Fatal(err)
	}

	for range 256 {
		response := b.receive(t)

		b.loaded = b.loaded || response.Method == "Page.loadEventFired"

		if response.ID != b.sequence {
			continue
		}

		if len(response.Error) != 0 {
			return nil, fmt.Errorf("%s: %s", method, response.Error)
		}

		return response.Result, nil
	}

	t.Fatal("browser event bound exceeded")

	return nil, fmt.Errorf("browser event bounds exceeded")
}

func webWorkflowBrowserFrame(reader *bufio.Reader) ([]byte, error) {
	var frame []byte
	for len(frame) <= 1<<20 {
		chunk, err := reader.ReadSlice(0)
		frame = append(frame, chunk...)

		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}

		if err != nil {
			return nil, err
		}

		return frame[:len(frame)-1], nil
	}

	return nil, fmt.Errorf("browser response bounds exceeded")
}

func (b *webWorkflowBrowser) evaluate(t *testing.T, expression string) string {
	t.Helper()

	result := b.readEvaluation(t, expression)

	var response struct {
		Result           struct{ Value json.RawMessage }
		ExceptionDetails json.RawMessage
	}
	if err := json.Unmarshal(result, &response); err != nil {
		t.Fatal(err)
	}

	if len(response.ExceptionDetails) != 0 {
		t.Fatal("browser script", string(response.ExceptionDetails))
	}

	var value string
	if err := json.Unmarshal(response.Result.Value, &value); err != nil {
		t.Fatal("browser value", string(result), err)
	}

	return value
}

func (b *webWorkflowBrowser) navigate(t *testing.T, location string) {
	t.Helper()

	b.loaded = false
	b.call(t, "Page.navigate", map[string]any{"url": location})
	b.waitLoaded(t)
}

func (b *webWorkflowBrowser) waitLoaded(t *testing.T) {
	t.Helper()

	if err := b.output.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		t.Fatal(err)
	}

	for range 256 {
		if b.loaded {
			return
		}

		frame, err := webWorkflowBrowserFrame(b.reader)
		if err != nil {
			t.Fatal(err)
		}

		var event struct{ Method string }
		if err := json.Unmarshal(frame, &event); err != nil {
			t.Fatal(err)
		}

		b.loaded = event.Method == "Page.loadEventFired"
	}

	t.Fatal("browser navigation event bounds exceeded")
}

func (b *webWorkflowBrowser) clickNavigation(t *testing.T, selector string) {
	t.Helper()

	b.loaded = false
	b.click(t, selector)
	b.waitLoaded(t)
}

func (b *webWorkflowBrowser) wait(t *testing.T, selector string) {
	t.Helper()

	escaped := strconv.Quote(selector)
	b.evaluate(t, `(async()=>{for(let n=0;n<200;n++){if(document.querySelector(`+escaped+`))return 'ready';await new Promise(r=>setTimeout(r,50));}throw Error('selector timeout '+`+escaped+`)})()`)
}

func (b *webWorkflowBrowser) click(t *testing.T, selector string) {
	t.Helper()
	b.wait(t, selector)

	escaped := strconv.Quote(selector)
	b.call(t, "Runtime.evaluate", map[string]any{"expression": `setTimeout(()=>document.querySelector(` + escaped + `).click(), 0); 'clicked'`, "returnByValue": true})
}

func webWorkflowBrowserPipe(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		for _, file := range []*os.File{reader, writer} {
			if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				t.Error(err)
			}
		}
	})

	return reader, writer
}

type webWorkflowBrowserResponse struct {
	ID     int
	Method string
	Result json.RawMessage
	Error  json.RawMessage
}

func (b *webWorkflowBrowser) receive(t *testing.T) webWorkflowBrowserResponse {
	t.Helper()

	frame, err := webWorkflowBrowserFrame(b.reader)
	if err != nil {
		t.Fatal(err)
	}

	var response webWorkflowBrowserResponse
	if err := json.Unmarshal(frame, &response); err != nil {
		t.Fatal(err)
	}

	return response
}

func (b *webWorkflowBrowser) call(t *testing.T, method string, params map[string]any) json.RawMessage {
	t.Helper()

	result, err := b.callResult(t, method, params)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func (b *webWorkflowBrowser) readEvaluation(t *testing.T, expression string) json.RawMessage {
	t.Helper()
	// Only read-only DOM observations may be repeated after a navigation destroys
	// their JavaScript context. Browser form submissions are never retried here.
	for range 3 {
		result, err := b.callResult(t, "Runtime.evaluate", map[string]any{"expression": expression, "awaitPromise": true, "returnByValue": true})
		if err == nil {
			return result
		}

		if !strings.Contains(err.Error(), "navigated or closed") && !strings.Contains(err.Error(), "context") {
			t.Fatal(err)
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("browser navigation did not settle")

	return nil
}
