# Read file

`read_file` reads one text file.

## Usage

Point Mycel at a path, the way you would point a colleague at one:

> What's the deadline for project Alpha? It's in ~/notes/alpha.md

Mycel opens the file and answers from what it actually read rather than from memory. A `~` is expanded, and relative paths resolve against the directory Mycel was started from — so a full path is always the clearer thing to give. Use [`list_files`](list_files.md) or [`grep_files`](grep_files.md) first if you don't know the exact path.

## Scope and limits

It only reads. Nothing here writes, moves or deletes a file.

- **Binary files are refused.** A file with NUL bytes in it has no text to report, so Mycel says so rather than filling the conversation with noise.
- **Long files are cut** at 8,000 characters, because a whole file would crowd out the conversation itself. Mycel says when it has only seen the beginning.

!!! warning "Any path your account can read"
    This tool is not confined to a folder: whatever your user account can read, Mycel can read
    if it is asked to, including `~/.ssh`, `~/.aws` and `~/.config/mycel/.env`. That matters because
    Mycel takes instructions from anywhere it is reachable — a [Telegram](../platforms/telegram.md)
    bot is public unless you keep its token private, and a web page Mycel
    [reads](fetch_url.md) can carry text that tries to talk to it. Keep the bot token to yourself, and
    think of the file tools as being as trusted as the people who can message the agent.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and `read_file` is ignored along with the rest.
