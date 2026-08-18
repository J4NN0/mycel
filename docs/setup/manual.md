# Manual installation

Mycel needs three things before it can start: a Go toolchain to build the project, Ollama to run the model(s), and Docker (i.e., Redis to store conversations, etc.). The installer puts all of them there for you.

## 1. Go

Mycel is built with Go — see [`go.mod`](https://github.com/J4NN0/mycel/blob/main/go.mod) for the
version the module targets.

```sh
brew install go   # macOS
go version
```

Go 1.21 and newer downloads the exact toolchain `go.mod` asks for on its own, so a slightly older Go still builds Mycel.

The binary lands in `$GOPATH/bin` (usually `~/go/bin`), so make sure that directory is on your `PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

## 2. Ollama

The model runs locally through [Ollama](https://ollama.com).

```sh
brew install ollama   # macOS, or download from ollama.com/download
```

You do not need to start it or pull the model yourself: on startup Mycel runs `ollama serve` if nothing is listening, then `ollama pull <LLM_MODEL>` and streams the download progress to the terminal. All it needs is the `ollama` binary on your `PATH`.

Pick the model with `LLM_MODEL` (see [Configuration](../configuration.md)). Mycel inspects the model's capabilities at startup and warns you about what it cannot do — for example a model without `vision` support will reject images, and one without `tools` support will ignore the email tool.

!!! tip "Choosing a model"
    A model with `tools`, `vision` and `thinking` support gets you every feature. Smaller models
    work fine for plain conversation but often drop tool calling first.

## 3. Redis

Conversation history lives in Redis. The repo ships a `docker-compose.yml` with everything you need:

```sh
docker compose up -d redis
```

`make run` does this for you, and also starts [Redis Commander](http://localhost:8081), a web UI for inspecting what Mycel has stored.

Already have a Redis instance? Point `REDIS_ADDR` at it instead and skip Docker entirely.

## 4. Configure and build

Copy the sample config, then fill in your values — every variable is documented in the [configuration reference](../configuration.md):

```sh
cp .env.sample .env
```

Then build and install the binary:

```sh
make install
```

This installs `mycel` into `$GOPATH/bin` and copies your `.env` to `~/.config/mycel/.env`, so the agent can be started from any directory. An existing config there is left untouched.

## Optional toolchains

Only needed if you plan to work on Mycel itself:

- **Docs** — `make docs` builds [MkDocs](https://www.mkdocs.org/) into its own virtualenv, so it
  needs Python 3.9+ with the `venv` and `pip` modules (on Debian and Ubuntu those live in the
  separate `python3-venv` and `python3-pip` packages).
- **Lint** — `make lint` needs [golangci-lint](https://golangci-lint.run):
  `brew install golangci-lint`.
