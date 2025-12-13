#!/usr/bin/env bash
set -euo pipefail

PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
ETC_DIR="${ETC_DIR:-/etc/serverpatcher}"
SYSTEMD_DIR="${SYSTEMD_DIR:-/etc/systemd/system}"
LOGROTATE_DIR="${LOGROTATE_DIR:-/etc/logrotate.d}"

need_root() {
  if [[ "$(id -u)" != "0" ]]; then
    echo "ERROR: must run as root" >&2
    exit 1
  fi
}

has_systemd() {
  command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]
}

has_openrc() {
  command -v rc-service >/dev/null 2>&1
}

main() {
  need_root

  if has_systemd; then
    systemctl disable --now serverpatcher.timer || true
    systemctl disable --now serverpatcher.service || true
    rm -f "$SYSTEMD_DIR/serverpatcher.timer" "$SYSTEMD_DIR/serverpatcher.service"
    systemctl daemon-reload || true
  fi

  if has_openrc; then
    rc-update del serverpatcher default || true
    rc-service serverpatcher stop || true
    rm -f /etc/init.d/serverpatcher
    rm -f /etc/cron.daily/serverpatcher
  fi

  rm -f "$LOGROTATE_DIR/serverpatcher" || true
  rm -f "$BIN_DIR/serverpatcher" || true

  echo "Removed serverpatcher binary and service definitions."
  echo "Config and data directories were NOT removed:"
  echo "  $ETC_DIR"
  echo "  /var/log/serverpatcher"
  echo "  /var/lib/serverpatcher"
  echo "Remove them manually if you want a full purge."
}
main "$@"
