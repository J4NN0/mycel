#!/usr/bin/env bash
# One ensure_* function per dependency. Each checks what is already there,
# installs only what is missing, and records a line in the summary.
#
# In --check mode nothing is installed: missing dependencies are reported instead.

GO_TOOLCHAIN_MIN="1.21" # from here on Go downloads the toolchain go.mod asks for
PYTHON_MIN="3.9"        # mkdocs-material 9.5 needs at least this

# dep_install_start <name> <message> — announce an install, or bail out in check
# mode after recording the gap.
dep_install_start() {
	if [ "$CHECK_ONLY" = "1" ]; then
		summary_add "$1" missing "$2"
		return 1
	fi
	log_step "$2"
	return 0
}

# --- required to build and run ----------------------------------------------

ensure_go() {
	local want="$GO_VERSION_WANTED" version=""
	have go && version="$(go version 2>/dev/null | awk '{print $3}' | sed 's/^go//')"

	if version_ge "$version" "$want"; then
		log_ok "Go $version"
		summary_add "go" ok "$version"
		return 0
	fi

	# Go 1.21+ reads the `go` directive in go.mod and fetches that exact toolchain
	# on demand, so an older-but-recent Go still builds Mycel.
	if version_ge "$version" "$GO_TOOLCHAIN_MIN"; then
		log_ok "Go $version (will fetch the go$want toolchain when building)"
		summary_add "go" ok "$version, fetches go$want on build"
		return 0
	fi

	local reason="Go $want or newer is required"
	[ -n "$version" ] && reason="Go $version is too old; $want or newer is required"
	dep_install_start "go" "$reason" || return 1

	if [ "$OS" = "macos" ]; then
		ensure_homebrew || return 1
		if [ -n "$version" ]; then
			run brew upgrade go || run brew install go || return 1
		else
			run brew install go || return 1
		fi
	else
		install_go_tarball "$want" || return 1
	fi

	have go || die "Go was installed but is not on PATH; open a new shell and re-run the installer"

	version="$(go version | awk '{print $3}' | sed 's/^go//')"
	if ! version_ge "$version" "$GO_TOOLCHAIN_MIN"; then
		log_err "Go $version is still too old; install go$want from https://go.dev/dl/"
		summary_add "go" failed "$version, need $GO_TOOLCHAIN_MIN+"
		return 1
	fi

	log_ok "Go $version installed"
	summary_add "go" installed "$version"
}

# install_go_tarball <version> — the official tarball into /usr/local/go. Linux
# distributions ship Go versions far behind what go.mod asks for, so we skip them.
install_go_tarball() {
	local version="$1" url tmp
	url="https://go.dev/dl/go${version}.linux-${ARCH}.tar.gz"
	tmp="$(mktemp -d)"

	log_info "downloading $url"
	curl -fsSL "$url" -o "$tmp/go.tar.gz" || {
		rm -rf "$tmp"
		log_err "could not download Go $version"
		return 1
	}

	run $SUDO rm -rf /usr/local/go
	run $SUDO tar -C /usr/local -xzf "$tmp/go.tar.gz" || {
		rm -rf "$tmp"
		return 1
	}
	rm -rf "$tmp"

	export PATH="$PATH:/usr/local/go/bin"
	profile_add_path "/usr/local/go/bin"
}

ensure_ollama() {
	if have ollama; then
		local version
		version="$(ollama --version 2>/dev/null | sed -n 's/.*version is //p')"
		log_ok "Ollama ${version:-installed}"
		summary_add "ollama" ok "${version:-present}"
		return 0
	fi

	dep_install_start "ollama" "Ollama not found (runs the local model)" || return 1

	if [ "$OS" = "macos" ]; then
		ensure_homebrew || return 1
		run brew install ollama || return 1
	else
		# The official script covers every distribution and sets up the systemd unit.
		log_info "running the official installer from ollama.com"
		curl -fsSL https://ollama.com/install.sh | sh || return 1
	fi

	have ollama || {
		log_err "Ollama install finished but 'ollama' is not on PATH"
		summary_add "ollama" failed "not on PATH"
		return 1
	}
	log_ok "Ollama installed"
	summary_add "ollama" installed "$(ollama --version 2>/dev/null | sed -n 's/.*version is //p')"
}

