package webui

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const loginIterations = 600_000

// LoginCredentials contains a salted password verifier, never a plaintext password.
type LoginCredentials struct {
	Version  int    `json:"version"`
	Username string `json:"username"`
	Salt     []byte `json:"salt"`
	Hash     []byte `json:"hash"`
}

// CreateLoginCredentials prepares the single account for private local storage.
func CreateLoginCredentials(username, password string) (LoginCredentials, error) {
	credentials := LoginCredentials{Version: 1, Username: username, Salt: make([]byte, 16)}
	if strings.TrimSpace(username) != username || username == "" || len(username) > 80 || strings.ContainsAny(username, "\r\n\t") {
		return LoginCredentials{}, errors.New("username must contain 1-80 characters without surrounding whitespace or control separators")
	}

	if len(password) < 12 || len(password) > 1024 {
		return LoginCredentials{}, errors.New("password must contain 12-1024 bytes")
	}

	if _, err := rand.Read(credentials.Salt); err != nil {
		return LoginCredentials{}, fmt.Errorf("create login salt: %w", err)
	}

	hash, err := pbkdf2.Key(sha256.New, password, credentials.Salt, loginIterations, 32)
	if err != nil {
		return LoginCredentials{}, fmt.Errorf("derive login verifier: %w", err)
	}

	credentials.Hash = hash

	return credentials, nil
}

func (c LoginCredentials) validate() error {
	if c.Version != 1 || c.Username == "" || len(c.Username) > 80 || len(c.Salt) != 16 || len(c.Hash) != 32 {
		return errors.New("invalid web login credentials; run chr web login setup --replace")
	}

	return nil
}

func (c LoginCredentials) matches(username, password string) bool {
	if len(password) > 1024 || len(username) > 80 {
		return false
	}

	hash, err := pbkdf2.Key(sha256.New, password, c.Salt, loginIterations, 32)
	if err != nil {
		return false
	}

	passwordOK := subtle.ConstantTimeCompare(hash, c.Hash)
	usernameOK := subtle.ConstantTimeCompare([]byte(username), []byte(c.Username))

	return passwordOK&usernameOK == 1
}

// LoadLoginCredentials refuses public, linked, oversized or malformed credential files.
func LoadLoginCredentials(path string) (*LoginCredentials, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > 4096 {
		return nil, errors.New("web login file must be a private regular file of at most 4096 bytes")
	}

	data, err := readLoginFile(path)
	if err != nil {
		return nil, err
	}

	var credentials LoginCredentials

	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&credentials); err != nil {
		return nil, errors.New("invalid web login file")
	}

	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected data in web login file")
	}

	return &credentials, credentials.validate()
}

func readLoginFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open web login file: %w", err)
	}

	data, readErr := io.ReadAll(io.LimitReader(file, 4097))

	return data, errors.Join(readErr, file.Close())
}

// SaveLoginCredentials atomically publishes a private verifier, with explicit replacement.
func SaveLoginCredentials(path string, credentials LoginCredentials, replace bool) (err error) {
	if err := credentials.validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create web login directory: %w", err)
	}

	data, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("encode web login credentials: %w", err)
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".web-login-*")
	if err != nil {
		return fmt.Errorf("create private web login file: %w", err)
	}
	defer func() {
		removeErr := os.Remove(file.Name())
		if !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, removeErr)
		}
	}()

	if _, err := file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Close(); err != nil {
		return err
	}

	if replace {
		return os.Rename(file.Name(), path)
	}

	if err := os.Link(file.Name(), path); err != nil {
		return fmt.Errorf("save web login; use --replace to change an existing account: %w", err)
	}

	return nil
}
