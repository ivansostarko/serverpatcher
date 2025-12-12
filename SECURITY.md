# Security Policy

## Supported Versions
This project is currently pre-1.0 and may change rapidly.

## Reporting a Vulnerability
Please open a private security advisory in the GitHub repository if available, or email the maintainers listed in the repository.
Include:
- Affected version/commit
- Linux distribution and package manager
- Steps to reproduce
- Expected vs actual behavior
- Logs with sensitive data removed

## Operational Guidance
- This service typically runs as root because package installation requires it.
- Prefer systemd timer + `run-once` over a long-lived daemon for minimal attack surface.
- Store SMTP credentials in environment variables (root-only) instead of plaintext config files.
- Restrict network egress if your environment requires it.
