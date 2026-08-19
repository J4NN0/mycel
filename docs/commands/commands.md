# Commands

Every command Mycel understands, and what it does.

| Command | What it does |
| --- | --- |
| `/start` | Say hello and start interacting with Mycel |
| `/help` | List the available commands |
| `/clear` | Start a new conversation |
| `/resume` | Resume a past conversation |
| `/goal <goal>` | Give Mycel a goal to work toward autonomously |
| `/model` | Show which model is currently in use |

## `/start`

Just a simple greeting and nothing else: Mycel greets you and allows you to greet it back. It does not change any state, your dialogue with its entire history remain exactly the way you have left it. It exists mostly as a convention: chat clients tend to open a new conversation with a `/start` button, and this is what answers it.

## `/help`

Prints the list above, each command with its one-line description, straight from the agent. Useful
when you can't remember whether it's `/clear` or `/new`.

## `/clear`

Starts a new conversation. The model loses the thread — the next message begins from a clean system prompt with no memory of what came before, which is what you want when you switch to an unrelated topic and don't want the old context steering the answers.

Nothing is deleted. The previous conversation is kept, and you can come back to it with `/resume`.

Conversations are shared across [platforms](../platforms/index.md), so clearing from Telegram also moves the terminal on to the new conversation.

## `/resume`

Brings back a past conversation. With no argument it lists your recent ones, each with a short preview of how it started, excluding the one you're already in:

```text
/resume
```

The choices are presented however the client you're on presents them. Whichever you choose becomes the active conversation, and your next message continues it  with the full history restored.

The list is not per platform: conversations you started on Telegram appear in the terminal's picker and vice versa, and resuming one makes it active everywhere.

Conversation IDs are plain incrementing numbers, so you can also skip the picker when you know the one you want:

```text
/resume 3
```

## `/goal`

Hands Mycel an objective and lets it work unattended. It plans its own next step, executes it — using any [tools](../tools/index.md) it has — reviews the result, and repeats, pausing between steps. The loop ends when Mycel decides the goal is met, or after a bounded number of steps if it never gets there.

```text
/goal draft and send me a summary of what we discussed today
```

It runs in the background: you get an immediate "Goal accepted" and stay free to keep chatting. Progress is reported through the agent's log output rather than the chat, so run with `LOG_LEVEL=debug` if you want to watch it think step by step.

!!! warning "Goals need a capable model"
    A goal loop leans hard on tool calling and multi-step reasoning. On a small model it tends to
    repeat itself or declare victory early.

## `/model`

Prints the model currently backing the agent — a quick way to confirm which `LLM_MODEL` took effect after editing your [configuration](../configuration.md).
