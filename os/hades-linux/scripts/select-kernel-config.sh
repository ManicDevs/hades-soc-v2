#!/usr/bin/env bash
set -euo pipefail

# select-kernel-config.sh <mode>
# Modes:
#   prod   -> use kernel-config-prod
#   minimal -> use kernel-config-minimal
#   show   -> print current kernel-config path

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/../configs"
TARGET="$CONFIG_DIR/kernel-config"

usage() {
  echo "Usage: $0 {prod|minimal|show}"
  exit 2
}

mode="${1:-}" || true
case "$mode" in
  prod)
    src="$CONFIG_DIR/kernel-config-prod"
    ;;
  minimal)
    src="$CONFIG_DIR/kernel-config-minimal"
    ;;
  show)
    echo "Current kernel-config: $TARGET"
    ls -l "$TARGET" || true
    exit 0
    ;;
  *)
    usage
    ;;
esac

if [ ! -f "$src" ]; then
  echo "Source config not found: $src"
  exit 1
fi

# Backup current
if [ -f "$TARGET" ]; then
  ts=$(date -u +%Y%m%dT%H%M%SZ)
  cp "$TARGET" "$TARGET.bak.$ts"
  echo "Backed up $TARGET -> $TARGET.bak.$ts"
fi

cp "$src" "$TARGET"
chmod 644 "$TARGET"

echo "Selected kernel config: $src -> $TARGET"
