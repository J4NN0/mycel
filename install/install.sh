#!/usr/bin/env bash
#
# Mycel installer: works out what is missing, installs it, and leaves you with a
# working agent. Safe to re-run — anything already in place is left alone.
#
#   ./install/install.sh               # install everything
#   ./install/install.sh --core-only   # skip the docs and lint toolchains
#
# To see what is missing without installing anything, run ./install/doctor.sh
# (or `make doctor`). See install/README.md for the details.

set -euo pipefail

# shellcheck source=lib/bootstrap.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/bootstrap.sh"

usage() {
	cat <<'USAGE'
Usage: install/install.sh [options]

Installs everything Mycel needs — Go, Ollama, Docker for Redis, the docs and lint
toolchains — then builds the agent, writes ~/.config/mycel/.env and pulls the model.
Already-installed pieces are detected and left alone, so re-running is harmless.

Options:
      --core-only  just what is needed to run: skip the docs and lint toolchains
      --skip-model don't pull the model (Mycel pulls it on first run instead)
  -h, --help       this message
USAGE
}

parse_args() {
	while [ $# -gt 0 ]; do
		case "$1" in
		--core-only) CORE_ONLY=1 ;;
		--skip-model) SKIP_MODEL=1 ;;
		-h | --help)
			usage
			exit 0
			;;
		*)
			usage >&2
			die "unknown option '$1'"
			;;
		esac
		shift
	done
}

# confirm_plan — installing Go, Docker or Ollama means running package managers and
# upstream install scripts, so show what will be touched and ask once. A dry run of
# run_steps produces the list; it is side-effect free by design.
confirm_plan() {
	local missing
	log_step "Looking at what this machine already has"

	CHECK_ONLY=1
	run_steps >/dev/null 2>&1 || true
	CHECK_ONLY=0

	# The config file and the binary are rebuilt on every run; only report the
	# things that actually change the machine.
	missing="$(summary_missing mycel config)"
	summary_reset

	if [ -z "$missing" ]; then
		log_info "nothing missing; rebuilding the agent and refreshing its setup"
		return 0
	fi

	log_info "to install or start:$(printf ' %s' "$missing" | tr -s ' ')"
	confirm "Go ahead?" || die "nothing was changed"
}

main() {
	parse_args "$@"

	[ -t 0 ] || die "this installs system packages, so it needs a terminal to confirm"

	log_title "Mycel — installing"
	detect_platform
	ensure_curl

	confirm_plan
	run_steps
	summary_print

	if summary_has_failures; then
		printf '\n  %sSome pieces need your attention — see the summary above.%s\n' \
			"$C_YELLOW" "$C_RESET"
		print_next_steps
		exit 1
	fi

	print_next_steps
}

main "$@"
