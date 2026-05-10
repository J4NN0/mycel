# Mycel

## Setup

### Telegram

1. Open Telegram and start a chat with [@BotFather](https://t.me/BotFather)
2. Send `/newbot` and follow the prompts to choose a name and username for your bot
3. Copy the token BotFather gives you and set it as an environment variable:
   ```sh
   export TELEGRAM_BOT_TOKEN=<your-token>
   ```

## Run the Agent

Copy `.env.sample` to `.env` and fill in your values:

```sh
cp .env.sample .env
```

Then start all services:

```sh
docker compose up --build
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
