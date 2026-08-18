# Automatic installation

Mycel needs three things before it can start: a Go toolchain to build the project, Ollama to run the model(s), and Docker (i.e., Redis to store conversations, etc.). The installer puts all of them there for you.

## Checking what is missing

To see the gaps without installing anything:

```sh
make doctor
```

```text
Summary
  ok        go               1.26.5
  missing   ollama           Ollama not found (runs the local model)
  ok        redis            localhost:6379
```

## Installation

From a fresh clone, the following one command will install anything that's missing:

```sh
make setup
```

The [installer](https://github.com/J4NN0/mycel/tree/main/install) inspects your machine, installs only the pieces that are absent, builds `mycel`, writes `~/.config/mycel/.env` and pulls the model. It is safe to re-run, and it never overwrites an existing config.

Since installing Go, Docker or Ollama means running package managers and upstream install scripts, the run starts with a dry pass and shows what it is about to touch:

```text
---> Looking at what this machine already has
  ·    to install or start: ollama docker
  ?    Go ahead? [Y/n]
```

There is nothing to confirm when nothing is missing.

### Options

Pass these to `./install/install.sh` directly:

| Flag           | Effect                                                        |
|----------------|---------------------------------------------------------------|
| `--core-only`  | only what is needed to run: skip the docs and lint toolchains  |
| `--skip-model` | don't pull the model; Mycel pulls it on first run instead       |

Already have a Redis you manage? Point `REDIS_ADDR` at it and the Docker step disappears on
its own — there is nothing to pass.

!!! note "Supported platforms"
    macOS and Linux. On macOS the installer goes through Homebrew, offering to install it
    if needed; on Linux it uses `apt`, `dnf`, `yum`, `pacman` or `zypper`, falling back to
    the upstream installers for Go, Ollama and Docker. Windows works through WSL2.

## Ready to run

After the installation is done, `mycel` is sitting in `$GOPATH/bin` (which is only added to your shell profile if it wasn't there previously), so you can launch the agent from any place on your filesystem.

The final step is up to you: provide values for your environment variables in `~/.config/mycel/.env` (all variables are described in [Configuration](../configuration.md)), then execute `mycel`.
