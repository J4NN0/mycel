#!/usr/bin/env bash
# Everything that happens once the dependencies are in place: the config file,
# Redis, the binary itself and the model.

CONFIG_DIR="$HOME/.config/mycel"
CONFIG_ENV="$CONFIG_DIR/.env"

REDIS_HOST=""
REDIS_PORT=""
LLM_MODEL_NAME=""

# env_file_value <key> <file> — read KEY=value from an env file, last one wins.
env_file_value() {
	[ -f "$2" ] || return 0
	sed -n "s/^[[:space:]]*$1[[:space:]]*=[[:space:]]*//p" "$2" |
		tail -1 |
		sed -e 's/[[:space:]]*$//' -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'$/\1/"
}

# setup_config_env — put a config file at ~/.config/mycel/.env so `mycel` works
# from any directory, and read back the values the rest of the install needs.
# An existing config is never overwritten.
setup_config_env() {
	if [ "$CHECK_ONLY" = "1" ]; then
		if [ -f "$CONFIG_ENV" ]; then
			log_ok "config at $CONFIG_ENV"
			summary_add "config" ok "$CONFIG_ENV"
		else
			log_warn "no config at $CONFIG_ENV"
			summary_add "config" missing "$CONFIG_ENV"
		fi
	elif [ -f "$CONFIG_ENV" ]; then
		log_ok "config already at $CONFIG_ENV, leaving it untouched"
		summary_add "config" ok "$CONFIG_ENV"
	else
		mkdir -p "$CONFIG_DIR"
		if [ -f "$REPO_ROOT/.env" ]; then
			cp "$REPO_ROOT/.env" "$CONFIG_ENV"
			log_ok "copied the repo's .env to $CONFIG_ENV"
		else
			cp "$REPO_ROOT/.env.sample" "$CONFIG_ENV"
			log_ok "created $CONFIG_ENV from .env.sample"
		fi
		summary_add "config" installed "$CONFIG_ENV"
	fi

	# Variables already exported win, exactly like config.ReadConfig does.
	local addr="${REDIS_ADDR:-}"
	[ -n "$addr" ] || addr="$(env_file_value REDIS_ADDR "$CONFIG_ENV")"
	[ -n "$addr" ] || addr="localhost:6379"
	REDIS_HOST="${addr%:*}"
	REDIS_PORT="${addr##*:}"

	LLM_MODEL_NAME="${LLM_MODEL:-}"
	[ -n "$LLM_MODEL_NAME" ] || LLM_MODEL_NAME="$(env_file_value LLM_MODEL "$CONFIG_ENV")"
	[ -n "$LLM_MODEL_NAME" ] || LLM_MODEL_NAME="$(env_file_value LLM_MODEL "$REPO_ROOT/.env.sample")"

	log_info "Redis at $REDIS_HOST:$REDIS_PORT, model $LLM_MODEL_NAME"
}

# docker_compose_up_redis — start the Redis the repo ships with.
docker_compose_up_redis() {
	if [ "$CHECK_ONLY" = "1" ]; then
		log_warn "nothing listening on $REDIS_HOST:$REDIS_PORT"
		summary_add "redis" missing "not running"
		return 1
	fi

	log_step "Starting Redis"
	if ! (cd "$REPO_ROOT" && run docker compose up -d --wait redis); then
		log_err "could not start Redis with docker compose"
		summary_add "redis" failed "docker compose up failed"
		return 1
	fi

	log_ok "Redis running at $REDIS_HOST:$REDIS_PORT"
	summary_add "redis" installed "$REDIS_HOST:$REDIS_PORT"
}

