# Prerequisites

Mycel needs three things before it can start: a Go toolchain to build it, Ollama to run the model,  and Redis to store conversations.

## Go

Mycel is built with Go — see [`go.mod`](https://github.com/J4NN0/mycel/blob/main/go.mod) for the
version the module targets.

```sh
brew install go   # macOS
go version
```

If you install the binary with `make install`, it lands in `$GOPATH/bin` (usually `~/go/bin`), so
make sure that directory is on your `PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Ollama

The model runs locally through [Ollama](https://ollama.com).

```sh
brew install ollama   # macOS, or download from ollama.com/download
```

You do not need to start it or pull the model yourself: on startup Mycel runs `ollama serve` if nothing is listening, then `ollama pull <LLM_MODEL>` and streams the download progress to the  terminal. All it needs is the `ollama` binary on your `PATH`.

Pick the model with `LLM_MODEL` (see [Configuration](../configuration.md)). Mycel inspects the model's capabilities at startup and warns you about what it cannot do — for example a model without `vision` support will reject images, and one without `tools` support will ignore the email tool.

!!! tip "Choosing a model"
    A model with `tools`, `vision` and `thinking` support gets you every feature. Smaller models
    work fine for plain conversation but often drop tool calling first.

## Redis

Conversation history lives in Redis. The repo ships a `docker-compose.yml` with everything you need:

```sh
docker compose up -d redis
```

`make run` does this for you, and also starts [Redis Commander](http://localhost:8081), a web UI for
inspecting what Mycel has stored.

Already have a Redis instance? Point `REDIS_ADDR` at it instead and skip Docker entirely.
