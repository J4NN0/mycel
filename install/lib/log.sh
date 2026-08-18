#!/usr/bin/env bash
# Output helpers: colours, log levels, prompts and a summary table.

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
	C_RESET="$(printf '\033[0m')"
	C_BOLD="$(printf '\033[1m')"
	C_DIM="$(printf '\033[2m')"
	C_RED="$(printf '\033[31m')"
	C_GREEN="$(printf '\033[32m')"
	C_YELLOW="$(printf '\033[33m')"
	C_BLUE="$(printf '\033[34m')"
else
	C_RESET="" C_BOLD="" C_DIM="" C_RED="" C_GREEN="" C_YELLOW="" C_BLUE=""
fi

log_title() { printf '\n%s%s%s\n' "$C_BOLD" "$*" "$C_RESET"; }
log_step() { printf '%s--->%s %s\n' "$C_BLUE" "$C_RESET" "$*"; }
log_ok() { printf '  %sok%s   %s\n' "$C_GREEN" "$C_RESET" "$*"; }
log_info() { printf '  %s·%s    %s\n' "$C_DIM" "$C_RESET" "$*"; }
log_warn() { printf '  %swarn%s %s\n' "$C_YELLOW" "$C_RESET" "$*" >&2; }
log_err() { printf '  %sfail%s %s\n' "$C_RED" "$C_RESET" "$*" >&2; }

die() {
	log_err "$*"
	exit 1
}

# run <command...> — echo the command, then run it.
run() {
	printf '  %s$ %s%s\n' "$C_DIM" "$*" "$C_RESET"
	"$@"
}

# confirm <question> — yes/no prompt, defaulting to yes. A missing terminal is not
# consent: the installer refuses to run without one, so nothing answers on your behalf.
confirm() {
	if [ ! -t 0 ]; then
		log_err "cannot ask '$1' without a terminal"
		return 1
	fi

	local reply=""
	printf '  %s?%s    %s [Y/n] ' "$C_YELLOW" "$C_RESET" "$1"
	read -r reply || return 1
	case "$reply" in
	"" | y | Y | yes | YES | Yes) return 0 ;;
	*) return 1 ;;
	esac
}

# --- summary -----------------------------------------------------------------
# Each dependency records one line so the run ends with a single readable table.

SUMMARY_NAMES=()
SUMMARY_STATES=()
SUMMARY_NOTES=()

# summary_add <name> <ok|installed|skipped|missing|failed> <note>
summary_add() {
	SUMMARY_NAMES+=("$1")
	SUMMARY_STATES+=("$2")
	SUMMARY_NOTES+=("$3")
}

summary_print() {
	log_title "Summary"

	local i=0
	while [ "$i" -lt "${#SUMMARY_NAMES[@]}" ]; do
		local state="${SUMMARY_STATES[$i]}" colour="$C_DIM"
		case "$state" in
		ok | installed) colour="$C_GREEN" ;;
		skipped) colour="$C_DIM" ;;
		missing | failed) colour="$C_RED" ;;
		esac
		printf '  %s%-9s%s %-16s %s%s%s\n' \
			"$colour" "$state" "$C_RESET" "${SUMMARY_NAMES[$i]}" \
			"$C_DIM" "${SUMMARY_NOTES[$i]}" "$C_RESET"
		i=$((i + 1))
	done
}

summary_reset() {
	SUMMARY_NAMES=()
	SUMMARY_STATES=()
	SUMMARY_NOTES=()
}

# summary_missing [name-to-skip...] — the names of everything reported as missing,
# space separated.
summary_missing() {
	local i=0 out="" name skip
	while [ "$i" -lt "${#SUMMARY_NAMES[@]}" ]; do
		name="${SUMMARY_NAMES[$i]}"
		for skip in "$@"; do
			[ "$name" = "$skip" ] && name=""
		done
		if [ -n "$name" ] && [ "${SUMMARY_STATES[$i]}" = "missing" ]; then
			out="$out $name"
		fi
		i=$((i + 1))
	done
	printf '%s' "${out# }"
}

# summary_has_failures — true when something is missing or failed to install.
summary_has_failures() {
	local i=0
	while [ "$i" -lt "${#SUMMARY_STATES[@]}" ]; do
		case "${SUMMARY_STATES[$i]}" in
		missing | failed) return 0 ;;
		esac
		i=$((i + 1))
	done
	return 1
}
