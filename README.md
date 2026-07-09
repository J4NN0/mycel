# Mycel

## Setup

### Telegram

1. Open Telegram and start a chat with [@BotFather](https://t.me/BotFather)
2. Send `/newbot` and follow the prompts to choose a name and username for your bot
3. Copy the token BotFather gives you and set it as an environment variable:
   ```sh
   export TELEGRAM_BOT_TOKEN=<your-token>
   ```

### Resend (email tool)

The agent can send emails through [Resend](https://resend.com).

1. Sign up at [resend.com](https://resend.com)
2. Create an API key (Dashboard → API Keys) and set it as `RESEND_API_KEY`
3. Set the sender address as `RESEND_FROM`:
   - For quick testing, use the shared sandbox sender `onboarding@resend.dev` — no setup required, but it can only send to the email address you signed up to Resend with.
   - To send to arbitrary recipients, verify a domain you own (Dashboard → Domains, add the SPF/DKIM records) and use an address on that domain, e.g. `agent@yourdomain.com`.

Leave `RESEND_API_KEY` empty to disable the email tool entirely.

## Run the Agent

Copy `.env.sample` to `.env` and fill in your values:

```sh
cp .env.sample .env
```

Then start all services:

```sh
docker compose up --build
```

Alternatively, for local development, run Redis in Docker and the agent directly on your machine:

```sh
make run
```

To stop:

```sh
docker compose down
```

To stop and wipe the Redis history:

```sh
docker compose down -v
```

## Resources
- [Bifrost](https://docs.getbifrost.ai/quickstart/go-sdk/setting-up)
