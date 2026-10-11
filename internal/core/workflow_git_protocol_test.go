package core

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

// TestWorkflowGitPackets refuses malformed and oversized wire frames before
// their contents can become atomic capability or success evidence.
func TestWorkflowGitPackets(t *testing.T) {
	for _, data := range []string{"", "000", "+004", "gggg", "0001", "0002", "ffff", "0008abc"} {
		t.Run(data, func(t *testing.T) {
			if _, err := readWorkflowPublicationPacket(strings.NewReader(data)); err == nil {
				t.Fatal("malformed packet accepted")
			}
		})
	}
}

// TestWorkflowGitMalformedAdvertisement rejects invalid capability framing,
// duplicate refs/capabilities and trailing partial packets with bounded reads.
func TestWorkflowGitMalformedAdvertisement(t *testing.T) {
	id := strings.Repeat("a", 40)
	first := workflowPublicationPacket(id + " refs/heads/source\x00report-status atomic\n")

	cases := []struct{ name, data string }{
		{"no_capabilities", string(workflowPublicationPacket(id+" refs/heads/source\n")) + "0000"},
		{"duplicate_capability", string(workflowPublicationPacket(id+" refs/heads/source\x00atomic atomic\n")) + "0000"},
		{"control_capability", string(workflowPublicationPacket(id+" refs/heads/source\x00atomic\treport-status\n")) + "0000"},
		{"duplicate_ref", string(first) + string(workflowPublicationPacket(id+" refs/heads/source\n")) + "0000"},
		{"extra_capabilities", string(first) + string(first) + "0000"},
		{"invalid_ref", string(first) + string(workflowPublicationPacket(id+" refs/heads/*\n")) + "0000"},
		{"partial", string(first) + "000"},
		{"bounded", workflowOversizedAdvertisement(id)},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readWorkflowAdvertisement(strings.NewReader(test.data)); err == nil {
				t.Fatal("malformed advertisement accepted")
			}
		})
	}
}

func workflowOversizedAdvertisement(id string) string {
	var output bytes.Buffer
	output.Write(workflowPublicationPacket(id + " refs/heads/source\x00report-status atomic\n"))

	for index := 0; output.Len() <= workflowPublicationTextLimit; index++ {
		output.Write(workflowPublicationPacket(id + " refs/heads/filler-" + strconv.Itoa(index) + "\n"))
	}

	output.WriteString("0000")

	return output.String()
}

// TestWorkflowGitReports requires complete matching results for both refs and
// retains uncertainty for missing, mixed, duplicate, extra or malformed status.
func TestWorkflowGitReports(t *testing.T) {
	request := WorkflowGitPublication{SourceRef: "refs/heads/source", TargetRef: "refs/heads/dev"}
	packet := func(line string) string { return string(workflowPublicationPacket(line)) }
	ok := packet("unpack ok\n") + packet("ok "+request.SourceRef+"\n") + packet("ok "+request.TargetRef+"\n") + "0000"

	cases := []struct{ name, data, state string }{
		{"success", ok, "applied"},
		{"without_lf", packet("unpack ok") + packet("ok "+request.SourceRef) + packet("ok "+request.TargetRef) + "0000", "applied"},
		{"rejected", packet("unpack ok\n") + packet("ng "+request.SourceRef+" denied\n") + packet("ng "+request.TargetRef+" atomic failure\n") + "0000", "rejected"},
		{"partial", ok[:len(ok)-1], "unknown"},
		{"missing_ref", packet("unpack ok\n") + packet("ok "+request.SourceRef+"\n") + "0000", "unknown"},
		{"duplicate", packet("unpack ok\n") + packet("ok "+request.SourceRef+"\n") + packet("ok "+request.SourceRef+"\n") + "0000", "unknown"},
		{"github_transport_flush", ok + "0000", "applied"},
		{"extra_flushes", ok + "00000000", "unknown"},
		{"partial_transport_flush", ok + "000", "unknown"},
		{"trailing_status", ok + packet("ok "+request.SourceRef+"\n"), "unknown"},
		{"mixed", packet("unpack ok\n") + packet("ng "+request.SourceRef+" denied\n") + packet("ok "+request.TargetRef+"\n") + "0000", "unknown"},
		{"bad_unpack", packet("unpack corrupt\n") + packet("ok "+request.SourceRef+"\n") + packet("ok "+request.TargetRef+"\n") + "0000", "unknown"},
		{"oversized", strings.Repeat("x", workflowPublicationTextLimit+1), "unknown"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			state, err := readWorkflowPublicationReport(bytes.NewBufferString(test.data), request)
			if state != test.state || (err != nil) != (test.state == "unknown") {
				t.Fatalf("report = %s, %v", state, err)
			}
		})
	}
}
