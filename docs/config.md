# Config file


/etc/serverpatcher/config.json
```bash
{
  "server": {
    "interval": "24h",
    "jitter": "30m",
    "timeout": "2h",
    "lock_file": "/var/lock/serverpatcher.lock"
  },
  "patching": {
    "dry_run": false,
    "security_only": false,
    "exclude_packages": [],
    "pre_hook": "",
    "post_hook": "",
    "reboot_policy": "notify",
    "allow_kernel_updates": true,
    "package_timeout": "90m",
    "command_nice": 10,
    "command_ionice": "best-effort:7"
  },
  "email": {
    "enabled": false,
    "from": "serverpatcher@your-domain",
    "to": [
      "ops@your-domain"
    ],
    "smtp_host": "smtp.your-domain",
    "smtp_port": 587,
    "username": "",
    "password_env": "SERVERPATCHER_EMAIL_PASSWORD",
    "starttls": true,
    "subject_prefix": "[Server Patcher]"
  },
  "logging": {
    "level": "info",
    "file": "/var/log/serverpatcher/serverpatcher.log",
    "json": true,
    "also_stdout": false
  },
  "report": {
    "dir": "/var/lib/serverpatcher/reports",
    "retain_days": 30
  },
  "health": {
    "enabled": false,
    "listen": "127.0.0.1:9109"
  }
}

```