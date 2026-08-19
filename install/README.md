# Installer

Everything needed to get Mycel running on a machine that has nothing installed yet.

```sh
make install
```

The script figures out what is already there, installs only what is missing, builds the agent and pulls the model. Re-running it is harmless: anything in place is left alone, and an existing `~/.config/mycel/.env` is never overwritten.

Because installing Go, Docker or Ollama means running package managers and upstream install scripts, the run starts with a dry pass and shows what it is about to touch:

```text
---> Looking at what this machine already has
  ·    to install or start: ollama docker
  ?    Go ahead? [Y/n]
```

That is the only question, and there is nothing to confirm when nothing is missing. A missing terminal is not consent: without one the script stops rather than assuming yes.

## What it takes care of

| Dependency      | Why Mycel needs it                      | Where it comes from                                  |
|-----------------|-----------------------------------------|------------------------------------------------------|
| `git`, `make`   | fetching and building                   | package manager (macOS: Xcode Command Line Tools)    |
| Go              | building the agent                      | Homebrew on macOS, official tarball on Linux         |
| Ollama          | running the local model                 | Homebrew on macOS, `ollama.com/install.sh` on Linux  |
| Docker          | running Redis (history) and SearXNG (search) | Docker Desktop on macOS, `get.docker.com` on Linux   |
| Python 3 + venv | `make docs` (MkDocs in its own venv)    | package manager (plus `python3-venv` on Debian)      |
| golangci-lint   | `make lint`                             | Homebrew on macOS, `go install` elsewhere            |

It then writes `~/.config/mycel/.env` (from your `.env`, or from `.env.sample`), puts `mycel` in `$GOPATH/bin`, adds that directory to your `PATH` if it is missing, starts Redis and SearXNG, and pulls the model named by `LLM_MODEL`.

The required Go version is read from `go.mod`, so the installer and the build can never disagree. A Go older than that is still accepted from 1.21 onwards, because Go downloads the toolchain `go.mod` asks for on its own.

Neither container is started when it does not have to be: if something already answers on `REDIS_ADDR` and `SEARXNG_URL`, Docker is never touched. Point either variable at an instance you manage and that step disappears.

## Options

| Flag           | Effect                                                        |
|----------------|---------------------------------------------------------------|
| `--core-only`  | only what is needed to run: skip the docs and lint toolchains  |
| `--skip-model` | don't pull the model; Mycel pulls it on first run instead      |

## Checking without installing

`doctor.sh` runs the installer's checks and nothing else, which is the quickest way to
see why a machine cannot run the agent:

```sh
make doctor              # or: ./install/doctor.sh
```

```
Summary
  ok        go               1.26.5
  missing   ollama           Ollama not found (runs the local model)
  ok        redis            localhost:6379
```

## Supported platforms

macOS and Linux. On macOS everything goes through Homebrew, which the script offers to
install if it is absent. On Linux it uses whichever of `apt`, `dnf`, `yum`, `pacman` or
`zypper` is present, and falls back to the upstream installers for Go, Ollama and Docker
because distribution packages tend to lag behind. Windows works through WSL2.

## Layout

```text
install/
├── install.sh        entry point: installs what is missing
├── doctor.sh         entry point: reports what is missing, changes nothing
└── lib/
    ├── bootstrap.sh  shared wiring, defaults, and the order the steps run in
    ├── log.sh        output, prompts, the summary table
    ├── platform.sh   OS/arch/package-manager detection, version compare, PATH edits
    ├── deps.sh       one ensure_* function per dependency
    └── setup.sh      config file, Redis, SearXNG, binary, model
```

Both entry points share `run_steps` in `bootstrap.sh`; the only difference is that
`doctor.sh` leaves `CHECK_ONLY` on for the whole run.

## Adding a dependency

Write an `ensure_<name>` function in `lib/deps.sh` following the shape of the others:

1. Detect what is present and `log_ok` + `summary_add "<name>" ok` when it is good enough.
2. Call `dep_install_start "<name>" "<why it is needed>"`, which records the gap and
   returns non-zero while checking so nothing gets installed.
3. Install it — `pkg_install` handles the package-manager differences, and takes
   `manager:name` overrides for packages that are named differently (for example
   `pkg_install make apt:build-essential`).
4. Verify, then `summary_add "<name>" installed`.

Then call it from `run_steps` in `lib/bootstrap.sh` with `|| true`, so one missing piece
still lets the rest of the run finish and report.

The scripts target bash 3.2, the version macOS ships, so no associative arrays or `mapfile`.
