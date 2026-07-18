#!/bin/bash
set -euo pipefail

BINARY_NAME="pbasic"

if [ "$(id -u)" -eq 0 ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
fi

rm -f "$INSTALL_DIR/$BINARY_NAME"
echo "Removed: $INSTALL_DIR/$BINARY_NAME"
