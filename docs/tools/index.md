# Tools

Mycel runs perfectly well with nothing plugged into it. Tools are what you add on top to give it more reach — another place to talk to it, or another action it can take on your behalf.

Every tool is optional and switched on the same way: set its environment variables, restart the agent. Leave them unset and the tool is simply never registered — Mycel logs that it is disabled and starts without it.

| Tool | What it adds | Enabled by |
| --- | --- | --- |
| [Telegram](telegram.md) | Chat with Mycel from your phone, alongside the terminal | `TELEGRAM_BOT_TOKEN` |
| [Resend](resend.md) | Let Mycel send emails on your behalf | `RESEND_API_KEY` + `RESEND_FROM` |
