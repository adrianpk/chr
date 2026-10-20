package pi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"forge.adrianpk.com/chr/internal/harness"
)

const resourceLimit = 65_536

// ResolveResources validates explicitly selected files without spawning Pi.
// Credential values are never hashed; model definitions must be credential-free.
func ResolveResources(configDir, credentialRef, modelRef, provider string) (harness.GenerationResources, error) {
	var resources harness.GenerationResources
	if credentialRef == "" {
		return resources, fmt.Errorf("configure a private static API-key credential reference for Pi generation: %w", ErrCapability)
	}

	resources.CredentialFile = resourcePath(configDir, credentialRef)
	if _, _, err := selectedCredential(resources.CredentialFile, provider); err != nil {
		return resources, err
	}

	if modelRef != "" {
		resources.ModelConfigFile = resourcePath(configDir, modelRef)

		models, err := selectedModels(resources.ModelConfigFile, provider)
		if err != nil {
			return resources, err
		}

		resources.ModelConfigHash = resourceHash(models)
	}

	return resources, nil
}

func resourcePath(configDir, reference string) string {
	if !filepath.IsAbs(reference) {
		reference = filepath.Join(configDir, reference)
	}

	return filepath.Clean(reference)
}

func readPrivateResource(path string) (data []byte, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, errors.New("selected private Pi resource is unavailable")
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > resourceLimit {
		return nil, errors.New("selected Pi resource must be a private regular file within 64 KiB")
	}

	data, err = io.ReadAll(io.LimitReader(file, resourceLimit+1))
	if err != nil || len(data) > resourceLimit || !json.Valid(data) {
		return nil, errors.New("selected Pi resource exceeds its JSON read boundary")
	}

	return data, nil
}

func selectedCredential(path, provider string) ([]byte, string, error) {
	data, err := readPrivateResource(path)
	if err != nil {
		return nil, "", err
	}

	var credentials map[string]json.RawMessage
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, "", errors.New("pi credentials require a provider mapping")
	}

	var credential struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	}

	decoder := json.NewDecoder(bytes.NewReader(credentials[provider]))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&credential); err != nil || credential.Type != "api_key" || credential.Key == "" {
		return nil, "", fmt.Errorf("selected Pi provider requires a static API-key credential; OAuth needs a refresh owner: %w", ErrCapability)
	}

	if strings.HasPrefix(credential.Key, "!") || strings.ContainsAny(credential.Key, "$\x00\r\n") {
		return nil, "", fmt.Errorf("dynamic Pi credential commands and environment expansion are unsupported: %w", ErrCapability)
	}

	selected, err := json.Marshal(map[string]any{provider: credential})

	return selected, credential.Key, err
}

func selectedModels(path, provider string) ([]byte, error) {
	data, err := readPrivateResource(path)
	if err != nil {
		return nil, err
	}

	var document struct {
		Providers map[string]json.RawMessage `json:"providers"`
	}
	if err := json.Unmarshal(data, &document); err != nil || len(document.Providers[provider]) == 0 {
		return nil, errors.New("pi model configuration requires the selected provider")
	}

	var definition any

	decoder := json.NewDecoder(bytes.NewReader(document.Providers[provider]))
	decoder.UseNumber()

	if err := decoder.Decode(&definition); err != nil {
		return nil, ErrCapability
	}

	fields, ok := definition.(map[string]any)
	if !ok || len(fields) == 0 {
		return nil, errors.New("pi model provider must be a nonempty definition")
	}

	for key := range fields {
		if !providerModelField(key) {
			return nil, fmt.Errorf("unsupported Pi model provider field; use a credential-free 1.3.0 definition: %w", ErrCapability)
		}
	}

	if err := validateModelDefinition(definition, 0); err != nil {
		return nil, err
	}

	return json.Marshal(map[string]any{"providers": map[string]any{provider: definition}})
}

func validateModelDefinition(value any, depth int) error {
	if depth > 32 {
		return fmt.Errorf("pi model configuration is too deeply nested: %w", ErrCapability)
	}

	switch value := value.(type) {
	case map[string]any:
		return validateModelFields(value, depth)
	case []any:
		for _, child := range value {
			if err := validateModelDefinition(child, depth+1); err != nil {
				return err
			}
		}
	case string:
		if strings.HasPrefix(value, "!") || strings.Contains(value, "$") {
			return fmt.Errorf("dynamic Pi model resources are unsupported: %w", ErrCapability)
		}
	}

	return nil
}

func validateModelFields(fields map[string]any, depth int) error {
	for key, child := range fields {
		if forbiddenModelField(key) {
			return fmt.Errorf("keep credentials in the selected auth file; model resources cannot supply authentication or commands: %w", ErrCapability)
		}

		if key == "baseUrl" {
			if err := validateModelEndpoint(child); err != nil {
				return err
			}
		}

		if err := validateModelDefinition(child, depth+1); err != nil {
			return err
		}
	}

	return nil
}

func forbiddenModelField(key string) bool {
	switch strings.ToLower(key) {
	case "apikey", "api_key", "headers", "authorization", "password", "token", "secret", "credentials", "oauth", "access", "refresh", "env", "command":
		return true
	default:
		return false
	}
}

func providerModelField(key string) bool {
	switch key {
	case "name", "api", "baseUrl", "compat", "authHeader", "models", "modelOverrides":
		return true
	default:
		return false
	}
}

func validateModelEndpoint(value any) error {
	text, ok := value.(string)

	parsed, err := url.Parse(text)
	if !ok || err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("pi model endpoints require an HTTP(S) URL without embedded credentials, queries or fragments: %w", ErrCapability)
	}

	return nil
}

func resourceHash(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

func loadCallResources(resources harness.GenerationResources, provider string) ([]byte, []byte, string, error) {
	auth, secret, err := selectedCredential(resources.CredentialFile, provider)
	if err != nil {
		return nil, nil, "", err
	}

	var models []byte
	if resources.ModelConfigFile != "" {
		models, err = selectedModels(resources.ModelConfigFile, provider)
		if err != nil || resourceHash(models) != resources.ModelConfigHash {
			return nil, nil, "", errors.New("selected Pi model configuration changed or is unavailable; restore its frozen definition")
		}
	}

	return auth, models, secret, nil
}
