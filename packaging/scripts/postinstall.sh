#!/usr/bin/env bash
set -euo pipefail

install -d /var/log/serverpatcher
install -d /var/lib/serverpatcher/reports
install -d /var/lock

if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
  systemctl daemon-reload || true
  systemctl enable --now serverpatcher.timer >/dev/null 2>&1 || true
fi

exit 0
