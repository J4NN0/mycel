# Platforms

A platform is a place you can talk to Mycel from. The agent itself is one process: whichever platform your message arrives on, the same model, history and [tools](../tools/index.md) answer it.

Each platform keeps its own history. The terminal is one session, and every Telegram chat is
another, keyed by chat ID — so a conversation you start on your phone stays separate from the one
in your terminal.

| Platform | How you reach it | Enabled by |
| --- | --- | --- |
| Terminal | The TUI that starts with `mycel` | always on |
| [Telegram](telegram.md) | A bot you chat with from your phone | `TELEGRAM_BOT_TOKEN` |

The terminal UI is always there. Telegram is optional: leave `TELEGRAM_BOT_TOKEN` unset and Mycel
logs that the platform is disabled and starts with the terminal alone.