# setup_binary — build and install `mycel`, and make sure its directory is on PATH.
setup_binary() {
	log_step "Installing the mycel binary"

	if ! have go; then
		log_warn "no Go toolchain, so the binary cannot be built"
		summary_add "mycel" missing "needs Go"
		return 1
	fi

	local bindir
	bindir="$(go env GOBIN)"
	[ -n "$bindir" ] || bindir="$(go env GOPATH)/bin"

	if [ "$CHECK_ONLY" = "1" ]; then
		if have mycel; then
			log_ok "mycel installed at $(command -v mycel)"
			summary_add "mycel" ok "$(command -v mycel)"
		else
			log_warn "the mycel binary is not on PATH"
			summary_add "mycel" missing "not built"
		fi
		return 0
	fi

	if ! (cd "$REPO_ROOT" && run go install ./cmd/mycel); then
		log_err "build failed"
		summary_add "mycel" failed "go install failed"
		return 1
	fi

	case ":$PATH:" in
	*":$bindir:"*) ;;
	*)
		log_warn "$bindir is not on your PATH"
		if confirm "Add it to $(shell_profile)?"; then
			profile_add_path "$bindir"
		fi
		export PATH="$PATH:$bindir"
		;;
	esac

	log_ok "installed $bindir/mycel"
	summary_add "mycel" installed "$bindir/mycel"
}

# ensure_ollama_serving — `ollama pull` needs a server to talk to. Mycel does this
# itself at startup; we do it here so the model is ready before the first run.
ensure_ollama_serving() {
	if ollama list >/dev/null 2>&1; then
		return 0
	fi

	log_info "starting 'ollama serve' in the background"
	nohup ollama serve >"${TMPDIR:-/tmp}/mycel-ollama-serve.log" 2>&1 &

	local waited=0
	while [ "$waited" -lt 30 ]; do
		if ollama list >/dev/null 2>&1; then
			return 0
		fi
		sleep 2
		waited=$((waited + 2))
	done

	log_warn "Ollama is not answering yet; Mycel will start it again on first run"
	return 1
}

# setup_model — pull the model up front so the first conversation is instant
# instead of waiting on a multi-gigabyte download.
setup_model() {
	log_step "Model $LLM_MODEL_NAME"

	if [ "$SKIP_MODEL" = "1" ]; then
		log_info "skipping the model pull (--skip-model); Mycel pulls it on first run"
		summary_add "model" skipped "--skip-model"
		return 0
	fi

	if ! have ollama; then
		log_warn "no Ollama, so the model cannot be pulled"
		summary_add "model" skipped "Ollama unavailable"
		return 1
	fi

	if [ "$CHECK_ONLY" = "1" ]; then
		# Checking must not start anything, so we can only look at a server
		# that is already up.
		if ! ollama list >/dev/null 2>&1; then
			log_info "Ollama is not serving; cannot tell whether $LLM_MODEL_NAME is pulled"
			summary_add "model" skipped "Ollama not serving"
			return 0
		fi
	else
		ensure_ollama_serving || {
			summary_add "model" skipped "Ollama not serving"
			return 1
		}
	fi

	if ollama list 2>/dev/null | awk 'NR > 1 {print $1}' | grep -qx "$LLM_MODEL_NAME"; then
		log_ok "$LLM_MODEL_NAME already pulled"
		summary_add "model" ok "$LLM_MODEL_NAME"
		return 0
	fi

	if [ "$CHECK_ONLY" = "1" ]; then
		log_warn "$LLM_MODEL_NAME is not pulled yet"
		summary_add "model" missing "$LLM_MODEL_NAME"
		return 1
	fi

	log_info "pulling $LLM_MODEL_NAME (this can take a while)"
	if ! ollama pull "$LLM_MODEL_NAME"; then
		log_warn "could not pull $LLM_MODEL_NAME; Mycel will retry on first run"
		summary_add "model" failed "$LLM_MODEL_NAME"
		return 1
	fi

	log_ok "$LLM_MODEL_NAME pulled"
	summary_add "model" installed "$LLM_MODEL_NAME"
}

print_next_steps() {
	log_title "Next steps"

	if [ "${PROFILE_CHANGED:-0}" = "1" ]; then
		printf '  Reload your shell so the new PATH applies:\n    %ssource %s%s\n\n' \
			"$C_BOLD" "$(shell_profile)" "$C_RESET"
	fi

	printf '  Review your configuration (tokens for Telegram, Resend, ...):\n    %s%s%s\n\n' \
		"$C_BOLD" "$CONFIG_ENV" "$C_RESET"
	printf '  Then start the agent from anywhere:\n    %smycel%s\n\n' "$C_BOLD" "$C_RESET"
	printf '  %sChanged the code? Re-run make install to rebuild the binary.%s\n' \
		"$C_DIM" "$C_RESET"
}
