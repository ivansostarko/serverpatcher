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