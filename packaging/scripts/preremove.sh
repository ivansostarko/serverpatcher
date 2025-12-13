#!/usr/bin/env bash
set -euo pipefail

if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
  systemctl disable --now serverpatcher.timer >/dev/null 2>&1 || true
  systemctl stop serverpatcher.service >/dev/null 2>&1 || true
fi

exit 0
