package core

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

func (git OSWorkflowGitPublisher) openSSHReceivePack(ctx context.Context) (workflowReceivePack, error) {
	args, err := git.publicationSSHArgs()
	if err != nil {
		return nil, err
	}

	owned, cancel := context.WithCancel(ctx)
	command := workflowPublicationCommand(owned, "ssh", args...)

	return startWorkflowReceivePack(command, cancel)
}

func (git OSWorkflowGitPublisher) publicationSSHArgs() ([]string, error) {
	remote, err := url.Parse(git.transport)
	if err != nil || !validWorkflowSSHRemote(remote) {
		return nil, fmt.Errorf("invalid host-owned SSH publication transport")
	}

	args := []string{"-T", "-o", "BatchMode=yes"}

	if git.sshConfig != "" {
		if !filepath.IsAbs(git.sshConfig) {
			return nil, fmt.Errorf("SSH configuration must use an absolute host-owned path")
		}

		args = append(args, "-F", git.sshConfig)
	}

	if remote.Port() != "" {
		args = append(args, "-p", remote.Port())
	}

	destination := remote.Hostname()
	if remote.User != nil {
		destination = remote.User.Username() + "@" + destination
	}

	// SSH executes one remote command through the server's login shell. Quote the
	// repository as one POSIX shell word; never splice refs or credentials into it.
	path := "'" + strings.ReplaceAll(remote.Path, "'", "'\\''") + "'"
	args = append(args, "--", destination, "git-receive-pack "+path)

	return args, nil
}

func validWorkflowSSHRemote(remote *url.URL) bool {
	if !validWorkflowSSHLocation(remote) {
		return false
	}

	if !validWorkflowSSHHost(remote.Hostname()) || !validWorkflowSSHUser(remote.User) {
		return false
	}

	if remote.Port() != "" {
		port, err := strconv.Atoi(remote.Port())
		if err != nil || port < 1 || port > 65535 {
			return false
		}
	}

	return true
}

func validWorkflowSSHLocation(remote *url.URL) bool {
	return remote.Scheme == "ssh" && remote.RawQuery == "" && remote.Fragment == "" && remote.Opaque == "" && strings.HasPrefix(remote.Path, "/") && !strings.ContainsAny(remote.Path, "\x00\r\n")
}

func validWorkflowSSHHost(host string) bool {
	if net.ParseIP(host) != nil {
		return true
	}

	return validWorkflowSSHName(host)
}

func validWorkflowSSHUser(user *url.Userinfo) bool {
	if user == nil {
		return true
	}

	if _, password := user.Password(); password {
		return false
	}

	return validWorkflowSSHName(user.Username())
}

func validWorkflowSSHName(value string) bool {
	if value == "" || strings.HasPrefix(value, "-") || len(value) > 255 {
		return false
	}

	for _, character := range value {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-", character) {
			continue
		}

		return false
	}

	return true
}
