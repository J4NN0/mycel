# Commands

Most of what you send Mycel is conversation. Exceptions to that rule are commands, which refer to a few instructions targeted towards the agent instead of the model.

A command is any message beginning with `/`:

```text
/help
```

This is handled by Mycel directly and responds instantly – without even reaching the model, incurs no tokens, and does not form part of your conversation history.

## Arguments

Some commands take an argument, written after a space. All that follows the command name is considered the argument, including spaces:

```text
/goal book a table for four on Friday
```

Commands that take an optional argument do something sensible without one — `/resume` on its own  offers a choice, `/resume 3` goes straight there.

## Getting them wrong

Nothing breaks. A `/` message that Mycel doesn’t understand will be regarded as normal chat and sent through to the model. This means you get a response even if there’s a typo. Use `/help` if you want to know what exists.

No matter where you’re chatting with Mycel, the commands will be the same, and the client will usually help you to write them: just start typing `/`.

**[See every command →](commands.md)**
