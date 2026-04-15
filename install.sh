#!/bin/sh
# Minefetch installation script for macOS and Linux.
#
# This script supports the fish, Zsh, and Bash shells. It detects the shell
# using the SHELL environment variable, and only modifies the configuration file
# for that shell. You can override the detected shell by setting the
# MINEFETCH_SHELL environment variable to "fish", "zsh", or "bash".
#
# The default installation directory is ~/.local/bin. You can override this by
# setting the MINEFETCH_INSTALL environment variable.

set -e

success() {
	printf '\n\033[1m\033[92m%s\033[0m\n' "$1"
}

error() {
	printf '\n\033[91m%s\033[0m\n' "$1" >&2
}

highlight() {
	printf '\033[7m %s \033[0m' "$1"
}

case "$(uname -ms)" in
'Darwin arm64') port='darwin_arm64' ;;
'Linux x86_64') port='linux_amd64' ;;
'Linux aarch64') port='linux_arm64' ;;
*)
	error "Unsupported platform: $(uname -ms)"
	echo 'This script only supports macOS on M-series chips and Linux on x64 or ARM.'
	exit 1
	;;
esac

url="https://github.com/bhavjitChauhan/minefetch/releases/latest/download/minefetch_${port}"
install="${MINEFETCH_INSTALL:-$HOME/.local/bin}"

set -x

mkdir -p "$install"
curl --fail --location --progress-bar --output "$install/minefetch" "$url"
chmod +x "$install/minefetch"

{ set +x; } 2>/dev/null

if command -v minefetch >/dev/null; then
	success 'Successfully installed Minefetch!'
	printf 'You can run it using %s in your terminal.\n' "$(highlight 'minefetch')"
	exit
fi

install_tilde=$(echo "$install" | sed "s|^$HOME|~|")
install_home=$(echo "$install" | sed "s|^$HOME|\$HOME|")
shell=${MINEFETCH_SHELL:-${SHELL##*/}}

case $shell in
fish)
	config='.config/fish/conf.d/minefetch.fish'
	command="fish_add_path $install_tilde"
	if grep -Fqx "$command" "$HOME/$config" 2>/dev/null; then
		error 'Installation already configured'
		printf 'Restart your terminal or run %s to apply the changes, or specify your shell using MINEFETCH_SHELL.\n' "$(highlight "source ~/$config")"
		exit
	fi
	set -x
	mkdir -p "${HOME}/${config%/*}"
	echo "$command" | tee "$HOME/$config" >/dev/null
	;;
zsh | bash)
	case $shell in
	zsh) config='.zshrc' ;;
	bash) config='.bashrc' ;;
	esac
	command=$(printf 'export PATH="%s:$PATH"\n' "$install_home")
	if grep -Fqx "$command" "$HOME/$config" 2>/dev/null; then
		error 'Installation already configured'
		printf 'Restart your terminal or run %s to apply the changes, or specify your shell using MINEFETCH_SHELL.\n' "$(highlight "source ~/$config")"
		exit
	fi
	set -x
	if [ ! -f "$HOME/$config" ]; then
		echo "$command" | tee "$HOME/$config" >/dev/null
	else
		echo "\n$command" | tee -a "$HOME/$config" >/dev/null
	fi
	;;
*)
	error 'Unsupported shell'
	echo "Add $(highlight "$install_tilde") to your PATH manually."
	exit 1
	;;
esac

{ set +x; } 2>/dev/null

success 'Successfully installed Minefetch!'
printf 'Restart your terminal or run %s to apply the changes.\n' "$(highlight "source ~/$config")"
