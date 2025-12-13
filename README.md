# Server Patcher

Server Patcher is an open-source Linux patch automation tool written in Go. It applies OS package updates using the host’s native package manager, writes structured logs, persists a JSON report for each run, and can optionally email a report.

This project is an execution-and-reporting core. It does **not** magically solve fleet orchestration, rollback, or application-aware maintenance windows. If you pretend it does, you will ship outages.


## Supported Linux families (package-manager backends)

Backend selection is done by detecting the package-manager binary on the host.

- Debian / Ubuntu: `apt-get` (best-effort security-only via `unattended-upgrade` if installed)
- RHEL / CentOS / Rocky / Alma / Fedora: `dnf` or `yum`
- SUSE / openSUSE: `zypper`
- Arch Linux: `pacman`
- Alpine Linux: `apk`


## How it works

Each run:
1. Acquires a lock file to prevent concurrent runs
2. Detects OS from `/etc/os-release` and selects a backend
3. Executes update/upgrade commands (non-interactive where possible)
4. Detects “reboot required” (best-effort, backend-dependent)
5. Writes a JSON report file
6. Optionally emails a report (JSON attached)


## Build from source

### Prerequisites
- Go 1.22+
- git
- make (optional)

Install prerequisites by distro:

**Ubuntu/Debian**
```bash
sudo apt-get update
sudo apt-get install -y git golang make
```

**Fedora/RHEL/Rocky/Alma**
```bash
sudo dnf install -y git golang make
```

**SUSE/openSUSE**
```bash
sudo zypper install -y git go make
```

**Arch**
```bash
sudo pacman -Syu --noconfirm git go make
```

**Alpine**
```bash
sudo apk add --no-cache git go make
```

### Build
```bash
git clone https://github.com/ivansostarko/serverpatcher.git
cd serverpatcher

# Option A: make
make build

# Option B: direct go build
go build -trimpath -ldflags "-s -w" -o bin/serverpatcher ./cmd/serverpatcher
```

Sanity checks:
```bash
./bin/serverpatcher version
./bin/serverpatcher detect
```

### Distros without systemd (e.g., Alpine with OpenRC)
You have two realistic options:
1. **cron + run-once** (recommended)
2. **OpenRC service running daemon mode** (works, but more moving parts)

The installer detects OpenRC and installs:
- `/etc/cron.daily/serverpatcher` (runs `serverpatcher run-once`)
- optional `/etc/init.d/serverpatcher` (daemon mode) if `INSTALL_DAEMON_OPENRC=1`

Enable cron on Alpine:
```bash
rc-update add crond default
rc-service crond start
```

## Install as a service

### Recommendation: systemd timer + oneshot service
This is the safest operational model: a short-lived process that runs on schedule.

#### Install (systemd)
From the repo root:
```bash
sudo ./scripts/install.sh
```

Verify:
```bash
serverpatcher validate-config --config /etc/serverpatcher/config.json
systemctl status serverpatcher.timer
journalctl -u serverpatcher.service -n 200 --no-pager
```

Force an immediate run:
```bash
sudo systemctl start serverpatcher.service
```

## Configuration

Config file format is JSON.

Default path (installer):
- `/etc/serverpatcher/config.json`

Generate a default config:
```bash
serverpatcher print-default-config --pretty > config.json
```

### Key settings
- `patching.dry_run`: best-effort simulation (varies by backend)
- `patching.security_only`: best-effort “security only” updates (varies by backend and repo configuration)
- `patching.reboot_policy`:
  - `none`: never reboot, just report
  - `notify`: include reboot-required in report/email
  - `reboot`: attempt to reboot the host when a reboot is required (**dangerous**)
- `patching.pre_hook` / `patching.post_hook`: executable paths
- `email.password_env`: environment variable name holding the SMTP password (recommended)
- `logging.file`: `/var/log/serverpatcher/serverpatcher.log` (rotated via logrotate)
- `report.dir`: `/var/lib/serverpatcher/reports`


### SMTP credentials 
Do **not** hardcode SMTP passwords in the JSON config. Use an environment variable.

Example for systemd (drop-in override):
```bash
sudo mkdir -p /etc/systemd/system/serverpatcher.service.d
sudo tee /etc/systemd/system/serverpatcher.service.d/override.conf >/dev/null <<'EOF'
[Service]
Environment=SERVERPATCHER_EMAIL_PASSWORD=REPLACE_ME
EOF

sudo systemctl daemon-reload
sudo systemctl restart serverpatcher.timer
```

## Reports and logs

### Reports
Default directory:
- `/var/lib/serverpatcher/reports/`

Files:
- `report_<hostname>_<timestamp>.json`

### Logs
Default log file:
- `/var/log/serverpatcher/serverpatcher.log`

Rotation:
- `/etc/logrotate.d/serverpatcher`

## CLI commands

```bash
serverpatcher run-once --config /etc/serverpatcher/config.json [--verbose]
serverpatcher daemon --config /etc/serverpatcher/config.json [--verbose]
serverpatcher detect
serverpatcher validate-config --config /etc/serverpatcher/config.json
serverpatcher print-default-config --pretty[=true|false]
serverpatcher version
```


## Uninstall

```bash
sudo ./scripts/uninstall.sh
```

## License
MIT. See `LICENSE`.
