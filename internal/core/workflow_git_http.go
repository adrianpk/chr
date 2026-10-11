package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

type workflowReceivePackHTTP struct {
	ctx           context.Context
	client        http.Client
	endpoint      string
	authorization string
	advertisement []byte
}

func (git OSWorkflowGitPublisher) openHTTPReceivePack(ctx context.Context, directory string) (workflowReceivePack, error) {
	remote, err := url.Parse(git.transport)
	if err != nil || !validWorkflowHTTPRemote(remote) {
		return nil, fmt.Errorf("invalid host-owned HTTPS publication transport")
	}

	client := *http.DefaultClient
	if git.httpClient != nil {
		client = *git.httpClient
	}

	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return fmt.Errorf("canonical publication redirects are forbidden")
	}

	session := &workflowReceivePackHTTP{ctx: ctx, client: client, endpoint: strings.TrimSuffix(remote.String(), "/")}
	if err := session.discover(directory); err != nil {
		return nil, err
	}

	return session, nil
}

func validWorkflowHTTPRemote(remote *url.URL) bool {
	if !validWorkflowHTTPIdentity(remote) {
		return false
	}

	for _, part := range strings.Split(remote.Path, "/") {
		if part == ".." || part == "." || strings.ContainsAny(part, "\x00\r\n") {
			return false
		}
	}

	return remote.Path != "" && remote.Path != "/"
}

func validWorkflowHTTPIdentity(remote *url.URL) bool {
	return remote.Scheme == "https" && remote.Host != "" && remote.User == nil && remote.RawQuery == "" && remote.Fragment == "" && remote.RawPath == "" && remote.Opaque == ""
}

func (session *workflowReceivePackHTTP) discover(directory string) error {
	unauthorized, err := session.readHTTPAdvertisement()
	if err != nil || !unauthorized {
		return err
	}

	authorization, err := workflowPublicationCredential(session.ctx, directory, session.endpoint)
	if err != nil {
		return err
	}

	session.authorization = authorization

	unauthorized, err = session.readHTTPAdvertisement()
	if unauthorized {
		return fmt.Errorf("host publication authorization rejected")
	}

	return err
}

func (session *workflowReceivePackHTTP) readHTTPAdvertisement() (unauthorized bool, err error) {
	response, err := session.request(http.MethodGet, "/info/refs?service=git-receive-pack", nil)
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()

	if response.StatusCode == http.StatusUnauthorized {
		return true, nil
	}

	return false, session.readAdvertisement(response)
}

func (session *workflowReceivePackHTTP) request(method, suffix string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(session.ctx, method, session.endpoint+suffix, body)
	if err != nil {
		return nil, fmt.Errorf("HTTPS publication request is invalid")
	}

	request.Header.Set("Cache-Control", "no-cache")

	if session.authorization != "" {
		request.Header.Set("Authorization", session.authorization)
	}

	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/x-git-receive-pack-request")
		request.Header.Set("Accept", "application/x-git-receive-pack-result")
	}

	response, err := session.client.Do(request)
	if err != nil {
		if response != nil {
			return nil, errors.Join(fmt.Errorf("HTTPS publication request failed"), response.Body.Close())
		}

		return nil, fmt.Errorf("HTTPS publication request failed")
	}

	return response, nil
}

func (session *workflowReceivePackHTTP) readAdvertisement(response *http.Response) error {
	if !validWorkflowHTTPResponse(response, "advertisement") {
		return fmt.Errorf("invalid smart HTTPS advertisement response")
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, workflowPublicationTextLimit+1))
	if err != nil || len(data) > workflowPublicationTextLimit {
		return fmt.Errorf("HTTPS advertisement exceeds bounds or is incomplete")
	}

	return session.parseAdvertisement(data)
}

func (session *workflowReceivePackHTTP) parseAdvertisement(data []byte) error {
	reader := bytes.NewReader(data)

	service, err := readWorkflowPublicationPacket(reader)
	if err != nil || strings.TrimSuffix(string(service), "\n") != "# service=git-receive-pack" {
		return fmt.Errorf("invalid HTTPS publication service")
	}

	flush, err := readWorkflowPublicationPacket(reader)
	if err != nil || flush != nil {
		return fmt.Errorf("invalid HTTPS publication service delimiter")
	}

	offset := len(data) - reader.Len()
	if _, err := readWorkflowAdvertisement(reader); err != nil || reader.Len() != 0 {
		return fmt.Errorf("malformed or trailing HTTPS advertisement")
	}

	session.advertisement = data[offset:]

	return nil
}

func validWorkflowHTTPResponse(response *http.Response, kind string) bool {
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))

	return err == nil && response.StatusCode == http.StatusOK && contentType == "application/x-git-receive-pack-"+kind
}

func (session *workflowReceivePackHTTP) Advertisement() io.Reader {
	return bytes.NewReader(session.advertisement)
}

func (session *workflowReceivePackHTTP) Send(data []byte) (report io.Reader, err error) {
	response, err := session.request(http.MethodPost, "/git-receive-pack", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer func() { err = errors.Join(err, response.Body.Close()) }()

	if !validWorkflowHTTPResponse(response, "result") {
		return nil, fmt.Errorf("invalid HTTPS publication report response")
	}

	status, err := io.ReadAll(io.LimitReader(response.Body, workflowPublicationTextLimit+1))
	if err != nil || len(status) > workflowPublicationTextLimit {
		return nil, fmt.Errorf("HTTPS report exceeds bounds or is incomplete")
	}

	return bytes.NewReader(status), nil
}

func (*workflowReceivePackHTTP) Finish() error { return nil }

func (*workflowReceivePackHTTP) Close() error { return nil }
