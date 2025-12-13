# Client installation from your hosted APT/YUM repositories

Assumes you publish `public/` via GitHub Pages:
Base URL:
- `https://<your-org>.github.io/serverpatcher/`

APT:
- `https://<your-org>.github.io/serverpatcher/apt`

YUM/DNF:
- `https://<your-org>.github.io/serverpatcher/yum/x86_64`

## Ubuntu / Debian (APT)

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg

sudo install -d -m 0755 /usr/share/keyrings
curl -fsSL https://<your-org>.github.io/serverpatcher/keys/serverpatcher.gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/serverpatcher-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/serverpatcher-archive-keyring.gpg] https://<your-org>.github.io/serverpatcher/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/serverpatcher.list >/dev/null

sudo apt-get update
sudo apt-get install -y serverpatcher

systemctl status serverpatcher.timer
```

## RHEL / Rocky / Alma / Fedora (DNF/YUM)

```bash
sudo tee /etc/yum.repos.d/serverpatcher.repo >/dev/null <<'EOF'
[serverpatcher]
name=Server Patcher
baseurl=https://<your-org>.github.io/serverpatcher/yum/x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF

sudo dnf install -y serverpatcher
systemctl status serverpatcher.timer
```

If you sign repo metadata, set:
- `repo_gpgcheck=1`
- add `gpgkey=...` pointing to your public key.
