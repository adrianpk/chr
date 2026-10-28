package integrate

import (
	"encoding/base64"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidSigningProfile checks the retained SSH key/principal checksum without using the private key.
func ValidSigningProfile(p workflowstate.CommitSigningProfile) bool {
	if !validSigningPrincipal(p.Principal) || !validSigningPublicKey(p.PublicKey) {
		return false
	}

	return p.Hash == contenthash.Sum([]byte("chr-ssh-commit-signing/v1:"+p.Principal+":"+p.PublicKey))
}

func validSigningPrincipal(principal string) bool {
	if principal == "" || len(principal) > 128 {
		return false
	}

	for _, character := range principal {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._@-", character) {
			return false
		}
	}

	return true
}

func validSigningPublicKey(publicKey string) bool {
	if len(publicKey) > 128 {
		return false
	}

	fields := strings.Fields(publicKey)
	if len(fields) != 2 || fields[0] != "ssh-ed25519" {
		return false
	}

	key, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil || len(key) != 51 || string(key[:19]) != "\x00\x00\x00\x0bssh-ed25519\x00\x00\x00\x20" {
		return false
	}

	return publicKey == strings.Join(fields, " ")
}

func validSignatures(proof workflowstate.IntegrationProof, p workflowstate.IntegrationPreparation) bool {
	if !p.RequiredSignatures {
		return p.Signing == nil && len(p.Signatures) == 0
	}

	if p.Signing == nil || !ValidSigningProfile(*p.Signing) || len(p.Signatures) != len(proof.Provenance.Prepared) {
		return false
	}

	for i, result := range p.Signatures {
		if result.Commit != proof.Provenance.Prepared[i].Commit || !workflowstate.ValidFingerprint(result.SignatureHash) {
			return false
		}
	}

	return true
}

// SignaturePayload extracts exact unsigned bytes and one required SSH signature.
func SignaturePayload(data string) (SignedPayload, error) {
	header, message, ok := strings.Cut(data, "\n\n")
	if !ok {
		return SignedPayload{Payload: "", Signature: ""}, workflowstate.ErrConflict
	}

	var unsigned, signature []string

	inside := false

	for _, line := range strings.Split(header, "\n") {
		switch {
		case strings.HasPrefix(line, "gpgsig ") && len(signature) != 0:
			return SignedPayload{Payload: "", Signature: ""}, workflowstate.ErrConflict
		case strings.HasPrefix(line, "gpgsig "):
			inside = true

			signature = append(signature, strings.TrimPrefix(line, "gpgsig "))
		case inside && strings.HasPrefix(line, " "):
			signature = append(signature, strings.TrimPrefix(line, " "))
		default:
			inside = false

			unsigned = append(unsigned, line)
		}
	}

	if len(signature) == 0 {
		return SignedPayload{Payload: "", Signature: ""}, fmt.Errorf("prepared commit lacks a required signature")
	}

	return SignedPayload{Payload: strings.Join(unsigned, "\n") + "\n\n" + message, Signature: strings.Join(signature, "\n") + "\n"}, nil
}

// SignedPayload retains exact unsigned bytes and the extracted SSH signature.
type SignedPayload struct{ Payload, Signature string }
