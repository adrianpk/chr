package workflow

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const (
	workflowPublicationPacketLimit = 65520
	workflowPublicationTextLimit   = 1 << 20
	workflowPublicationPackLimit   = 64 << 20
)

type workflowAdvertisement struct {
	refs         map[string]string
	capabilities map[string]bool
}

func workflowPublicationPacket(payload string) []byte {
	return []byte(fmt.Sprintf("%04x%s", len(payload)+4, payload))
}

func readWorkflowPublicationPacket(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, fmt.Errorf("incomplete packet header: %w", err)
	}

	for _, digit := range header {
		if !bytes.ContainsRune([]byte("0123456789abcdef"), rune(digit)) {
			return nil, fmt.Errorf("invalid publication packet header")
		}
	}

	length, err := strconv.ParseUint(string(header[:]), 16, 16)
	if err != nil || length > workflowPublicationPacketLimit || length > 0 && length < 4 {
		return nil, fmt.Errorf("invalid publication packet length")
	}

	if length == 0 {
		return nil, nil
	}

	payload := make([]byte, length-4)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, fmt.Errorf("incomplete packet payload: %w", err)
	}

	return payload, nil
}

func readWorkflowAdvertisement(reader io.Reader) (workflowAdvertisement, error) {
	advertisement := workflowAdvertisement{refs: make(map[string]string), capabilities: make(map[string]bool)}
	bounded := &io.LimitedReader{R: reader, N: workflowPublicationTextLimit}

	for first := true; ; first = false {
		packet, err := readWorkflowPublicationPacket(bounded)
		if err != nil {
			return advertisement, err
		}

		if packet == nil {
			return advertisement, nil
		}

		if err := advertisement.add(packet, first); err != nil {
			return advertisement, err
		}
	}
}

func (advertisement workflowAdvertisement) add(packet []byte, first bool) error {
	line := strings.TrimSuffix(string(packet), "\n")

	ref, capabilities, found := strings.Cut(line, "\x00")
	if found != first || strings.ContainsAny(capabilities, "\x00\r\n") {
		return fmt.Errorf("invalid advertisement capabilities")
	}

	fields := strings.Split(ref, " ")
	if len(fields) != 2 || !workflowstate.ValidCommitID(fields[0]) || !validWorkflowAdvertisedRef(fields[1]) {
		return fmt.Errorf("invalid advertised ref")
	}

	if fields[1] == ".have" {
		return nil
	}

	if _, exists := advertisement.refs[fields[1]]; exists {
		return fmt.Errorf("duplicate advertised ref")
	}

	advertisement.refs[fields[1]] = fields[0]

	if !first {
		return nil
	}

	return advertisement.addCapabilities(capabilities)
}

func validWorkflowAdvertisedRef(ref string) bool {
	if ref == "HEAD" || ref == ".have" || ref == "capabilities^{}" {
		return true
	}

	return strings.HasPrefix(ref, "refs/") && workflowstate.ValidBranch(strings.TrimPrefix(ref, "refs/"))
}

func (advertisement workflowAdvertisement) addCapabilities(capabilities string) error {
	for _, capability := range strings.Split(capabilities, " ") {
		if !validWorkflowPublicationCapability(capability) || advertisement.capabilities[capability] {
			return fmt.Errorf("duplicate advertisement capability")
		}

		advertisement.capabilities[capability] = true
	}

	return nil
}

func validWorkflowPublicationCapability(capability string) bool {
	if capability == "" || len(capability) > 256 {
		return false
	}

	for _, character := range capability {
		if character < 33 || character > 126 {
			return false
		}
	}

	return true
}

func (advertisement workflowAdvertisement) check(request GitPublication) error {
	if !advertisement.capabilities["atomic"] || !advertisement.capabilities["report-status"] {
		return workflowstate.ErrWorkflowAtomicGuard
	}

	if len(request.Git.ResultCommit) == 64 && !advertisement.capabilities["object-format=sha256"] {
		return workflowstate.ErrWorkflowAtomicGuard
	}

	source, target := advertisement.refs[request.SourceRef], advertisement.refs[request.TargetRef]
	if source == request.Git.ResultCommit || target == request.Git.ResultCommit {
		return ErrWorkflowObservationRequired
	}

	if source != request.Git.SourceCommit || target != request.Git.BaseCommit {
		return workflowstate.ErrConflict
	}

	return nil
}

func readWorkflowPublicationReport(reader io.Reader, request GitPublication) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, workflowPublicationTextLimit+1))
	if err != nil || len(data) > workflowPublicationTextLimit {
		return "unknown", fmt.Errorf("publication report exceeds bounds or is incomplete")
	}

	packets := bytes.NewReader(data)

	unpack, err := readWorkflowPublicationPacket(packets)
	if err != nil || !bytes.HasPrefix(unpack, []byte("unpack ")) {
		return "unknown", fmt.Errorf("missing unpack status")
	}

	statuses, err := workflowPublicationRefStatuses(packets, request)
	if err != nil {
		return "unknown", err
	}

	return classifyWorkflowPublicationReport(string(unpack), statuses)
}

func classifyWorkflowPublicationReport(unpack string, statuses [2]string) (string, error) {
	if statuses[0] == "ng" && statuses[1] == "ng" {
		return "rejected", nil
	}

	if strings.TrimSuffix(unpack, "\n") != "unpack ok" || statuses[0] != "ok" || statuses[1] != "ok" {
		return "unknown", fmt.Errorf("inconsistent atomic publication report")
	}

	return "applied", nil
}

func workflowPublicationRefStatuses(reader *bytes.Reader, request GitPublication) ([2]string, error) {
	var statuses [2]string

	refs := []string{request.SourceRef, request.TargetRef}

	for range 2 {
		packet, err := readWorkflowPublicationPacket(reader)
		if err != nil || packet == nil {
			return statuses, fmt.Errorf("missing ref status")
		}

		if err := workflowPublicationRefStatus(string(packet), refs, &statuses); err != nil {
			return statuses, err
		}
	}

	if err := workflowPublicationReportEnd(reader); err != nil {
		return statuses, err
	}

	return statuses, nil
}

func workflowPublicationReportEnd(reader *bytes.Reader) error {
	flush, err := readWorkflowPublicationPacket(reader)
	if err != nil || flush != nil {
		return fmt.Errorf("incomplete or trailing ref status")
	}

	// GitHub's receive-pack appends one empty transport flush to the report
	// flush. Both refs still require exactly one explicit status; no payload,
	// partial packet or further terminator is accepted after those statuses.
	if reader.Len() != 0 {
		flush, err = readWorkflowPublicationPacket(reader)
		if err != nil || flush != nil || reader.Len() != 0 {
			return fmt.Errorf("trailing publication report data")
		}
	}

	return nil
}

func workflowPublicationRefStatus(packet string, refs []string, statuses *[2]string) error {
	fields := strings.Split(strings.TrimSuffix(packet, "\n"), " ")
	if !validWorkflowPublicationStatus(fields) {
		return fmt.Errorf("invalid ref status")
	}

	for index, ref := range refs {
		if fields[1] == ref && statuses[index] == "" {
			statuses[index] = fields[0]

			return nil
		}
	}

	return fmt.Errorf("unknown or duplicate ref status")
}

func validWorkflowPublicationStatus(fields []string) bool {
	if len(fields) < 2 {
		return false
	}

	switch fields[0] {
	case "ok":
		return len(fields) == 2
	case "ng":
		return len(fields) >= 3
	default:
		return false
	}
}
