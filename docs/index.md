# Mycel

Mycel is a personal AI assistant that runs on your own machine. It talks to a local model through [Ollama](https://ollama.com), keeps your conversations stored, and is reachable from many places at  once (e.g., terminal, Telegram bot, etc.).

These pages cover everything you need to set it up.

## Where to start

- **[Automatic install](setup/automatic.md)** — one command that installs whatever your machine
  is missing, then builds the agent.
- **[Manual install](setup/manual.md)** — Go, Ollama and Redis: the pieces Mycel needs to run at
  all, installed one by one.
- **[Configuration](configuration.md)** — every environment variable, what it does and where the
  `.env` file lives.
- **[Commands](commands/index.md)** — the `/` instructions you can send the agent itself.
- **[Platforms](platforms/index.md)** — the places you can talk to Mycel from: the terminal, a
  Telegram chat.
- **[Tools](tools/index.md)** — optional integrations that extend what Mycel can do for you.

## How the pieces fit

```text
        you                       you
     (terminal)               (Telegram)      ← platforms
         │                         │
         └──────────┬──────────────┘
                    │
                 Mycel ──────────────► Resend  ← tools
                    │
        ┌───────────┴───────────┐
        │                       │
     Ollama                   Redis
  (local model)        (conversation history)
```

Platforms are where you reach Mycel from; tools are what Mycel can reach out to. Everything else
runs locally: the model, the history, the agent itself. No conversation leaves your machine unless
you enable a platform or a tool that sends it somewhere.
