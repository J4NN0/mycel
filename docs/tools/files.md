# Files

Three tools let Mycel look at files on the machine it is running on: `list_files` to see what is in a folder, `grep_files` to find where something is written, and `read_file` to read one.

There is nothing to configure. They take no key and no address, so they are always available — tell Mycel where to look and it looks.

| Tool | What it does |
| --- | --- |
| `list_files` | Lists a folder: names, sizes, which entries are folders |
| `grep_files` | Searches the text of every file below a folder, returning matching lines with file and line number |
| `read_file` | Reads one text file |

## Using them

Point Mycel at a path, the way you would point a colleague at one:

> What's the deadline for project Alpha? My notes are in ~/notes

Mycel works down from there: it lists the folder, opens what looks relevant, and answers from what it actually read rather than from memory. A `~` is expanded, and relative paths resolve against the directory Mycel was started from — so a full path is always the clearer thing to give.

For "where did I write that?" questions, searching beats reading:

> Search ~/notes for everywhere the invoice is mentioned

## What they will and won't do

They only read. Nothing here writes, moves or deletes a file.

- **Binary files are refused.** A file with NUL bytes in it has no text to report, so Mycel says so rather than filling the conversation with noise.
- **Long files are cut** at 8,000 characters, because a whole file would crowd out the conversation itself. Mycel says when it has only seen the beginning.
- **A search skips machine-generated folders** — `.git`, `node_modules`, `vendor`, `.venv`, `__pycache__` — and stops after 50 matching lines or 2,000 files, so pointing it at something enormous degrades instead of hanging.
- **Long lines are shortened** in search results, to keep a match readable rather than dumping a minified file into the reply.

!!! warning "Any path your account can read"
    These tools are not confined to a folder: whatever your user account can read, Mycel can read
    if it is asked to, including `~/.ssh`, `~/.aws` and `~/.config/mycel/.env`. That matters because
    Mycel takes instructions from anywhere it is reachable — a [Telegram](../platforms/telegram.md)
    bot is public unless you keep its token private, and a web page Mycel
    [reads](web.md) can carry text that tries to talk to it. Keep the bot token to yourself, and
    think of the file tools as being as trusted as the people who can message the agent.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and the file tools are ignored along with the rest.
