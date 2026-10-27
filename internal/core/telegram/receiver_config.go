package telegram

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

// ReceiverConfig is private user configuration, separate from notifications.
type ReceiverConfig struct {
	Enabled            bool          `json:"enabled"`
	BotTokenEnv        string        `json:"bot_token_env"`
	PollTimeoutSeconds int           `json:"poll_timeout_seconds"`
	DefaultProjectID   string        `json:"default_project_id"`
	AllowedChatIDs     []int64       `json:"allowed_chat_ids"`
	AllowedTopicIDs    []int64       `json:"allowed_topic_ids,omitempty"`
	AllowedUsers       []AllowedUser `json:"allowed_users"`
}

// AllowedUser binds an exact sender to local authority.
type AllowedUser struct {
	TelegramUserID int64    `json:"telegram_user_id"`
	OperatorID     string   `json:"operator_id"`
	ProjectIDs     []string `json:"project_ids"`
}

// Validate rejects missing or ambiguous authority before any network access.
func (c ReceiverConfig) Validate() error {
	if !c.Enabled || !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,127}$`).MatchString(c.BotTokenEnv) {
		return errors.New("telegram must be enabled with a valid token environment reference")
	}

	if c.PollTimeoutSeconds < 1 || c.PollTimeoutSeconds > 50 || len(c.AllowedUsers) == 0 || len(c.AllowedUsers) > 100 {
		return errors.New("telegram polling or user configuration is outside its bounds")
	}

	if err := validateUsers(c); err != nil {
		return err
	}

	return validateNumericAllowlist(c.AllowedChatIDs, c.AllowedTopicIDs)
}

func validateUsers(c ReceiverConfig) error {
	seen := map[int64]bool{}
	for _, user := range c.AllowedUsers {
		if user.TelegramUserID <= 0 || seen[user.TelegramUserID] || !configID(user.OperatorID) || !projectIDs(user.ProjectIDs) {
			return errors.New("telegram users require unique numeric identities, an operator and explicit projects")
		}

		seen[user.TelegramUserID] = true
	}

	if c.DefaultProjectID != "" && !configID(c.DefaultProjectID) {
		return errors.New("invalid telegram default project")
	}

	return nil
}

func configID(value string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`).MatchString(value)
}

func projectIDs(ids []string) bool {
	if len(ids) == 0 || len(ids) > 50 {
		return false
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if !configID(id) || seen[id] {
			return false
		}

		seen[id] = true
	}

	return true
}

func validateNumericAllowlist(chats, topics []int64) error {
	if len(chats) > 100 || len(topics) > 100 {
		return errors.New("telegram numeric allowlist exceeds 100 entries")
	}

	for _, chat := range chats {
		if chat == 0 {
			return errors.New("telegram chat ID must be nonzero")
		}
	}

	for _, topic := range topics {
		if topic <= 0 {
			return errors.New("telegram topic ID must be positive")
		}
	}

	return nil
}

// AuthorizedUser checks the entire numeric peer before interpretation or lookup.
func (c ReceiverConfig) AuthorizedUser(chatID, topicID, userID int64, chatType string) (AllowedUser, bool) {
	if !c.authorizedPeer(chatID, topicID, userID, chatType) {
		return AllowedUser{}, false
	}

	for _, user := range c.AllowedUsers {
		if user.TelegramUserID == userID {
			return user, true
		}
	}

	return AllowedUser{}, false
}

func (c ReceiverConfig) authorizedPeer(chatID, topicID, userID int64, chatType string) bool {
	if topicID < 0 || (topicID != 0 && !slices.Contains(c.AllowedTopicIDs, topicID)) {
		return false
	}

	private := chatType == "private" && chatID == userID && topicID == 0

	group := (chatType == "group" || chatType == "supergroup") && slices.Contains(c.AllowedChatIDs, chatID)

	return private || group
}

// Digest identifies non-secret configuration for retained receiver diagnostics.
func (c ReceiverConfig) Digest() string {
	data := mustJSON(c)
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

func mustJSON(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err) // Callers use only statically JSON-encodable internal values.
	}

	return data
}

// LoadReceiverConfig reads bounded, private YAML without changing files.
func LoadReceiverConfig(env ConfigDirectory) (ReceiverConfig, error) {
	data, err := readPrivateFile(filepath.Join(env.ConfigDir(), "telegram.yml"), 64*1024)
	if err != nil {
		return ReceiverConfig{}, err
	}

	config, err := parseReceiverConfig(data)
	if err != nil {
		return ReceiverConfig{}, err
	}

	return config, config.Validate()
}

// SaveReceiverConfig preserves a differing configuration unless replacement is explicit.
func SaveReceiverConfig(env ConfigDirectory, config ReceiverConfig, replace bool) error {
	if err := config.Validate(); err != nil {
		return err
	}

	path := filepath.Join(env.ConfigDir(), "telegram.yml")

	prior, err := LoadReceiverConfig(env)
	if err == nil && prior.Digest() == config.Digest() {
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) && !replace {
		return errors.New("telegram configuration exists; use --replace to change it")
	}

	return writePrivateFile(path, renderReceiverConfig(config))
}

func readPrivateFile(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect private telegram file: %w", err)
	}

	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > limit {
		return nil, errors.New("telegram file must be private, regular and within its size limit")
	}

	return os.ReadFile(path)
}

func writePrivateFile(path string, data []byte) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".telegram-*")
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, removeTemp(file.Name())) }()

	if _, err := file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), path)
}

func removeTemp(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
