# Server Patcher

Server Patcher is an open-source Linux patching service written in Go. It applies OS package updates using the host’s native package manager (APT, DNF, YUM, Zypper, Pacman, APK), writes structured logs, persists a JSON report per run, and can email a run report.

This repository is a production-oriented baseline, not a promise of risk-free “patch everything on every distro.” Patching is inherently high risk; the tool focuses on determinism, auditability, and operational guardrails.

## Supported backends

The backend is selected by detecting the package manager binary:
- Debian/Ubuntu: `apt-get` (optional security-only via `unattended-upgrade` if installed)
- RHEL/Fedora/Alma/Rocky: `dnf` or `yum`
- SUSE/openSUSE: `zypper`
- Arch: `pacman`
- Alpine: `apk`

Not yet implemented (by design, requires dedicated logic):
- rpm-ostree (Silverblue/CoreOS)
- transactional-update / MicroOS
- image-based/immutable update systems

## How it works

Each run:
1. Acquires a lock file to prevent concurrent patch runs
2. Detects OS from `/etc/os-release` and selects a backend
3. Executes update/upgrade commands (non-interactive where possible)
4. Detects “reboot required” (best-effort)
5. Writes a JSON report file
6. Optionally emails the report (JSON attached)

## Installation (systemd timer recommended)

### Build
```bash
git clone https://github.com/ivansostarko/serverpatcher.git
cd serverpatcher
make build
```

### Install service + timer
The installer:
- installs the binary to `/usr/local/bin/serverpatcher`
- installs config (if missing) to `/etc/serverpatcher/config.json`
- installs systemd unit + timer
- installs a logrotate snippet
- enables and starts the timer

```bash
sudo ./scripts/install.sh
```

### Verify
```bash
serverpatcher detect
serverpatcher validate-config --config /etc/serverpatcher/config.json
systemctl status serverpatcher.timer
journalctl -u serverpatcher.service -n 200 --no-pager
```

## Commands

```bash
serverpatcher run-once --config /etc/serverpatcher/config.json
serverpatcher daemon --config /etc/serverpatcher/config.json
serverpatcher detect
serverpatcher print-default-config --pretty
serverpatcher validate-config --config /etc/serverpatcher/config.json
serverpatcher version
```

## Configuration

Config is JSON. Start from the example in `configs/config.example.json` or generate defaults:

```bash
serverpatcher print-default-config --pretty > config.json
```

### Important settings
- `patching.dry_run`: best-effort simulation (varies by package manager)
- `patching.security_only`: best-effort security-only updates (varies by package manager)
- `patching.reboot_policy`:
  - `none`: never reboot, just report
  - `notify`: report reboot-required in JSON/email
  - `reboot`: attempt to reboot the machine when a reboot is required (dangerous)
- `patching.pre_hook` / `patching.post_hook`: executable paths, run before/after patching
- `email.password_env`: environment variable name holding the SMTP password (recommended)
- `logging.file`: log path (rotation is handled via logrotate sample installed by the script)

### SMTP credentials
Set your SMTP password via environment variable for the service. Example approach:
- put a root-owned drop-in in `/etc/systemd/system/serverpatcher.service.d/override.conf`
- set `Environment=SERVERPATCHER_EMAIL_PASSWORD=...`
- run `systemctl daemon-reload`

## Reports and logs

### Reports
Default: `/var/lib/serverpatcher/reports/report_<hostname>_<timestamp>.json`

The JSON includes:
- step timing
- command stdout/stderr
- exit codes
- reboot-required signal (best-effort)

### Logs
Default: `/var/log/serverpatcher/serverpatcher.log`

This repository includes a logrotate snippet (`systemd/logrotate.serverpatcher`). The installer places it at `/etc/logrotate.d/serverpatcher`.

## Operational guardrails (read this)

1. **Patching requires root.** Treat this as privileged code.
2. **Use systemd timers for fleets.** Prefer `run-once` via a timer, not a daemon.
3. **Reboots are a business decision.** Do not enable `reboot_policy=reboot` unless you explicitly accept surprise downtime.
4. **Security-only updates are not uniform.** “Security only” support varies across distros and repositories; the tool uses best-effort behavior and may fall back to full upgrades.
5. **Test per distro.** Repo configs, proxies, and interactive prompts differ.

## License
MIT. See `LICENSE`.
