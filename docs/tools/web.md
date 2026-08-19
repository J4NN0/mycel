# Web

Two tools cover the web, and they are meant to be used together: `web_search` finds pages, `fetch_url` reads one.

Without them the agent can only answer from what the model already knows, which was fixed when the model was trained. With them it can check what is true today.

| Tool | What it does | Enabled by |
| --- | --- | --- |
| `fetch_url` | Reads one page and returns its text | always on |
| `web_search` | Searches the web and returns five results — title, link, snippet | on by default, through the bundled SearXNG |

## Reading a page

`fetch_url` needs no configuration, so it is always available: ask about a link and the agent can go and read it.

> Have a look at https://go.dev/doc/go1.26 and tell me what changed in the garbage collector.

It reads and nothing more. It cannot log in, fill in a form or run a page's JavaScript, so a site that renders entirely in the browser comes back empty. Pages are cut to roughly 6,000 characters, because a whole page would crowd out the conversation itself.

It also refuses to connect to your own machine or your local network — `localhost`, `127.0.0.1`, `10.x`, `192.168.x` and the rest, including a public hostname that resolves to one of them. Ollama, Redis and anything else you run stay out of reach, whoever is doing the asking.

## Searching

There is nothing to set up: `make install` starts a [SearXNG](https://docs.searxng.org) container, and `SEARXNG_URL` already points at it.

SearXNG queries other search engines on your behalf and holds no account of its own, which is why this needs no API key and no signup — there is no company on the other end building a profile from what you look up.

=== "The bundled instance"

    `make install` and `make doctor` treat it exactly like Redis: started if it is not already
    running, left alone if it is. To manage it by hand:

    ```sh
    docker compose up -d searxng     # start
    docker compose stop searxng      # stop
    ```

    Mycel checks the instance at startup and only offers the tool if it answers, so a stopped
    container means an agent that knows it cannot search, rather than one that tries and fails.

=== "Use your own"

    Any SearXNG instance works, as long as its settings allow JSON output:

    ```yaml
    search:
      formats:
        - html
        - json
    ```

    JSON is off by default, and without it every search comes back `403`. The bot limiter answers
    programmatic requests the same way, so leave `server.limiter: false` on an instance only you can
    reach.

    ```dotenv
    SEARXNG_URL=https://searxng.example.com
    ```

Then just ask:

> What did Go 1.26 change about the garbage collector?

The agent searches, reads the snippets, and calls `fetch_url` on whichever result looks worth opening.

!!! note "Inside Docker"
    Running Mycel itself with `docker compose` puts it on the same network as SearXNG, where
    `localhost` means the container. Use `SEARXNG_URL=http://searxng:8080` there.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and both web tools are ignored no matter how they are configured.
