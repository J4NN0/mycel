# Grep files

`grep_files` searches the text of every file below a folder, returning matching lines with file and line number.

## Usage

For "where did I write that?" questions, searching beats reading one file at a time:

> Search ~/notes for everywhere the invoice is mentioned

A `~` is expanded, and relative paths resolve against the directory Mycel was started from — so a full path is always the clearer thing to give. Follow up with [`read_file`](read_file.md) once a match points you at the right one.

## Scope and limits

It only reads. Nothing here writes, moves or deletes a file.

- **A search skips machine-generated folders** — `.git`, `node_modules`, `vendor`, `.venv`, `__pycache__` — and stops after 50 matching lines or 2,000 files, so pointing it at something enormous degrades instead of hanging.
- **Long lines are shortened** in search results, to keep a match readable rather than dumping a minified file into the reply.

!!! warning "Any path your account can read"
    This tool is not confined to a folder: whatever your user account can read, Mycel can search
    if it is asked to, including `~/.ssh`, `~/.aws` and `~/.config/mycel/.env`. That matters because
    Mycel takes instructions from anywhere it is reachable — a [Telegram](../platforms/telegram.md)
    bot is public unless you keep its token private, and a web page Mycel
    [reads](fetch_url.md) can carry text that tries to talk to it. Keep the bot token to yourself, and
    think of the file tools as being as trusted as the people who can message the agent.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and `grep_files` is ignored along with the rest.
