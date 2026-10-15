package core

import (
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
)

// TestNativeTransportLoss drops the real HTTPS result only after the remote
// processes one publication; observation reconciles P without another POST.
func TestNativeTransportLoss(t *testing.T) {
	if os.Getenv("CHR_NATIVE_INTEGRATION_QA") != "1" {
		t.Skip("live certification requires explicit fixture opt-in")
	}

	state := nativeLiveState(t)
	for _, profile := range readNativeLiveProfiles(t, state) {
		t.Run(profile.Forge, func(t *testing.T) { nativeTransportLoss(t, state, profile) })
	}
}

type nativeLostHTTPReport struct {
	posts   int
	dropped bool
}

func (fault *nativeLostHTTPReport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if request.Method != http.MethodPost || err != nil {
		return response, err
	}

	fault.posts++

	if response.StatusCode != http.StatusOK {
		return response, nil
	}

	_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, workflowPublicationTextLimit+1))

	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		return nil, errors.New("fixture failed while retaining the actual publication reply")
	}

	fault.dropped = true

	return nil, io.ErrUnexpectedEOF
}

func nativeTransportLoss(t *testing.T, state string, profile nativeLiveProfile) {
	t.Helper()

	f := newNativeFaultCase(t, state, profile, "lost_https_report")
	f.admit(t)

	fault := &nativeLostHTTPReport{}
	result, err := f.publisher.WithHTTPClient(&http.Client{Transport: fault}).PublishWorkflowGit(t.Context(), f.request)

	f.evidence.Native.Binding.Publication, f.evidence.Posts = result, fault.posts
	if !errors.Is(err, ErrWorkflowObservationRequired) || result.State != "unknown" || !result.CommandsSent || !fault.dropped || fault.posts != 1 {
		t.Fatalf("lost reply did not retain one unknown publication: %s, posts %d, %v", result.State, fault.posts, err)
	}

	f.checkRefs(t, f.request.Git.ResultCommit, f.request.Git.ResultCommit)
	observeNativeLiveRebase(t, f.adapter, profile, &f.evidence.Native)

	if fault.posts != 1 {
		t.Fatal("observation resubmitted the unknown publication")
	}

	f.checkRefs(t, f.request.Git.ResultCommit, f.request.Git.ResultCommit)
}
