# List files

`list_files` shows what is in a folder: names, sizes, and which entries are folders.

## Usage

Point Mycel at a path, the way you would point a colleague at one:

> What's in my notes folder? It's ~/notes

Mycel lists the folder and works from what is actually there. A `~` is expanded, and relative paths resolve against the directory Mycel was started from — so a full path is always the clearer thing to give.

## Scope and limits

It only reads. Nothing here writes, moves or deletes a file.

!!! warning "Any path your account can read"
    This tool is not confined to a folder: whatever your user account can read, Mycel can list
    if it is asked to, including `~/.ssh`, `~/.aws` and `~/.config/mycel/.env`. That matters because
    Mycel takes instructions from anywhere it is reachable — a [Telegram](../platforms/telegram.md)
    bot is public unless you keep its token private, and a web page Mycel
    [reads](fetch_url.md) can carry text that tries to talk to it. Keep the bot token to yourself, and
    think of the file tools as being as trusted as the people who can message the agent.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and `list_files` is ignored along with the rest.
