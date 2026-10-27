package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// The receiver accepts this small configuration grammar, not executable YAML
// tags, aliases or implicit conversions. Unknown and repeated fields fail closed.
func parseTelegramReceiverConfig(data []byte) (TelegramReceiverConfig, error) {
	config := TelegramReceiverConfig{PollTimeoutSeconds: 30}
	parser := telegramConfigParser{config: &config, seen: map[string]bool{}}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if err := parser.line(line); err != nil {
			return TelegramReceiverConfig{}, errors.New("invalid telegram YAML configuration")
		}
	}

	if !parser.seen["telegram"] {
		return TelegramReceiverConfig{}, errors.New("telegram configuration section is required")
	}

	return config, nil
}

type telegramConfigParser struct {
	config     *TelegramReceiverConfig
	seen       map[string]bool
	userFields map[string]bool
	projects   bool
	userList   bool
}

func (p *telegramConfigParser) line(line string) error {
	if line == "telegram:" && !p.seen["telegram"] {
		p.seen["telegram"] = true

		return nil
	}

	if !p.seen["telegram"] || strings.ContainsAny(line, "\t\r") {
		return errors.New("invalid section")
	}

	indent := len(line) - len(strings.TrimLeft(line, " "))
	if indent == 2 {
		return p.field(strings.TrimSpace(line))
	}

	if !p.userList {
		return errors.New("unexpected list")
	}

	return p.userLine(indent, strings.TrimSpace(line))
}

func (p *telegramConfigParser) field(line string) error {
	key, value, err := p.fieldHeader(line)
	if err != nil {
		return err
	}

	switch key {
	case "enabled":
		p.config.Enabled, err = strconv.ParseBool(value)
	case "bot_token_env":
		p.config.BotTokenEnv, err = telegramYAMLScalar(value)
	case "poll_timeout_seconds":
		p.config.PollTimeoutSeconds, err = strconv.Atoi(value)
	case "default_project_id":
		p.config.DefaultProjectID, err = telegramYAMLScalar(value)
	case "allowed_chat_ids":
		err = json.Unmarshal([]byte(value), &p.config.AllowedChatIDs)
	case "allowed_topic_ids":
		err = json.Unmarshal([]byte(value), &p.config.AllowedTopicIDs)
	case "allowed_users":
		err = telegramUserListHeader(value)
	default:
		err = errors.New("unknown field")
	}

	return err
}

func telegramUserListHeader(value string) error {
	if value != "" {
		return errors.New("expected user list")
	}

	return nil
}

func (p *telegramConfigParser) fieldHeader(line string) (string, string, error) {
	key, value, ok := strings.Cut(line, ":")
	if !ok || p.seen[key] {
		return "", "", errors.New("duplicate or malformed field")
	}

	p.seen[key] = true
	p.projects = false
	p.userList = key == "allowed_users"

	return key, strings.TrimSpace(value), nil
}

func (p *telegramConfigParser) userLine(indent int, line string) error {
	if indent == 4 && strings.HasPrefix(line, "- telegram_user_id:") {
		id, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, "- telegram_user_id:")), 10, 64)
		if err != nil || len(p.config.AllowedUsers) >= 100 {
			return errors.New("invalid user ID or capacity")
		}

		p.config.AllowedUsers = append(p.config.AllowedUsers, TelegramAllowedUser{TelegramUserID: id})
		p.userFields = map[string]bool{}
		p.projects = false

		return nil
	}

	if len(p.config.AllowedUsers) == 0 {
		return errors.New("missing user")
	}

	user := &p.config.AllowedUsers[len(p.config.AllowedUsers)-1]
	if indent == 8 && p.projects && strings.HasPrefix(line, "- ") {
		id, err := telegramYAMLScalar(strings.TrimPrefix(line, "- "))
		user.ProjectIDs = append(user.ProjectIDs, id)

		return err
	}

	return p.userField(user, indent, line)
}

func (p *telegramConfigParser) userField(user *TelegramAllowedUser, indent int, line string) error {
	key, value, ok := strings.Cut(line, ":")
	if !ok || indent != 6 || p.userFields[key] {
		return errors.New("invalid user field")
	}

	p.userFields[key] = true
	p.projects = false
	value = strings.TrimSpace(value)

	switch key {
	case "operator_id":
		var err error

		user.OperatorID, err = telegramYAMLScalar(value)

		return err
	case "project_ids":
		p.projects = value == ""
		if p.projects {
			return nil
		}

		return json.Unmarshal([]byte(value), &user.ProjectIDs)
	default:
		return errors.New("unknown user field")
	}
}

func telegramYAMLScalar(value string) (string, error) {
	if strings.HasPrefix(value, `"`) {
		return strconv.Unquote(value)
	}

	if value == "" || !telegramConfigID(value) {
		return "", errors.New("invalid scalar")
	}

	return value, nil
}

func renderTelegramReceiverConfig(config TelegramReceiverConfig) []byte {
	chats := telegramMustJSON(config.AllowedChatIDs)
	topics := telegramMustJSON(config.AllowedTopicIDs)

	var body strings.Builder

	fmt.Fprintf(&body, "telegram:\n  enabled: %t\n  bot_token_env: %s\n  poll_timeout_seconds: %d\n  default_project_id: %q\n  allowed_chat_ids: %s\n  allowed_topic_ids: %s\n  allowed_users:\n",
		config.Enabled, config.BotTokenEnv, config.PollTimeoutSeconds, config.DefaultProjectID, chats, topics)

	for _, user := range config.AllowedUsers {
		fmt.Fprintf(&body, "    - telegram_user_id: %d\n      operator_id: %s\n      project_ids:\n", user.TelegramUserID, user.OperatorID)

		for _, id := range user.ProjectIDs {
			fmt.Fprintf(&body, "        - %s\n", id)
		}
	}

	return []byte(body.String())
}
