# Mycel

<p align="center">
  <img src="assets/mycel.png" width="300" alt="mycel">
</p>

Mycel is a personal AI agent that runs on your own machine. It remembers your conversations and is reachable from wherever you are: a terminal, a Telegram chat, and more, with one shared conversation across all of them.

Hold a real conversation, think a problem through, look at a screenshot, search the web, read through your own files, send an email, or hand it a goal and let it work through the steps on its own.

See the [full documentation](https://j4nn0.github.io/mycel/) for how it works, what it can do, and how to configure it.

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
make install
```

It checks what your machine is missing, installs it, and builds the agent. See [setup](docs/setup/automatic.md) and [configuration](docs/configuration.md) in the docs for the details.

## Run

From any directory:

```sh
mycel
```

That opens the terminal UI. See [platforms](docs/platforms/index.md) for the other ways to reach Mycel.

## Documentation

The full docs are hosted at [j4nn0.github.io/mycel](https://j4nn0.github.io/mycel/), built from [`docs/`](docs/). To read them locally instead:

```sh
make docs
```

Then open [localhost:8000](http://localhost:8000). The first run installs the docs toolchain into its own virtualenv under `build/`.

## Resources
- [Bifrost](https://docs.getbifrost.ai/quickstart/go-sdk/setting-up)
- [Ollama](https://ollama.com)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [MkDocs](https://www.mkdocs.org/)
