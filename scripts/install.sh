#!/usr/bin/env bash
set -euo pipefail

PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
ETC_DIR="${ETC_DIR:-/etc/serverpatcher}"
CONFIG_NAME="${CONFIG_NAME:-config.json}"
SYSTEMD_DIR="${SYSTEMD_DIR:-/etc/systemd/system}"
LOGROTATE_DIR="${LOGROTATE_DIR:-/etc/logrotate.d}"
INIT_SYSTEM="${INIT_SYSTEM:-auto}"          # auto|systemd|openrc|none
INSTALL_DAEMON_OPENRC="${INSTALL_DAEMON_OPENRC:-0}"  # 0|1

need_root() {
  if [[ "$(id -u)" != "0" ]]; then
    echo "ERROR: must run as root" >&2
    exit 1
  fi
}

detect_init() {
  if [[ "$INIT_SYSTEM" != "auto" ]]; then
    echo "$INIT_SYSTEM"
    return
  fi
  if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
    echo "systemd"
    return
  fi
  if command -v rc-service >/dev/null 2>&1; then
    echo "openrc"
    return
  fi
  echo "none"
}

build_binary() {
  if [[ -x ./bin/serverpatcher ]]; then
    echo "./bin/serverpatcher"
    return
  fi
  if command -v go >/dev/null 2>&1; then
    echo "Building serverpatcher..."
    mkdir -p ./bin
    go build -trimpath -ldflags "-s -w" -o ./bin/serverpatcher ./cmd/serverpatcher
    echo "./bin/serverpatcher"
    return
  fi
  echo "ERROR: no binary found at ./bin/serverpatcher and Go not installed." >&2
  exit 1
}

install_binary() {
  local src="$1"
  echo "Installing binary to $BIN_DIR/serverpatcher"
  install -d "$BIN_DIR"
  install -m 0755 "$src" "$BIN_DIR/serverpatcher"
}

install_runtime_dirs() {
  install -d /var/log/serverpatcher
  install -d /var/lib/serverpatcher/reports
  install -d /var/lock
}

install_config() {
  echo "Installing config to $ETC_DIR/$CONFIG_NAME (if missing)"
  install -d "$ETC_DIR"
  if [[ ! -f "$ETC_DIR/$CONFIG_NAME" ]]; then
    if [[ -f ./configs/config.example.json ]]; then
      install -m 0644 ./configs/config.example.json "$ETC_DIR/$CONFIG_NAME"
    else
      "$BIN_DIR/serverpatcher" print-default-config --pretty > "$ETC_DIR/$CONFIG_NAME"
      chmod 0644 "$ETC_DIR/$CONFIG_NAME"
    fi
  fi
}

install_logrotate() {
  if [[ -f ./systemd/logrotate.serverpatcher && -d "$LOGROTATE_DIR" ]]; then
    echo "Installing logrotate config to $LOGROTATE_DIR/serverpatcher"
    install -m 0644 ./systemd/logrotate.serverpatcher "$LOGROTATE_DIR/serverpatcher"
  fi
}

install_systemd() {
  echo "Installing systemd unit files to $SYSTEMD_DIR"
  install -d "$SYSTEMD_DIR"
  install -m 0644 ./systemd/serverpatcher.service "$SYSTEMD_DIR/serverpatcher.service"
  install -m 0644 ./systemd/serverpatcher.timer "$SYSTEMD_DIR/serverpatcher.timer"

  systemctl daemon-reload
  systemctl enable --now serverpatcher.timer
}

install_openrc() {
  echo "Detected OpenRC: installing cron.daily runner"
  install -d /etc/cron.daily
  install -m 0755 ./openrc/cron.daily.serverpatcher /etc/cron.daily/serverpatcher

  if [[ "$INSTALL_DAEMON_OPENRC" == "1" ]]; then
    echo "Installing OpenRC init.d service (daemon mode)"
    install -d /etc/init.d
    install -m 0755 ./openrc/serverpatcher.initd /etc/init.d/serverpatcher
    rc-update add serverpatcher default || true
  fi

  echo "NOTE: ensure crond is enabled/running:"
  echo "  rc-update add crond default"
  echo "  rc-service crond start"
}

main() {
  need_root

  if [[ ! -f ./cmd/serverpatcher/main.go ]]; then
    echo "ERROR: run this script from the repository root." >&2
    exit 1
  fi

  local init
  init="$(detect_init)"
  echo "Init system: $init"

  local bin
  bin="$(build_binary)"

  install_runtime_dirs
  install_binary "$bin"
  install_config
  install_logrotate

  case "$init" in
    systemd) install_systemd ;;
    openrc) install_openrc ;;
    none)
      echo "WARNING: no systemd/openrc detected. Installed binary and config only."
      echo "Schedule runs yourself, e.g.:"
      echo "  $BIN_DIR/serverpatcher run-once --config $ETC_DIR/$CONFIG_NAME"
      ;;
    *)
      echo "ERROR: unknown init system: $init" >&2
      exit 1
      ;;
  esac

  echo "Done."
  echo "Binary: $BIN_DIR/serverpatcher"
  echo "Config:  $ETC_DIR/$CONFIG_NAME"
}
main "$@"
