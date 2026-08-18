#!/usr/bin/env bash
# The shared spine of install.sh and doctor.sh: loads the libraries, sets the
# defaults both need, and defines the order the steps run in.

LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$LIB_DIR/../.." && pwd)"

# shellcheck source=log.sh
. "$LIB_DIR/log.sh"
# shellcheck source=platform.sh
. "$LIB_DIR/platform.sh"
# shellcheck source=deps.sh
. "$LIB_DIR/deps.sh"
# shellcheck source=setup.sh
. "$LIB_DIR/setup.sh"

# CHECK_ONLY makes every step report what it would do instead of doing it. The
# installer flips it on for a dry run before asking to go ahead; doctor.sh leaves
# it on for the whole run.
CHECK_ONLY=0
CORE_ONLY=0
SKIP_MODEL=0
PROFILE_CHANGED=0

# The Go version go.mod asks for, so the installer and the build never disagree.
GO_VERSION_WANTED="$(awk '/^go /{print $2; exit}' "$REPO_ROOT/go.mod")"

# run_steps — every step, in dependency order. Each one is allowed to fail so that
# a single missing piece still lets the rest of the run report back.
run_steps() {
	log_title "Build toolchain"
	ensure_git || true
	ensure_make || true
	ensure_go || true

	# The config file comes before the services: it says which Redis to reach
	# and which model to pull.
	log_title "Configuration"
	setup_config_env

	log_title "Runtime dependencies"
	ensure_ollama || true
	ensure_redis || true

	if [ "$CORE_ONLY" = "1" ]; then
		log_info "skipping the docs and lint toolchains (--core-only)"
	else
		log_title "Optional toolchains"
		ensure_python || true
		ensure_golangci_lint || true
	fi

	log_title "Mycel"
	setup_binary || true
	setup_model || true
}