# ensure_redis — Mycel keeps history in Redis. Anything already listening on
# REDIS_ADDR is good enough; otherwise we need Docker to run the bundled one, so
# pointing REDIS_ADDR at a Redis you manage skips the Docker step entirely.
ensure_redis() {
	local host="${REDIS_HOST:-localhost}" port="${REDIS_PORT:-6379}"

	if port_open "$host" "$port"; then
		log_ok "Redis reachable at $host:$port"
		summary_add "redis" ok "$host:$port"
		return 0
	fi

	case "$host" in
	localhost | 127.0.0.1 | ::1 | "") ;;
	*)
		log_warn "nothing listening on $host:$port — start that Redis before running mycel"
		summary_add "redis" missing "$host:$port unreachable"
		return 1
		;;
	esac

	ensure_docker || {
		summary_add "redis" missing "needs Docker"
		return 1
	}
	docker_compose_up_redis
}

# ensure_searxng — web search runs through SearXNG. Anything already listening on
# SEARXNG_URL is good enough; otherwise Docker runs the bundled one, so pointing
# SEARXNG_URL at an instance you manage skips the Docker step entirely.
ensure_searxng() {
	local host="${SEARXNG_HOST:-localhost}" port="${SEARXNG_PORT:-8888}"

	if port_open "$host" "$port"; then
		log_ok "SearXNG reachable at $host:$port"
		summary_add "searxng" ok "$host:$port"
		return 0
	fi

	case "$host" in
	localhost | 127.0.0.1 | ::1 | "") ;;
	*)
		log_warn "nothing listening on $host:$port — web search stays off until it answers"
		summary_add "searxng" missing "$host:$port unreachable"
		return 1
		;;
	esac

	ensure_docker || {
		summary_add "searxng" missing "needs Docker"
		return 1
	}
	docker_compose_up_searxng
}

ensure_docker() {
	if have docker && docker info >/dev/null 2>&1; then
		log_ok "Docker running"
		summary_add "docker" ok "daemon running"
		ensure_docker_compose
		return 0
	fi

	if have docker; then
		dep_install_start "docker" "Docker is installed but the daemon is not running" || return 1
		start_docker_daemon || {
			summary_add "docker" failed "daemon not running"
			return 1
		}
		summary_add "docker" ok "daemon started"
		ensure_docker_compose
		return 0
	fi

	dep_install_start "docker" "Docker not found (runs Redis and SearXNG)" || return 1

	if [ "$OS" = "macos" ]; then
		ensure_homebrew || return 1
		run brew install --cask docker || return 1
	else
		log_info "running the official installer from get.docker.com"
		curl -fsSL https://get.docker.com | $SUDO sh || return 1
		run $SUDO systemctl enable --now docker || true
		if [ -n "$SUDO" ]; then
			run $SUDO usermod -aG docker "$(id -un)" || true
			log_warn "you were added to the 'docker' group: log out and back in for it to apply"
		fi
	fi

	start_docker_daemon || {
		summary_add "docker" failed "installed, daemon not running"
		return 1
	}
	log_ok "Docker installed and running"
	summary_add "docker" installed "daemon running"
	ensure_docker_compose
}

start_docker_daemon() {
	if docker info >/dev/null 2>&1; then
		return 0
	fi

	if [ "$OS" = "macos" ]; then
		[ -d /Applications/Docker.app ] || {
			log_err "Docker Desktop is not in /Applications; start it manually"
			return 1
		}
		log_step "Starting Docker Desktop"
		open -a Docker
	else
		run $SUDO systemctl start docker || true
	fi

	# The daemon takes a while to accept connections, especially on a first launch.
	local waited=0
	while [ "$waited" -lt 120 ]; do
		if docker info >/dev/null 2>&1; then
			log_ok "Docker daemon ready"
			return 0
		fi
		sleep 3
		waited=$((waited + 3))
		[ $((waited % 15)) -eq 0 ] && log_info "waiting for the Docker daemon (${waited}s)"
	done

	log_err "the Docker daemon did not come up within 120s; start Docker and re-run"
	return 1
}

