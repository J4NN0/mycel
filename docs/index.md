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
                 Mycel ──────────────► the web, Resend  ← tools
                    │                └────► your files    ← local, read only
        ┌───────────┼───────────────┐
        │           │               │
     Ollama       Redis          SearXNG
  (local model)  (history)   (search, also local)
```

Platforms are where you reach Mycel from; tools are what Mycel can reach out to. Everything else
runs locally: the model, the history, the agent itself — your conversation is never sent to anyone
to be answered.

What does leave your machine is what a platform or a tool carries out: a Telegram chat goes through
Telegram, a search goes to your own SearXNG instance, an email goes to Resend, and asking Mycel to
read a link sends a request to that site.

Some tools point the other way. The [file tools](tools/files.md) read what is already on your
machine and send nothing anywhere — but they can read any file your account can, so what Mycel is
able to see is worth knowing about if the agent is reachable from more than your own terminal.
