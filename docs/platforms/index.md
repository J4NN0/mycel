# Platforms

A platform is a place you can talk to Mycel from. The agent itself is one process: whichever platform your message arrives on, the same model, history and [tools](../tools/index.md) answer it.

There is one shared set of conversations, not one per platform. The active conversation is the same
everywhere: ask something in the terminal, reply from Telegram, and the agent picks up the same
thread with the full history behind it. `/clear` and `/resume` act on that shared set too, so a
conversation started on your phone shows up in the terminal's `/resume` list and vice versa.

| Platform | How you reach it | Enabled by |
| --- | --- | --- |
| Terminal | The TUI that starts with `mycel` | always on |
| [Telegram](telegram.md) | A bot you chat with from your phone | `TELEGRAM_BOT_TOKEN` |

The terminal UI is always there. Telegram is optional: leave `TELEGRAM_BOT_TOKEN` unset and Mycel
logs that the platform is disabled and starts with the terminal alone.

!!! note "One conversation at a time, across all of them"
    Because the active conversation is shared, switching it anywhere switches it everywhere. Turns
    are handled one at a time, so two platforms writing at once cannot interleave — but a message
    sent from Telegram does not appear in the terminal transcript already on your screen. Run
    `/resume` there to bring the thread back into view.