ensure_docker_compose() {
	if docker compose version >/dev/null 2>&1; then
		return 0
	fi
	log_warn "'docker compose' (v2) is unavailable — the bundled Redis needs it"
	if [ "$OS" = "linux" ] && [ "$CHECK_ONLY" != "1" ]; then
		pkg_install docker-compose-plugin || true
	fi
}

ensure_make() {
	if have make; then
		log_ok "make $(make --version 2>/dev/null | head -1 | awk '{print $3}')"
		summary_add "make" ok "present"
		return 0
	fi

	dep_install_start "make" "make not found (drives the build)" || return 1

	if [ "$OS" = "macos" ]; then
		# make ships with the Command Line Tools rather than as a formula.
		run xcode-select --install || true
		log_warn "finish the Command Line Tools install in the dialog, then re-run this script"
		summary_add "make" failed "needs Xcode Command Line Tools"
		return 1
	fi

	pkg_install make apt:build-essential dnf:make || {
		summary_add "make" failed "install failed"
		return 1
	}
	summary_add "make" installed "present"
}

ensure_git() {
	if have git; then
		log_ok "git $(git --version | awk '{print $3}')"
		summary_add "git" ok "present"
		return 0
	fi

	dep_install_start "git" "git not found" || return 1
	pkg_install git || {
		summary_add "git" failed "install failed"
		return 1
	}
	summary_add "git" installed "present"
}

# --- optional: docs toolchain (make docs) ------------------------------------

# ensure_python — `make docs` builds a virtualenv and pip-installs MkDocs into it,
# so python3 needs both the venv and ensurepip modules. On Debian and Ubuntu those
# live in a separate package that is often absent.
ensure_python() {
	local version=""
	have python3 && version="$(python3 --version 2>/dev/null | awk '{print $2}')"

	if version_ge "$version" "$PYTHON_MIN" && python3 -c 'import venv, ensurepip' 2>/dev/null; then
		log_ok "Python $version with venv and pip"
		summary_add "python3" ok "$version"
		return 0
	fi

	local reason="Python 3 with venv support not found (needed for 'make docs')"
	if [ -n "$version" ]; then
		version_ge "$version" "$PYTHON_MIN" ||
			reason="Python $version is too old for the docs toolchain (need $PYTHON_MIN+)"
		python3 -c 'import venv, ensurepip' 2>/dev/null ||
			reason="Python $version is missing the venv/pip modules"
	fi
	dep_install_start "python3" "$reason" || return 1

	if [ "$OS" = "macos" ]; then
		ensure_homebrew || return 1
		run brew install python || return 1
	else
		pkg_install python3 || return 1
		# Debian/Ubuntu split venv and pip out of the base package.
		case "$PKG_MGR" in
		apt) pkg_install python3-venv && pkg_install python3-pip ;;
		esac
	fi

	if python3 -c 'import venv, ensurepip' 2>/dev/null; then
		log_ok "Python $(python3 --version | awk '{print $2}') ready"
		summary_add "python3" installed "$(python3 --version | awk '{print $2}')"
	else
		log_err "python3 still cannot create virtualenvs; 'make docs' will not work"
		summary_add "python3" failed "venv unavailable"
		return 1
	fi
}

# --- optional: developer tooling (make lint) ---------------------------------

ensure_golangci_lint() {
	if have golangci-lint; then
		log_ok "golangci-lint $(golangci-lint version 2>/dev/null | awk '{print $4}')"
		summary_add "golangci-lint" ok "present"
		return 0
	fi

	dep_install_start "golangci-lint" "golangci-lint not found (needed for 'make lint')" || return 1

	if [ "$OS" = "macos" ] && have brew; then
		run brew install golangci-lint || return 1
	elif ! have go; then
		log_warn "no Go toolchain to install golangci-lint with"
		summary_add "golangci-lint" missing "needs Go"
		return 1
	else
		local bindir
		bindir="$(go env GOPATH)/bin"
		run go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest || {
			summary_add "golangci-lint" failed "install failed"
			return 1
		}
		export PATH="$PATH:$bindir"
	fi

	if ! have golangci-lint; then
		log_err "golangci-lint is still not on PATH; 'make lint' will not work"
		summary_add "golangci-lint" failed "not on PATH"
		return 1
	fi

	summary_add "golangci-lint" installed "present"
}

ensure_curl() {
	have curl || die "curl is required to bootstrap the other dependencies; install it first"
}
