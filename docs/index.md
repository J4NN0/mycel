# Mycel

Mycel is a personal AI assistant that runs on your own machine. It talks to a local model through [Ollama](https://ollama.com), keeps your conversations stored, and is reachable from many places at  once (e.g., terminal, Telegram bot, etc.).

These pages cover everything you need to set it up.

## Where to start

- **[Prerequisites](setup/prerequisites.md)** — Go, Ollama and Redis: the pieces Mycel needs to run
  at all.
- **[Configuration](configuration.md)** — every environment variable, what it does and where the
  `.env` file lives.
- **[Commands](commands/index.md)** — the `/` instructions you can send the agent itself.
- **[Tools](tools/index.md)** — optional integrations that extend what Mycel can reach.

## How the pieces fit

```text
        you                       you
     (terminal)               (Telegram)
         │                         │
         └──────────┬──────────────┘
                    │
                 Mycel
                    │
        ┌───────────┴───────────┐
        │                       │
     Ollama                   Redis
  (local model)        (conversation history)
```

Everything except Telegram and Resend runs locally: the model, the history, the agent itself. No
conversation leaves your machine unless you enable a tool that sends it somewhere.
