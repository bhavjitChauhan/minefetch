#!/usr/bin/env bash
set -ex

version=$(curl -s https://api.github.com/repos/bhavjitChauhan/minefetch/releases/latest | grep -Po '"tag_name": "\K.*?(?=")' | sed 's/^v//')
if [ -z "$version" ]; then
    echo -e "\e[91mFailed to get latest version\e[0m"
    exit 1
fi
os="$(uname | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case $arch in
  amd64 | x86_64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo -e "\e[91mUnsupported architecture: $arch\e[0m" >&2; exit 1 ;;
esac

url="https://github.com/bhavjitChauhan/minefetch/releases/download/v${version}/minefetch_${version}_${os}_${arch}"

mkdir -p $HOME/.local/bin
set +x
if ! echo "$PATH" | grep -q "$HOME/.local/bin"; then
    set -x
    if [ -f $HOME/.bashrc ] && ! grep -q 'minefetch' $HOME/.bashrc; then
        echo -e '\nexport PATH="$HOME/.local/bin:$PATH" # minefetch' >> $HOME/.bashrc
    fi
    if [ -f $HOME/.zshrc ] && ! grep -q 'minefetch' $HOME/.zshrc; then
        echo -e '\nexport PATH="$HOME/.local/bin:$PATH" # minefetch' >> $HOME/.zshrc
    fi
    if [ -f $HOME/.config/fish/config.fish ] && ! grep -q 'minefetch' $HOME/.config/fish/config.fish; then
        echo -e '\nset -Ua fish_user_paths ~/.local/bin # minefetch' >> $HOME/.config/fish/config.fish
    fi
    exec $SHELL
fi

set -x
curl -LSfs "$url" -o $HOME/.local/bin/minefetch
chmod +x $HOME/.local/bin/minefetch

set +x
if command -v minefetch >/dev/null 2>&1; then
    echo -e "\n\e[92mSuccessfully installed Minefetch!\e[0m"
    echo -e "You can run it by executing \e[94mminefetch\e[0m in your terminal."
else
    echo -e "\n\e[91mSomething went wrong.\e[0m" >&2
    exit 1
fi
