# Telegram

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `TELEGRAM_BOT_TOKEN` | no | — | Enables the Telegram platform. |
| `TELEGRAM_ALLOWED_USER_IDS` | yes, if `TELEGRAM_BOT_TOKEN` is set | — | Comma-separated Telegram user IDs allowed to talk to the bot. Required once the token is set; the bot refuses to start without at least one. |

Mycel always runs in your terminal. Give it a Telegram bot token and the same agent is reachable from your phone as well, both platforms served by the one process you started.

History is shared, not per channel: the chat continues whatever conversation is active, so you can start something in the terminal and pick it up on your phone. `/resume` on Telegram lists the same past conversations the terminal offers.

This step is optional: without `TELEGRAM_BOT_TOKEN` the agent simply logs that the Telegram platform is disabled and carries on with the terminal UI.

## Create the bot

1. Open Telegram and start a chat with [@BotFather](https://t.me/BotFather).
2. Send `/newbot` and follow the prompts to choose a name and a username for your bot.
3. Copy the token BotFather gives you and set it as `TELEGRAM_BOT_TOKEN` in your `.env`.
4. Get your numeric Telegram user ID — for example by messaging [@userinfobot](https://t.me/userinfobot) — and set it as `TELEGRAM_ALLOWED_USER_IDS`.

    ```dotenv
    TELEGRAM_BOT_TOKEN=123456789:AA...
    TELEGRAM_ALLOWED_USER_IDS=987654321
    ```

Restart Mycel, open the chat with your bot and send `/start`.

## What you get

Mycel registers its [commands](../commands/commands.md) with Telegram on startup, so typing `/` in the chat brings up the same list as the terminal, with `/resume` rendering past conversations as inline buttons you can tap.

You can also send photos: Telegram downloads them, and Mycel passes them to the model as long as the model has vision support.

!!! warning "A Telegram bot is public by default — only the token is secret"
    Anyone who finds your bot's username can open a chat with it, so Mycel only responds to the user
    IDs listed in `TELEGRAM_ALLOWED_USER_IDS`; the bot refuses to start if that list is empty. It's a
    comma-separated list of numeric Telegram user IDs, e.g. `TELEGRAM_ALLOWED_USER_IDS=111,222` if
    more than one person you trust should be able to reach the agent. Keep the bot token out of
    version control too (`.env` is already gitignored), and revoke it through BotFather if it leaks.
