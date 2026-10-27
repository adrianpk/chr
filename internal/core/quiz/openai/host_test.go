package openai

import (
	"context"
	"strings"
	"testing"
)

// A missing host must fail before credentials, HTTP or retained generation can run.
func TestHostRequired(t *testing.T) {
	_, err := (Adapter{APIKey: "unused-secret"}).Invoke(context.Background(), validQuizModelRequest())
	if err == nil || !strings.Contains(err.Error(), "host is required") {
		t.Fatalf("missing host: %v", err)
	}
}
