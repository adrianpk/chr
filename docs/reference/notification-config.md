# Notification Configuration

Notification configuration lives in the managed project's `.chr/default/config.yml`
file under the `notifications:` key.

Notifications are disabled by default. A project must explicitly enable a
channel before chr treats it as dispatchable.

## Telegram Fields

Supported Telegram fields:

- `telegram_enabled`: enables Telegram notifications when set to `true`.
- `telegram_bot_token_env`: environment variable name that contains the bot
  token.
- `telegram_chat_id`: Telegram chat ID that should receive project
  notifications.

Example:

```yml
notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHR_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
```

The bot token value is not stored in `.chr/default/config.yml`. The config stores the
environment variable name that later dispatch code can resolve.

## Validation

Disabled Telegram configuration is valid even when token and chat fields are
missing.

When Telegram notifications are enabled:

- `telegram_bot_token_env` is required;
- `telegram_chat_id` is required.

Validation does not contact Telegram or any external service. It only checks the
local configuration shape.
