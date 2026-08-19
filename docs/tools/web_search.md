# Web search

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `SEARXNG_URL` | no | `http://localhost:8888` | Where the SearXNG instance behind `web_search` listens. |

`web_search` searches the web and returns five results — title, link, snippet — through a bundled [SearXNG](https://docs.searxng.org) instance. Without it the agent can only answer from what the model already knows, which was fixed when the model was trained. With it, it can check what is true today.

There is nothing to set up: `make install` starts the SearXNG container, and `SEARXNG_URL` already points at it.

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

## Usage

> What did Go 1.26 change about the garbage collector?

The agent searches, reads the snippets, and calls [`fetch_url`](fetch_url.md) on whichever result looks worth opening.

!!! note "Inside Docker"
    Running Mycel itself with `docker compose` puts it on the same network as SearXNG, where
    `localhost` means the container. Use `SEARXNG_URL=http://searxng:8080` there.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and `web_search` is ignored along with the rest.
