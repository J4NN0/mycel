# Mycel

<p align="center">
  <img src="assets/mycel.png" width="300" alt="mycel">
</p>

Mycel is a personal AI assistant that runs on your own machine. It talks to a local model through [Ollama](https://ollama.com), remembers your conversations, and answers from wherever you are: a terminal UI, a Telegram chat, etc.

Hold a real conversation, think a problem through, look at a screenshot, or hand it a goal and let it work through the steps on its own. Nothing leaves your machine unless you explicitly wire up a tool that sends something out.

## Table of Contents

- [Installation](#installation)
- [Run](#run)
- [Documentation](#documentation)
- [Resources](#resources)

## Installation

Clone the repo and run the installer:

```sh
git clone https://github.com/J4NN0/mycel.git
cd mycel
make setup
```

It checks what your machine is missing and installs it. See [its README](install/README.md) or the
[automatic install guide](docs/setup/automatic.md). Prefer to install the pieces yourself? [Manual install](docs/setup/manual.md) walks through them one by one.

Not sure what is missing? `make doctor` reports it without installing anything.

Once installed, fill in the variable that are relevant for you (e.g., Telegram, Resend, etc.) in `~/.config/mycel/.env`. Every variable is documented in the [configuration reference](docs/configuration.md).

## Run

Once installed, from any directory:

```sh
mycel
```

Or straight from the repo, which also brings up everything the agent depends on:

```sh
make run
```

The first reads `~/.config/mycel/.env`, the second the repo's own `.env`. Either way, Ollama is started and the model pulled automatically on first run.

## Documentation

The full docs live in [`docs/`](docs/). To read them locally:

```sh
make docs
```

Then open [localhost:8000](http://localhost:8000). The first run installs the docs toolchain into its own virtualenv under `build/`.

## Resources
- [Bifrost](https://docs.getbifrost.ai/quickstart/go-sdk/setting-up)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [MkDocs](https://www.mkdocs.org/)
