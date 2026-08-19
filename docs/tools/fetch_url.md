# Fetch URL

`fetch_url` reads one page and returns its text. Without it the agent can only answer from what the model already knows, which was fixed when the model was trained. With it, it can check what a specific page says today.

## Usage

> Have a look at https://go.dev/doc/go1.26 and tell me what changed in the garbage collector.

It reads and nothing more. It cannot log in, fill in a form or run a page's JavaScript, so a site that renders entirely in the browser comes back empty. Pages are cut to roughly 6,000 characters, because a whole page would crowd out the conversation itself.

It also refuses to connect to your own machine or your local network — `localhost`, `127.0.0.1`, `10.x`, `192.168.x` and the rest, including a public hostname that resolves to one of them. Ollama, Redis and anything else you run stay out of reach, whoever is doing the asking.

[`web_search`](web_search.md) is the tool to reach for first when you don't already have a link — the agent calls `fetch_url` itself on whichever result looks worth opening.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and `fetch_url` is ignored along with the rest.
