#!/bin/sh
# Run from an extracted release directory. Installs no runner or models.
set -eu
source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
install_dir=${INSTALL_DIR:-"$HOME/.local/bin"}
for command_name in myLlama myLlamactl; do
  if [ ! -f "$source_dir/$command_name" ]; then
    echo "Missing $command_name beside install.sh" >&2
    exit 1
  fi
done
mkdir -p "$install_dir"
for command_name in myLlama myLlamactl; do
  install -m 755 "$source_dir/$command_name" "$install_dir/$command_name"
done
printf 'Installed myLlama and myLlamactl in %s\nAdd that directory to PATH, then run myLlamactl init.\n' "$install_dir"
