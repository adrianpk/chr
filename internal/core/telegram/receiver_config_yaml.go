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
func parseReceiverConfig(data []byte) (ReceiverConfig, error) {
	config := ReceiverConfig{PollTimeoutSeconds: 30}
	parser := configParser{config: &config, seen: map[string]bool{}}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if err := parser.line(line); err != nil {
			return ReceiverConfig{}, errors.New("invalid telegram YAML configuration")
		}
	}

	if !parser.seen["telegram"] {
		return ReceiverConfig{}, errors.New("telegram configuration section is required")
	}

	return config, nil
}

type configParser struct {
	config     *ReceiverConfig
	seen       map[string]bool
	userFields map[string]bool
	projects   bool
	userList   bool
}

func (p *configParser) line(line string) error {
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

func (p *configParser) field(line string) error {
	key, value, err := p.fieldHeader(line)
	if err != nil {
		return err
	}

	switch key {
	case "enabled":
		p.config.Enabled, err = strconv.ParseBool(value)
	case "bot_token_env":
		p.config.BotTokenEnv, err = yAMLScalar(value)
	case "poll_timeout_seconds":
		p.config.PollTimeoutSeconds, err = strconv.Atoi(value)
	case "default_project_id":
		p.config.DefaultProjectID, err = yAMLScalar(value)
	case "allowed_chat_ids":
		err = json.Unmarshal([]byte(value), &p.config.AllowedChatIDs)
	case "allowed_topic_ids":
		err = json.Unmarshal([]byte(value), &p.config.AllowedTopicIDs)
	case "allowed_users":
		err = userListHeader(value)
	default:
		err = errors.New("unknown field")
	}

	return err
}

func userListHeader(value string) error {
	if value != "" {
		return errors.New("expected user list")
	}

	return nil
}

func (p *configParser) fieldHeader(line string) (string, string, error) {
	key, value, ok := strings.Cut(line, ":")
	if !ok || p.seen[key] {
		return "", "", errors.New("duplicate or malformed field")
	}

	p.seen[key] = true
	p.projects = false
	p.userList = key == "allowed_users"

	return key, strings.TrimSpace(value), nil
}

func (p *configParser) userLine(indent int, line string) error {
	if indent == 4 && strings.HasPrefix(line, "- telegram_user_id:") {
		id, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, "- telegram_user_id:")), 10, 64)
		if err != nil || len(p.config.AllowedUsers) >= 100 {
			return errors.New("invalid user ID or capacity")
		}

		p.config.AllowedUsers = append(p.config.AllowedUsers, AllowedUser{TelegramUserID: id})
		p.userFields = map[string]bool{}
		p.projects = false

		return nil
	}

	if len(p.config.AllowedUsers) == 0 {
		return errors.New("missing user")
	}

	user := &p.config.AllowedUsers[len(p.config.AllowedUsers)-1]
	if indent == 8 && p.projects && strings.HasPrefix(line, "- ") {
		id, err := yAMLScalar(strings.TrimPrefix(line, "- "))
		user.ProjectIDs = append(user.ProjectIDs, id)

		return err
	}

	return p.userField(user, indent, line)
}

func (p *configParser) userField(user *AllowedUser, indent int, line string) error {
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

		user.OperatorID, err = yAMLScalar(value)

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

func yAMLScalar(value string) (string, error) {
	if strings.HasPrefix(value, `"`) {
		return strconv.Unquote(value)
	}

	if value == "" || !configID(value) {
		return "", errors.New("invalid scalar")
	}

	return value, nil
}

func renderReceiverConfig(config ReceiverConfig) []byte {
	chats := mustJSON(config.AllowedChatIDs)
	topics := mustJSON(config.AllowedTopicIDs)

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
