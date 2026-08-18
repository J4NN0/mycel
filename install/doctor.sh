#!/usr/bin/env bash
#
# Reports what Mycel is missing, changing nothing: the same checks the installer
# runs, without the installing. Exits non-zero when something is missing.
# `make doctor` runs this.

set -euo pipefail

# shellcheck source=lib/bootstrap.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/bootstrap.sh"

CHECK_ONLY=1

log_title "Mycel — checking dependencies"
detect_platform
ensure_curl

run_steps
summary_print

if summary_has_failures; then
	printf '\n  %sRun make install to fix the gaps above.%s\n' "$C_YELLOW" "$C_RESET"
	exit 1
fi

printf '\n  %sEverything is in place — run '\''mycel'\'' to start.%s\n' "$C_GREEN" "$C_RESET"
