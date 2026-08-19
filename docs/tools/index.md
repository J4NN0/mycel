# Tools

Tools are the actions Mycel can take on your behalf, beyond writing an answer back to you. They are what the model reaches for when a reply is not enough — sending an email, for instance.

Mycel runs perfectly well with nothing plugged into it. The tools that need an account or an address elsewhere are switched on the same way: set the environment variables, restart the agent. Leave them unset and the tool is simply never registered — Mycel logs that it is disabled and starts without it.

The rest have nothing to configure, so they are always there (e.g., reading a web page, file tools).

| Tool | What it lets Mycel do | Enabled by |
| --- | --- | --- |
| [Resend](resend.md) | Send emails on your behalf | `RESEND_API_KEY` + `RESEND_FROM` |
| [Web search](web.md) | Search the web for what is true today | on by default (bundled SearXNG) |
| [Web read](web.md) | Read a page and use what it says | always on |
| [Files](files.md) | List, search and read files on your machine | always on |

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and every tool is ignored no matter how it is configured.
