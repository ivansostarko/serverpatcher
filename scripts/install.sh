#!/usr/bin/env bash
set -euo pipefail

PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
ETC_DIR="${ETC_DIR:-/etc/serverpatcher}"
SYSTEMD_DIR="${SYSTEMD_DIR:-/etc/systemd/system}"
LOGROTATE_DIR="${LOGROTATE_DIR:-/etc/logrotate.d}"

echo "Building..."
go build -trimpath -ldflags "-s -w" -o serverpatcher ./cmd/serverpatcher

echo "Installing binary to $BIN_DIR"
install -d "$BIN_DIR"
install -m 0755 ./serverpatcher "$BIN_DIR/serverpatcher"

echo "Installing config example to $ETC_DIR (if missing)"
install -d "$ETC_DIR"
if [[ ! -f "$ETC_DIR/config.json" ]]; then
  install -m 0644 ./configs/config.example.json "$ETC_DIR/config.json"
fi

echo "Installing systemd unit files to $SYSTEMD_DIR"
install -m 0644 ./systemd/serverpatcher.service "$SYSTEMD_DIR/serverpatcher.service"
install -m 0644 ./systemd/serverpatcher.timer "$SYSTEMD_DIR/serverpatcher.timer"

echo "Installing logrotate config to $LOGROTATE_DIR"
install -m 0644 ./systemd/logrotate.serverpatcher "$LOGROTATE_DIR/serverpatcher"

echo "Reloading systemd and enabling timer..."
systemctl daemon-reload
systemctl enable --now serverpatcher.timer

echo "Done. Check status with: systemctl status serverpatcher.timer"
