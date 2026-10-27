package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowSSHCommitSigning is trusted host configuration. Only the public key
// and principal enter evidence; the private signing path stays on the host.
type WorkflowSSHCommitSigning struct {
	KeyFile   string
	PublicKey string
	Principal string
}

// WorkflowCommitSigningProfile identifies the certified SSH signer policy.
type WorkflowCommitSigningProfile struct {
	Hash      string `json:"hash"`
	PublicKey string `json:"public_key"`
	Principal string `json:"principal"`
}

// WorkflowIntegrationSignature binds cryptographic verification to one object.
type WorkflowIntegrationSignature struct {
	Commit        string `json:"commit"`
	SignatureHash string `json:"signature_hash"`
}

// Profile returns the nonsecret identity bound by publication capability.
func (signing WorkflowSSHCommitSigning) Profile() (WorkflowCommitSigningProfile, error) {
	return workflowSSHSigningProfile(&signing)
}

// WithSSHCommitSigning stores an explicitly supplied trusted host signer.
func (git OSWorkflowPreparationGit) WithSSHCommitSigning(signing WorkflowSSHCommitSigning) OSWorkflowPreparationGit {
	git.signing = &signing

	return git
}

func workflowSSHSigningProfile(signing *WorkflowSSHCommitSigning) (WorkflowCommitSigningProfile, error) {
	if signing == nil || !filepath.IsAbs(signing.KeyFile) {
		return WorkflowCommitSigningProfile{}, workflowstate.ErrConflict
	}

	p := WorkflowCommitSigningProfile{PublicKey: strings.TrimSpace(signing.PublicKey), Principal: signing.Principal}

	p.Hash = contenthash.Sum([]byte("chr-ssh-commit-signing/v1:" + p.Principal + ":" + p.PublicKey))
	if !validWorkflowSigningProfile(p) {
		return p, workflowstate.ErrConflict
	}

	return p, nil
}

func validWorkflowSigningProfile(p WorkflowCommitSigningProfile) bool {
	if !validWorkflowSigningPrincipal(p.Principal) || !validWorkflowSigningPublicKey(p.PublicKey) {
		return false
	}

	return p.Hash == contenthash.Sum([]byte("chr-ssh-commit-signing/v1:"+p.Principal+":"+p.PublicKey))
}

func validWorkflowSigningPrincipal(principal string) bool {
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

func validWorkflowSigningPublicKey(publicKey string) bool {
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

func validWorkflowPreparationSignatures(proof WorkflowIntegrationProof, p WorkflowIntegrationPreparation) bool {
	if !p.RequiredSignatures {
		return p.Signing == nil && len(p.Signatures) == 0
	}

	if p.Signing == nil || !validWorkflowSigningProfile(*p.Signing) || len(p.Signatures) != len(proof.Provenance.Prepared) {
		return false
	}

	for i, result := range p.Signatures {
		if result.Commit != proof.Provenance.Prepared[i].Commit || !validWorkflowNativeFingerprint(result.SignatureHash) {
			return false
		}
	}

	return true
}

func (reader workflowIntegrationReader) signPreparationObject(ctx context.Context, data string) (string, error) {
	if reader.signing == nil {
		return data, nil
	}

	profile, err := workflowSSHSigningProfile(reader.signing)
	if err != nil {
		return "", err
	}

	if reader.preparation != nil && (!reader.preparation.RequiredSignatures || reader.preparation.Signing == nil || reader.preparation.Signing.Hash != profile.Hash) {
		return "", workflowstate.ErrConflict
	}

	signature, err := workflowCommitSSHCommand(ctx, data, "-Y", "sign", "-n", "git", "-f", reader.signing.KeyFile)
	if err != nil {
		return "", err
	}

	header, message, ok := strings.Cut(data, "\n\n")
	if !ok || !strings.HasPrefix(signature, "-----BEGIN SSH SIGNATURE-----\n") {
		return "", workflowstate.ErrConflict
	}

	signature = strings.TrimSuffix(signature, "\n")

	return header + "\ngpgsig " + strings.ReplaceAll(signature, "\n", "\n ") + "\n\n" + message, nil
}

func (reader workflowIntegrationReader) verifyPreparationSignature(ctx context.Context, commit string, profile *WorkflowCommitSigningProfile) (WorkflowIntegrationSignature, error) {
	if profile == nil || !validWorkflowSigningProfile(*profile) {
		return WorkflowIntegrationSignature{}, workflowstate.ErrConflict
	}

	data, err := reader.rawOutput(ctx, "cat-file", "commit", commit)
	if err != nil {
		return WorkflowIntegrationSignature{}, err
	}

	payload, signature, err := workflowPreparationSignaturePayload(data)
	if err != nil {
		return WorkflowIntegrationSignature{}, err
	}

	if err := verifyWorkflowSSHSignature(ctx, payload, signature, *profile); err != nil {
		return WorkflowIntegrationSignature{}, err
	}

	return WorkflowIntegrationSignature{Commit: commit, SignatureHash: contenthash.Sum([]byte(signature))}, nil
}

func workflowPreparationSignaturePayload(data string) (string, string, error) {
	header, message, ok := strings.Cut(data, "\n\n")
	if !ok {
		return "", "", workflowstate.ErrConflict
	}

	var unsigned, signature []string

	inside := false

	for _, line := range strings.Split(header, "\n") {
		switch {
		case strings.HasPrefix(line, "gpgsig ") && len(signature) != 0:
			return "", "", workflowstate.ErrConflict
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
		return "", "", fmt.Errorf("prepared commit lacks a required signature")
	}

	return strings.Join(unsigned, "\n") + "\n\n" + message, strings.Join(signature, "\n") + "\n", nil
}

func verifyWorkflowSSHSignature(ctx context.Context, payload, signature string, profile WorkflowCommitSigningProfile) (err error) {
	dir, err := os.MkdirTemp("", "chr-commit-signature-")
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()

	allowed, sig := filepath.Join(dir, "allowed"), filepath.Join(dir, "signature")
	if err := os.WriteFile(allowed, []byte(profile.Principal+" namespaces=\"git\" "+profile.PublicKey+"\n"), 0o600); err != nil {
		return err
	}

	if err := os.WriteFile(sig, []byte(signature), 0o600); err != nil {
		return err
	}

	_, err = workflowCommitSSHCommand(ctx, payload, "-Y", "verify", "-n", "git", "-I", profile.Principal, "-f", allowed, "-s", sig)

	return err
}

func workflowCommitSSHCommand(ctx context.Context, input string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "ssh-keygen", args...)
	command.Stdin = strings.NewReader(input)
	stdout, stderr := &workflowProviderOutput{}, &workflowCommandOutput{}

	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("host commit signature operation failed: %w", err)
	}

	if stdout.overflow {
		return "", workflowstate.ErrConflict
	}

	return stdout.buffer.String(), nil
}
