# Telegram

Mycel always runs in your terminal. Give it a Telegram bot token and the same agent is reachable from your phone as well, both platforms served by the one process you started.

Each channel keeps its own history: the terminal is one session, and every Telegram chat is another, keyed by chat ID.

This step is optional: without `TELEGRAM_BOT_TOKEN` the agent simply logs that the Telegram platform is disabled and carries on with the terminal UI.

## Create the bot

1. Open Telegram and start a chat with [@BotFather](https://t.me/BotFather).
2. Send `/newbot` and follow the prompts to choose a name and a username for your bot.
3. Copy the token BotFather gives you and set it as `TELEGRAM_BOT_TOKEN` in your `.env`.

```dotenv
TELEGRAM_BOT_TOKEN=123456789:AA...
```

Restart Mycel, open the chat with your bot and send `/start`.

## What you get

Mycel registers its [commands](../commands/commands.md) with Telegram on startup, so typing `/` in the chat brings up the same list as the terminal, with `/resume` rendering past conversations as inline buttons you can tap.

You can also send photos: Telegram downloads them, and Mycel passes them to the model as long as the model has vision support.

!!! warning "Anyone who finds your bot can talk to it"
    A Telegram bot is public by default — the token is the only secret. Keep it out of version
    control (`.env` is already gitignored), and revoke it through BotFather if it leaks.
