package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SSHCommitSigning is trusted host configuration. Only the public key
// and principal enter evidence; the private signing path stays on the host.
type SSHCommitSigning struct {
	KeyFile   string
	PublicKey string
	Principal string
}

// Profile returns the nonsecret identity bound by publication capability.
func (signing SSHCommitSigning) Profile() (workflowstate.CommitSigningProfile, error) {
	return workflowSSHSigningProfile(&signing)
}

// WithSSHCommitSigning stores an explicitly supplied trusted host signer.
func (git OSWorkflowPreparationGit) WithSSHCommitSigning(signing SSHCommitSigning) OSWorkflowPreparationGit {
	git.signing = &signing

	return git
}

func workflowSSHSigningProfile(signing *SSHCommitSigning) (workflowstate.CommitSigningProfile, error) {
	if signing == nil || !filepath.IsAbs(signing.KeyFile) {
		return workflowstate.CommitSigningProfile{}, workflowstate.ErrConflict
	}

	p := workflowstate.CommitSigningProfile{PublicKey: strings.TrimSpace(signing.PublicKey), Principal: signing.Principal}

	p.Hash = contenthash.Sum([]byte("chr-ssh-commit-signing/v1:" + p.Principal + ":" + p.PublicKey))
	if !workflowintegrate.ValidWorkflowSigningProfile(p) {
		return p, workflowstate.ErrConflict
	}

	return p, nil
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

func (reader workflowIntegrationReader) verifyPreparationSignature(ctx context.Context, commit string, profile *workflowstate.CommitSigningProfile) (workflowstate.IntegrationSignature, error) {
	if profile == nil || !workflowintegrate.ValidWorkflowSigningProfile(*profile) {
		return workflowstate.IntegrationSignature{}, workflowstate.ErrConflict
	}

	data, err := reader.RawOutput(ctx, "cat-file", "commit", commit)
	if err != nil {
		return workflowstate.IntegrationSignature{}, err
	}

	signedPayload, err := workflowintegrate.WorkflowPreparationSignaturePayload(data)

	payload, signature := signedPayload.Payload, signedPayload.Signature

	if err != nil {
		return workflowstate.IntegrationSignature{}, err
	}

	if err := verifyWorkflowSSHSignature(ctx, payload, signature, *profile); err != nil {
		return workflowstate.IntegrationSignature{}, err
	}

	return workflowstate.IntegrationSignature{Commit: commit, SignatureHash: contenthash.Sum([]byte(signature))}, nil
}

func verifyWorkflowSSHSignature(ctx context.Context, payload, signature string, profile workflowstate.CommitSigningProfile) (err error) {
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
	stdout, stderr := &opsfile.ProviderOutput{}, &opsfile.CommandOutput{}

	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("host commit signature operation failed: %w", err)
	}

	if stdout.Overflow {
		return "", workflowstate.ErrConflict
	}

	return stdout.Buffer.String(), nil
}
