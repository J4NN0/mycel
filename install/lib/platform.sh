#!/usr/bin/env bash
# Platform detection and the few primitives every dependency needs:
# a package manager, sudo, version comparison.

OS=""      # macos | linux
ARCH=""    # amd64 | arm64
PKG_MGR="" # brew | apt | dnf | yum | pacman | zypper | none
SUDO=""    # "sudo" when privilege escalation is needed and available

detect_platform() {
	case "$(uname -s)" in
	Darwin) OS="macos" ;;
	Linux) OS="linux" ;;
	*) die "unsupported OS '$(uname -s)'. Mycel installs on macOS and Linux (Windows via WSL2)." ;;
	esac

	case "$(uname -m)" in
	x86_64 | amd64) ARCH="amd64" ;;
	arm64 | aarch64) ARCH="arm64" ;;
	*) die "unsupported architecture '$(uname -m)'" ;;
	esac

	if [ "$(id -u)" != "0" ] && have sudo; then
		SUDO="sudo"
	fi

	detect_pkg_mgr
	log_info "$OS/$ARCH, package manager: ${PKG_MGR}"
}

detect_pkg_mgr() {
	if have brew; then
		PKG_MGR="brew"
	elif have apt-get; then
		PKG_MGR="apt"
	elif have dnf; then
		PKG_MGR="dnf"
	elif have yum; then
		PKG_MGR="yum"
	elif have pacman; then
		PKG_MGR="pacman"
	elif have zypper; then
		PKG_MGR="zypper"
	else
		PKG_MGR="none"
	fi
}

have() { command -v "$1" >/dev/null 2>&1; }

# version_ge <have> <want> — true when <have> is at least <want>.
# Compares dot-separated numbers, ignoring any suffix (1.26.2-rc1 == 1.26.2).
version_ge() {
	[ -n "$1" ] || return 1
	awk -v have="$1" -v want="$2" '
		BEGIN {
			gsub(/[^0-9.].*$/, "", have)
			gsub(/[^0-9.].*$/, "", want)
			n = split(have, h, "."); m = split(want, w, ".")
			for (i = 1; i <= (n > m ? n : m); i++) {
				hi = (i <= n ? h[i] + 0 : 0); wi = (i <= m ? w[i] + 0 : 0)
				if (hi > wi) exit 0
				if (hi < wi) exit 1
			}
			exit 0
		}'
}

# ensure_homebrew — Homebrew is how we install nearly everything on macOS.
ensure_homebrew() {
	if have brew; then
		return 0
	fi

	log_step "Homebrew not found (needed to install packages on macOS)"
	confirm "Install Homebrew from brew.sh?" || return 1

	NONINTERACTIVE=1 /bin/bash -c \
		"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" ||
		return 1

	# The installer does not touch the current shell, so put brew on PATH by hand.
	local prefix
	for prefix in /opt/homebrew /usr/local /home/linuxbrew/.linuxbrew; do
		if [ -x "$prefix/bin/brew" ]; then
			eval "$("$prefix/bin/brew" shellenv)"
			break
		fi
	done

	have brew || return 1
	detect_pkg_mgr
	log_ok "Homebrew installed"
}

# pkg_install <package> [package-for-this-manager...] — install through the system
# package manager. Extra arguments override the name per manager, in the form
# "manager:name" (e.g. pkg_install make apt:build-essential).
pkg_install() {
	local pkg="$1"
	shift

	local override
	for override in "$@"; do
		case "$override" in
		"$PKG_MGR":*) pkg="${override#*:}" ;;
		esac
	done

	case "$PKG_MGR" in
	brew) run brew install "$pkg" ;;
	apt)
		run $SUDO apt-get update -qq
		run $SUDO env DEBIAN_FRONTEND=noninteractive apt-get install -y "$pkg"
		;;
	dnf) run $SUDO dnf install -y "$pkg" ;;
	yum) run $SUDO yum install -y "$pkg" ;;
	pacman) run $SUDO pacman -S --needed --noconfirm "$pkg" ;;
	zypper) run $SUDO zypper --non-interactive install "$pkg" ;;
	none)
		log_err "no supported package manager found; install '$pkg' manually"
		return 1
		;;
	esac
}

# port_open <host> <port> — true when something is listening.
port_open() {
	(exec 3<>"/dev/tcp/$1/$2") >/dev/null 2>&1
}

# --- shell profile -----------------------------------------------------------

# shell_profile — the file that sets PATH for the user's login shell.
shell_profile() {
	case "$(basename "${SHELL:-/bin/bash}")" in
	zsh) printf '%s\n' "${ZDOTDIR:-$HOME}/.zshrc" ;;
	bash)
		if [ "$OS" = "macos" ]; then printf '%s\n' "$HOME/.bash_profile"; else printf '%s\n' "$HOME/.bashrc"; fi
		;;
	fish) printf '%s\n' "$HOME/.config/fish/config.fish" ;;
	*) printf '%s\n' "$HOME/.profile" ;;
	esac
}

# profile_add_path <dir> — put <dir> on PATH for future shells, once. Repeat runs
# of the installer are no-ops thanks to the marker comment.
profile_add_path() {
	local dir="$1" profile line marker="# added by mycel installer"
	profile="$(shell_profile)"

	case "$profile" in
	*/fish/*) line="fish_add_path $dir" ;;
	*) line="export PATH=\"\$PATH:$dir\"" ;;
	esac

	if [ -f "$profile" ] && grep -qF "$dir" "$profile"; then
		return 0
	fi

	mkdir -p "$(dirname "$profile")"
	printf '\n%s\n%s\n' "$marker" "$line" >>"$profile"
	log_ok "added $dir to PATH in $profile"
	PROFILE_CHANGED=1
}
