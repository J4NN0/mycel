# Configuration

Mycel is configured entirely through environment variables, usually kept in a `.env` file. Start
from the sample:

```sh
cp .env.sample .env
```

## Where the `.env` file lives

| How you run it | File used |
| --- | --- |
| `make run` (from the repo) | `.env` in the repo root, exported into the process |
| `mycel` (installed binary) | `~/.config/mycel/.env` |

`make install` creates `~/.config/mycel/` and seeds it with your repo `.env` — or with `.env.sample`
if you don't have one yet — and never overwrites a config that is already there.

Variables already set in your shell always win: the config file is loaded without overriding the
existing environment, so `LLM_MODEL=llama3.2 mycel` works as a one-off override.

## Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `PROVIDER` | yes | — | LLM provider. Only `ollama` is supported today. |
| `LLM_MODEL` | yes | — | Model to run, e.g. `qwen3.5:9b`. Pulled automatically at startup. |
| `PERSONA` | no | `neutral` | Voice the agent speaks in: `neutral`, `oracle` or `influencer`. |
| `MAX_HISTORY_MESSAGES` | no | `20` | Message-count cap before history is compacted. |
| `MAX_HISTORY_TOKENS` | no | `6000` | Prompt-token cap before history is compacted. |
| `TELEGRAM_BOT_TOKEN` | no | — | Enables the [Telegram](platforms/telegram.md) platform. |
| `REDIS_ADDR` | no | `localhost:6379` | Address of the Redis instance holding conversations. |
| `LOG_LEVEL` | no | `info` | `panic`, `fatal`, `error`, `warn`, `info`, `debug` or `trace`. |
| `RESEND_API_KEY` | no | — | Enables the [email tool](tools/resend.md), together with `RESEND_FROM`. |
| `RESEND_FROM` | no | — | Sender address used by the email tool. |

## Personas

`PERSONA` swaps the voice, not the behaviour — the agent's purpose and tools stay the same.

- `neutral` — plain, warm, unhurried. The default.
- `oracle` — calm and contemplative, speaks to the question beneath the question.
- `influencer` — sharp, witty, tech-lifestyle energy.

## History and compaction

Conversations are stored in Redis and shared by every [platform](platforms/index.md): one active
conversation at a time, with as many past ones as you have started alongside it.

When a conversation outgrows its limits, Mycel summarizes the older messages instead of dropping
them: the system prompt and the last few exchanges are kept verbatim, everything before them is
folded into a running summary. `MAX_HISTORY_TOKENS` is the limit that normally applies, measured
against the prompt tokens the provider reports; `MAX_HISTORY_MESSAGES` is the fallback when no token
count is available.
